package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/opentdf/platform/service/logger"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type Table struct {
	name       string
	schema     string
	withSchema bool
}

func NewTable(schema string) func(name string) Table {
	s := schema
	return func(name string) Table {
		return Table{
			name:       name,
			schema:     s,
			withSchema: true,
		}
	}
}

func (t Table) WithoutSchema() Table {
	nT := NewTable(t.schema)(t.name)
	nT.withSchema = false
	return nT
}

func (t Table) Name() string {
	if t.withSchema {
		return t.schema + "." + t.name
	}
	return t.name
}

func (t Table) Field(field string) string {
	return t.Name() + "." + field
}

// We can rename this but wanted to get mocks working.
type PgxIface interface {
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	Begin(ctx context.Context) (pgx.Tx, error)
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...any) pgx.Row
	Query(context.Context, string, ...any) (pgx.Rows, error)
	Ping(context.Context) error
	Close()
	Config() *pgxpool.Config
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

/*
A wrapper around a pgxpool.Pool and sql.DB reference for Postgres.
*/
type PostgresClient struct {
	Pgx           PgxIface
	Logger        *logger.Logger
	config        Config
	pgConfig      PostgresConfig
	ranMigrations bool
	SQLDB         *sql.DB
	tracer        trace.Tracer
}

func newPostgresClient(ctx context.Context, config Config, logCfg logger.Config, tracer *trace.Tracer) (*PostgresClient, error) {
	c := PostgresClient{
		config:   config,
		pgConfig: config.Postgres,
	}

	if tracer != nil {
		c.tracer = *tracer
	} else {
		c.tracer = noop.NewTracerProvider().Tracer("db")
	}

	l, err := logger.NewLogger(logger.Config{
		Output: logCfg.Output,
		Type:   logCfg.Type,
		Level:  logCfg.Level,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	c.Logger = l.With("schema", c.pgConfig.Schema)

	dbConfig, err := c.pgConfig.buildConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}

	dbConfig.ConnConfig.OnNotice = func(_ *pgconn.PgConn, n *pgconn.Notice) {
		switch n.Severity {
		case "DEBUG":
			c.Logger.Debug("database notice", slog.String("message", n.Message))
		case "NOTICE":
			c.Logger.Info("database notice", slog.String("message", n.Message))
		case "WARNING":
			c.Logger.Warn("database notice", slog.String("message", n.Message))
		case "ERROR":
			c.Logger.Error("database notice", slog.String("message", n.Message))
		}
	}

	slog.Info("opening new database pool", slog.String("schema", c.pgConfig.Schema))
	pool, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgxpool: %w", err)
	}
	c.Pgx = pool
	c.SQLDB = stdlib.OpenDBFromPool(pool)

	if c.config.VerifyConnection {
		if err := c.Pgx.Ping(ctx); err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
	}

	return &c, nil
}

func (c *PostgresClient) Driver() string { return DriverPostgres }

func (c *PostgresClient) PgxPool() PgxIface { return c.Pgx }

func (c *PostgresClient) DB() *sql.DB { return c.SQLDB }

func (c *PostgresClient) Tracer() trace.Tracer { return c.tracer }

func (c *PostgresClient) Close() {
	c.Pgx.Close()
	c.SQLDB.Close()
}

func (c PostgresConfig) buildConfig() (*pgxpool.Config, error) {
	u := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		c.User,
		url.QueryEscape(c.Password),
		net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		c.Database,
		c.SSLMode,
	)
	parsed, err := pgxpool.ParseConfig(u)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}

	// Apply connection and pool configurations
	if c.Pool.MinConns > 0 {
		parsed.MinConns = c.Pool.MinConns
	}
	if c.Pool.MinIdleConns > 0 {
		parsed.MinIdleConns = c.Pool.MinIdleConns
	}
	// Non-zero defaults
	parsed.ConnConfig.ConnectTimeout = time.Duration(c.ConnectTimeout) * time.Second
	parsed.MaxConns = c.Pool.MaxConns
	parsed.MaxConnLifetime = time.Duration(c.Pool.MaxConnLifetime) * time.Second
	parsed.MaxConnIdleTime = time.Duration(c.Pool.MaxConnIdleTime) * time.Second
	parsed.HealthCheckPeriod = time.Duration(c.Pool.HealthCheckPeriod) * time.Second

	// Configure the search_path schema immediately on connection opening
	parsed.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET search_path TO "+pgx.Identifier{c.Schema}.Sanitize())
		if err != nil {
			slog.Error("failed to set database client search_path",
				slog.String("schema", c.Schema),
				slog.Any("error", err),
			)
			return err
		}
		slog.Debug("successfully set database client search_path", slog.String("schema", c.Schema))
		return nil
	}
	return parsed, nil
}

// Common function for all queryRow calls
func (c PostgresClient) QueryRow(ctx context.Context, sql string, args []interface{}) (pgx.Row, error) {
	c.Logger.TraceContext(ctx, "sql", slog.String("sql", sql), slog.Any("args", args))
	return c.Pgx.QueryRow(ctx, sql, args...), nil
}

// Common function for all query calls
func (c PostgresClient) Query(ctx context.Context, sql string, args []interface{}) (pgx.Rows, error) {
	c.Logger.TraceContext(ctx, "sql", slog.String("sql", sql), slog.Any("args", args))
	r, e := c.Pgx.Query(ctx, sql, args...)
	if e != nil {
		return nil, WrapIfKnownInvalidQueryErr(e)
	}
	if r.Err() != nil {
		return nil, WrapIfKnownInvalidQueryErr(r.Err())
	}
	return r, nil
}

// Common function for all exec calls
func (c PostgresClient) Exec(ctx context.Context, sql string, args []interface{}) error {
	c.Logger.TraceContext(ctx, "sql", slog.String("sql", sql), slog.Any("args", args))
	tag, err := c.Pgx.Exec(ctx, sql, args...)
	if err != nil {
		return WrapIfKnownInvalidQueryErr(err)
	}

	if tag.RowsAffected() == 0 {
		return WrapIfKnownInvalidQueryErr(pgx.ErrNoRows)
	}

	return nil
}

//
// Helper functions for building SQL
//

// Postgres uses $1, $2, etc. for placeholders
func NewStatementBuilder() sq.StatementBuilderType {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
}
