package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/opentdf/platform/service/logger"
	"github.com/pressly/goose/v3"
	"go.opentelemetry.io/otel/trace"
)

// Client is the driver-agnostic database client interface used by the platform.
type Client interface {
	Driver() string
	DB() *sql.DB
	Tracer() trace.Tracer
	Close()
	RunMigrations(ctx context.Context, migrations fs.FS) (int, error)
	MigrationStatus(ctx context.Context) ([]*MigrationStatus, error)
	MigrationDown(ctx context.Context, migrations fs.FS) error
	MigrationsEnabled() bool
	RanMigrations() bool
}

// PgxClient provides access to pgx-specific features for Postgres-backed stores.
type PgxClient interface {
	Client
	PgxPool() PgxIface
}

// New constructs a driver-agnostic database client based on configuration.
func New(ctx context.Context, config Config, logCfg logger.Config, tracer *trace.Tracer, o ...OptsFunc) (Client, error) {
	for _, f := range o {
		config = f(config)
	}
	config.Normalize()

	switch config.Driver {
	case DriverPostgres:
		return newPostgresClient(ctx, config, logCfg, tracer)
	case DriverSQLite:
		return newSQLiteClient(ctx, config, logCfg, tracer)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}
}

// MigrationStatus is an alias to avoid leaking goose types outside the db package.
type MigrationStatus = goose.MigrationStatus
