package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/opentdf/platform/protocol/go/common"
	"github.com/opentdf/platform/service/logger"
	"github.com/opentdf/platform/service/pkg/db"
	dbsqlite "github.com/opentdf/platform/service/policy/db/sqlite"
	"go.opentelemetry.io/otel/trace"
)

const (
	stateInactive    transformedState = "INACTIVE"
	stateActive      transformedState = "ACTIVE"
	stateAny         transformedState = "ANY"
	stateUnspecified transformedState = "UNSPECIFIED"
)

type transformedState string

type ListConfig struct {
	limitDefault int32
	limitMax     int32
}

type PolicyDBClient struct {
	dbClient  db.Client
	pgxClient db.PgxClient
	logger    *logger.Logger
	queries   policyQueries
	listCfg   ListConfig
	trace.Tracer
}

func NewClient(c db.Client, logger *logger.Logger, configuredListLimitMax, configuredListLimitDefault int32) (PolicyDBClient, error) {
	listCfg := ListConfig{limitDefault: configuredListLimitDefault, limitMax: configuredListLimitMax}
	switch c.Driver() {
	case db.DriverPostgres:
		pgxClient, ok := c.(db.PgxClient)
		if !ok {
			return PolicyDBClient{}, fmt.Errorf("policy db requires postgres pgx client, got %T", c)
		}
		return PolicyDBClient{
			dbClient:  c,
			pgxClient: pgxClient,
			logger:    logger,
			queries:   pgQueries{Queries: New(pgxClient.PgxPool())},
			listCfg:   listCfg,
			Tracer:    pgxClient.Tracer(),
		}, nil
	case db.DriverSQLite:
		wrapper := wrapSQLiteDB(c.DB())
		return PolicyDBClient{
			dbClient: c,
			logger:   logger,
			queries:  sqliteQueries{q: dbsqlite.New(wrapper), db: wrapper},
			listCfg:  listCfg,
			Tracer:   c.Tracer(),
		}, nil
	default:
		return PolicyDBClient{}, fmt.Errorf("unsupported policy db driver: %s", c.Driver())
	}
}

func (c *PolicyDBClient) RunInTx(ctx context.Context, query func(txClient *PolicyDBClient) error) error {
	switch c.dbClient.Driver() {
	case db.DriverPostgres:
		tx, err := c.pgxClient.PgxPool().Begin(ctx)
		if err != nil {
			return fmt.Errorf("%w: %w", db.ErrTxBeginFailed, err)
		}

		txClient := &PolicyDBClient{
			dbClient:  c.dbClient,
			pgxClient: c.pgxClient,
			logger:    c.logger,
			queries:   c.queries.WithTx(tx),
			listCfg:   c.listCfg,
			Tracer:    c.Tracer,
		}

		if err = query(txClient); err != nil {
			c.logger.WarnContext(ctx, "error during DB transaction, rolling back")
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				return fmt.Errorf("%w, transaction [%w]: %w", db.ErrTxRollbackFailed, err, rollbackErr)
			}
			return err
		}

		if err = tx.Commit(ctx); err != nil {
			return fmt.Errorf("%w: %w", db.ErrTxCommitFailed, err)
		}
		return nil
	case db.DriverSQLite:
		tx, err := c.dbClient.DB().BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("%w: %w", db.ErrTxBeginFailed, err)
		}

		txClient := &PolicyDBClient{
			dbClient: c.dbClient,
			logger:   c.logger,
			queries:  c.queries.WithTx(tx),
			listCfg:  c.listCfg,
			Tracer:   c.Tracer,
		}

		if err = query(txClient); err != nil {
			c.logger.WarnContext(ctx, "error during DB transaction, rolling back")
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				return fmt.Errorf("%w, transaction [%w]: %w", db.ErrTxRollbackFailed, err, rollbackErr)
			}
			return err
		}

		if err = tx.Commit(); err != nil {
			return fmt.Errorf("%w: %w", db.ErrTxCommitFailed, err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported policy db driver: %s", c.dbClient.Driver())
	}
}

// DBClient exposes the underlying database client for reconfiguration.
func (c PolicyDBClient) DBClient() db.Client {
	return c.dbClient
}

func (c PolicyDBClient) Close() {
	c.dbClient.Close()
}

func (c PolicyDBClient) SQLDB() *sql.DB {
	return c.dbClient.DB()
}

func (c PolicyDBClient) PgxPool() db.PgxIface {
	if c.pgxClient == nil {
		return nil
	}
	return c.pgxClient.PgxPool()
}

func (c PolicyDBClient) QueryRow(ctx context.Context, sql string, args []interface{}) (pgx.Row, error) {
	if c.pgxClient == nil {
		if c.dbClient.Driver() == db.DriverSQLite {
			wrapper := wrapSQLiteDB(c.dbClient.DB())
			return sqliteRow{row: wrapper.QueryRowContext(ctx, sql, args...)}, nil
		}
		return nil, fmt.Errorf("pgx pool not available for driver %s", c.dbClient.Driver())
	}
	return c.pgxClient.PgxPool().QueryRow(ctx, sql, args...), nil
}

type sqliteRow struct {
	row *sql.Row
}

func (r sqliteRow) Scan(dest ...any) error {
	return r.row.Scan(dest...)
}

func getDBStateTypeTransformedEnum(state common.ActiveStateEnum) transformedState {
	switch state.String() {
	case common.ActiveStateEnum_ACTIVE_STATE_ENUM_ACTIVE.String():
		return stateActive
	case common.ActiveStateEnum_ACTIVE_STATE_ENUM_INACTIVE.String():
		return stateInactive
	case common.ActiveStateEnum_ACTIVE_STATE_ENUM_ANY.String():
		return stateAny
	case common.ActiveStateEnum_ACTIVE_STATE_ENUM_UNSPECIFIED.String():
		return stateActive
	default:
		return stateActive
	}
}
