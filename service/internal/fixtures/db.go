package fixtures

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/opentdf/platform/service/logger"
	"github.com/opentdf/platform/service/pkg/config"
	"github.com/opentdf/platform/service/pkg/db"
	policydb "github.com/opentdf/platform/service/policy/db"
	"github.com/opentdf/platform/service/tracing"
	"go.opentelemetry.io/otel"
)

var (
	// Configured default LIST Limit when working with fixtures
	fixtureLimitDefault int32 = 1000
	fixtureLimitMax     int32 = 5000
)

type DBInterface struct {
	Client       db.Client
	PgxClient    db.PgxClient
	PolicyClient policydb.PolicyDBClient
	Schema       string
	LimitDefault int32
	LimitMax     int32
	sqlitePath   string
}

//nolint:nestif // Driver-specific sqlite setup keeps fixture DB handling explicit in one place.
func NewDBInterface(ctx context.Context, cfg config.Config) DBInterface {
	config := cfg.DB
	config.Normalize()
	schema := ""
	if config.Driver == db.DriverPostgres {
		schema = config.Postgres.Schema
		if schema == "" {
			schema = config.Schema
		}
	}
	sqlitePath := ""
	if config.Driver == db.DriverSQLite {
		if config.SQLite.InMemory {
			if config.SQLite.Path == "" || config.SQLite.Path == "opentdf.db" {
				prefix := "opentdf-integration-"
				if schema != "" {
					prefix = strings.ReplaceAll(schema, ".", "_") + "-"
				}
				config.SQLite.Path = prefix + uuid.NewString()
			}
		} else if config.SQLite.Path == "" || config.SQLite.Path == "opentdf.db" {
			prefix := "opentdf-integration-"
			if schema != "" {
				prefix = strings.ReplaceAll(schema, ".", "_") + "-"
			}
			tmp, err := os.CreateTemp("", prefix+"*.db")
			if err != nil {
				slog.Error("issue creating sqlite temp database", slog.Any("error", err))
				panic(err)
			}
			sqlitePath = tmp.Name()
			if err := tmp.Close(); err != nil {
				slog.Error("issue closing sqlite temp database", slog.Any("error", err))
				panic(err)
			}
			config.SQLite.Path = sqlitePath
		}
	}
	logCfg := cfg.Logger
	tracer := otel.Tracer(tracing.ServiceName)

	client, err := db.New(ctx, config, logCfg, &tracer)
	if err != nil {
		slog.Error("issue creating database client", slog.Any("error", err))
		panic(err)
	}
	logger, err := logger.NewLogger(logger.Config{
		Level:  cfg.Logger.Level,
		Output: cfg.Logger.Output,
		Type:   cfg.Logger.Type,
	})
	if err != nil {
		slog.Error("issue creating logger", slog.Any("error", err))
		panic(err)
	}

	pgxClient, _ := client.(db.PgxClient)

	return DBInterface{
		Client:       client,
		PgxClient:    pgxClient,
		Schema:       schema,
		PolicyClient: mustPolicyClient(client, logger, fixtureLimitMax, fixtureLimitDefault),
		LimitDefault: fixtureLimitDefault,
		LimitMax:     fixtureLimitMax,
		sqlitePath:   sqlitePath,
	}
}

func mustPolicyClient(c db.Client, logger *logger.Logger, limitMax, def int32) policydb.PolicyDBClient {
	client, err := policydb.NewClient(c, logger, limitMax, def)
	if err != nil {
		panic(err)
	}
	return client
}

// TableName returns a sanitized fully-qualified table name.
func (d *DBInterface) TableName(v string) string {
	if d.Client.Driver() == db.DriverPostgres && d.Schema != "" {
		return pgx.Identifier{d.Schema, v}.Sanitize()
	}
	return d.sanitizeIdentifier(v)
}

// ExecInsert inserts multiple rows into a table using parameterized queries.
// Each row's values are passed as any types, allowing pgx to handle type conversion.
func (d *DBInterface) ExecInsert(ctx context.Context, table string, columns []string, values ...[]any) (int64, error) {
	if len(values) == 0 {
		return 0, nil
	}

	// Build the INSERT statement with placeholders
	numColumns := len(columns)
	var placeholders []string
	var allArgs []any

	placeholderNum := 1
	for _, row := range values {
		if len(row) != numColumns {
			slog.Error("column count mismatch",
				slog.Int("expected", numColumns),
				slog.Int("got", len(row)),
			)
			return 0, fmt.Errorf("column count mismatch: expected %d, got %d", numColumns, len(row))
		}

		var rowPlaceholders []string
		for _, arg := range row {
			normalized, err := d.normalizeArg(arg)
			if err != nil {
				return 0, err
			}
			rowPlaceholders = append(rowPlaceholders, d.placeholder(placeholderNum))
			placeholderNum++
			allArgs = append(allArgs, normalized)
		}
		placeholders = append(placeholders, "("+strings.Join(rowPlaceholders, ",")+")")
	}

	// Safely sanitize table name using pgx.Identifier
	tableName := d.TableName(table)

	// Safely sanitize column names using pgx.Identifier
	sanitizedColumns := make([]string, len(columns))
	for i, col := range columns {
		sanitizedColumns[i] = d.sanitizeIdentifier(col)
	}

	sql := "INSERT INTO " + tableName +
		" (" + strings.Join(sanitizedColumns, ",") + ")" +
		" VALUES " + strings.Join(placeholders, ",")

	return d.exec(ctx, sql, allArgs...)
}

func (d *DBInterface) DropSchema(ctx context.Context) error {
	if d.Client.Driver() == db.DriverSQLite {
		return d.dropSQLiteTables(ctx)
	}
	if d.Schema == "" {
		return nil
	}
	stmt := "DROP SCHEMA IF EXISTS " + pgx.Identifier{d.Schema}.Sanitize() + " CASCADE"
	_, err := d.exec(ctx, stmt)
	return err
}

func (d DBInterface) Close() {
	if d.Client != nil {
		d.Client.Close()
	}
	if d.sqlitePath != "" {
		_ = os.Remove(d.sqlitePath)
	}
}

func (d *DBInterface) Exec(ctx context.Context, stmt string, args ...any) (int64, error) {
	return d.exec(ctx, stmt, args...)
}

func (d *DBInterface) Query(ctx context.Context, stmt string, args ...any) (*sql.Rows, error) {
	stmt = d.rewritePlaceholders(stmt)
	rows, err := d.Client.DB().QueryContext(ctx, stmt, args...)
	if err != nil {
		slog.Error("query error",
			slog.String("stmt", stmt),
			slog.Any("err", err),
		)
		return nil, err
	}
	return rows, nil
}

func (d *DBInterface) normalizeArg(arg any) (any, error) {
	if d.Client.Driver() != db.DriverSQLite {
		return arg, nil
	}
	switch v := arg.(type) {
	case []string:
		encoded, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("encode sqlite slice: %w", err)
		}
		return string(encoded), nil
	default:
		return arg, nil
	}
}

func (d *DBInterface) exec(ctx context.Context, stmt string, args ...any) (int64, error) {
	stmt = d.rewritePlaceholders(stmt)
	if d.PgxClient != nil {
		res, err := d.PgxClient.PgxPool().Exec(ctx, stmt, args...)
		if err != nil {
			slog.Error("exec error",
				slog.String("stmt", stmt),
				slog.Any("err", err),
			)
			return 0, err
		}
		return res.RowsAffected(), nil
	}
	result, err := d.Client.DB().ExecContext(ctx, stmt, args...)
	if err != nil {
		slog.Error("exec error",
			slog.String("stmt", stmt),
			slog.Any("err", err),
		)
		return 0, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		slog.Error("exec rows affected error",
			slog.String("stmt", stmt),
			slog.Any("err", err),
		)
		return 0, err
	}
	return rows, nil
}

func (d *DBInterface) placeholder(i int) string {
	if d.Client.Driver() == db.DriverSQLite {
		return "?"
	}
	return fmt.Sprintf("$%d", i)
}

func (d *DBInterface) sanitizeIdentifier(name string) string {
	if d.Client.Driver() == db.DriverPostgres {
		return pgx.Identifier{name}.Sanitize()
	}
	return quoteSQLiteIdent(name)
}

func quoteSQLiteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

var sqlitePlaceholders = regexp.MustCompile(`\$\d+`)

func (d *DBInterface) rewritePlaceholders(stmt string) string {
	if d.Client.Driver() != db.DriverSQLite {
		return stmt
	}
	return sqlitePlaceholders.ReplaceAllString(stmt, "?")
}

func (d *DBInterface) dropSQLiteTables(ctx context.Context) error {
	dbConn := d.Client.DB()
	if dbConn == nil {
		return nil
	}
	if _, err := dbConn.ExecContext(ctx, "PRAGMA foreign_keys = OFF"); err != nil {
		slog.Error("sqlite pragma foreign_keys off error", slog.Any("err", err))
		return err
	}

	rows, err := dbConn.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		slog.Error("sqlite list tables error", slog.Any("err", err))
		return err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if scanErr := rows.Scan(&name); scanErr != nil {
			slog.Error("sqlite scan table error", slog.Any("err", scanErr))
			return scanErr
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		slog.Error("sqlite rows error", slog.Any("err", err))
		return err
	}

	for _, name := range names {
		stmt := "DROP TABLE IF EXISTS " + quoteSQLiteIdent(name)
		if _, err := dbConn.ExecContext(ctx, stmt); err != nil {
			slog.Error("sqlite drop table error",
				slog.String("table", name),
				slog.Any("err", err),
			)
			return err
		}
	}

	if _, err := dbConn.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		slog.Error("sqlite pragma foreign_keys on error", slog.Any("err", err))
		return err
	}
	return nil
}
