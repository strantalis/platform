package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func migrationInit(ctx context.Context, dialect goose.Dialect, db *sql.DB, migrations fs.FS, runMigrations bool, closeFunc func() error) (*goose.Provider, int64, func(), error) {
	if !runMigrations {
		return nil, 0, nil, errors.New("migrations are disabled")
	}

	// Create the goose provider
	provider, err := goose.NewProvider(dialect, db, migrations)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to create goose provider: %w", err)
	}

	// Get the current version
	v, e := provider.GetDBVersion(ctx)
	if e != nil {
		return nil, 0, nil, errors.Join(errors.New("failed to get current version"), e)
	}
	slog.Info("migration db info", slog.Int64("current_version", v))

	// Return the provider, version, and close function
	return provider, v, func() {
		if closeFunc == nil {
			return
		}
		if err := closeFunc(); err != nil {
			slog.Error("failed to close connection", slog.Any("err", err))
		}
	}, nil
}

func selectMigrationsFS(migrations fs.FS, driver string) (fs.FS, error) {
	if migrations == nil {
		return nil, errors.New("migrations FS is required to run migrations")
	}
	subdir := driver
	switch driver {
	case DriverPostgres:
		subdir = "postgres"
	case DriverSQLite:
		subdir = "sqlite"
	}
	return fs.Sub(migrations, subdir)
}

// RunMigrations runs the migrations for the schema
// Schema will be created if it doesn't exist
func (c *PostgresClient) RunMigrations(ctx context.Context, migrations fs.FS) (int, error) {
	migrationFS, err := selectMigrationsFS(migrations, DriverPostgres)
	if err != nil {
		return 0, err
	}
	slog.Info("running migration up",
		slog.String("schema", c.pgConfig.Schema),
		slog.String("database", c.pgConfig.Database),
	)

	// Create schema if it doesn't exist
	q := "CREATE SCHEMA IF NOT EXISTS " + pgx.Identifier{c.pgConfig.Schema}.Sanitize()
	tag, err := c.Pgx.Exec(ctx, q)
	if err != nil {
		slog.ErrorContext(ctx,
			"error while running command",
			slog.String("command", q),
			slog.Any("error", err),
		)
		return 0, err
	}
	applied := int(tag.RowsAffected())

	pool, ok := c.Pgx.(*pgxpool.Pool)
	if !ok || pool == nil {
		return applied, errors.New("failed to cast pgxpool.Pool")
	}
	conn := stdlib.OpenDBFromPool(pool)
	provider, version, closeProvider, err := migrationInit(ctx, goose.DialectPostgres, conn, migrationFS, c.config.RunMigrations, conn.Close)
	if err != nil {
		slog.Error("failed to create goose provider", slog.Any("err", err))
		return 0, err
	}
	defer closeProvider()

	res, err := provider.Up(ctx)
	if err != nil {
		return applied, errors.Join(errors.New("failed to run migrations"), err)
	}

	if len(res) != 0 {
		version = res[len(res)-1].Source.Version
	}

	for _, r := range res {
		if r.Error != nil {
			return applied, errors.Join(errors.New("failed to run migrations"), err)
		}
		if !r.Empty {
			applied++
		}
	}
	c.ranMigrations = true
	slog.Info("migration up complete", slog.Int64("post_op_version", version))
	return applied, nil
}

func (c *PostgresClient) MigrationStatus(ctx context.Context) ([]*goose.MigrationStatus, error) {
	slog.Info("running migrations status",
		slog.String("schema", c.pgConfig.Schema),
		slog.String("database", c.pgConfig.Database),
	)
	pool, ok := c.Pgx.(*pgxpool.Pool)
	if !ok || pool == nil {
		return nil, errors.New("failed to cast pgxpool.Pool")
	}
	conn := stdlib.OpenDBFromPool(pool)
	provider, _, closeProvider, err := migrationInit(ctx, goose.DialectPostgres, conn, nil, c.config.RunMigrations, conn.Close)
	if err != nil {
		slog.Error("failed to create goose provider", slog.Any("err", err))
		return nil, err
	}
	defer closeProvider()

	return provider.Status(ctx)
}

func (c *PostgresClient) MigrationDown(ctx context.Context, migrations fs.FS) error {
	slog.Info("running migration down",
		slog.String("schema", c.pgConfig.Schema),
		slog.String("database", c.pgConfig.Database),
	)
	migrationFS, err := selectMigrationsFS(migrations, DriverPostgres)
	if err != nil {
		return err
	}
	pool, ok := c.Pgx.(*pgxpool.Pool)
	if !ok || pool == nil {
		return errors.New("failed to cast pgxpool.Pool")
	}
	conn := stdlib.OpenDBFromPool(pool)
	provider, _, closeProvider, err := migrationInit(ctx, goose.DialectPostgres, conn, migrationFS, c.config.RunMigrations, conn.Close)
	if err != nil {
		slog.Error("failed to create goose provider", slog.Any("err", err))
		return err
	}
	defer closeProvider()

	res, err := provider.Down(ctx)
	if err != nil {
		return errors.Join(errors.New("failed to run migrations"), err)
	}
	if res.Error != nil {
		return errors.Join(errors.New("failed to run migrations"), res.Error)
	}

	slog.Info("migration down complete", slog.Int64("post_op_version", res.Source.Version))
	return nil
}

func (c *PostgresClient) MigrationsEnabled() bool {
	return c.config.RunMigrations
}

func (c *PostgresClient) RanMigrations() bool {
	return c.ranMigrations
}
