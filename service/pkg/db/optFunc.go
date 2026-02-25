package db

import (
	"embed"
	"strings"
)

type OptsFunc func(c Config) Config

func WithService(name string) OptsFunc {
	return func(c Config) Config {
		baseSchema := c.Postgres.Schema
		if baseSchema == "" {
			baseSchema = c.Schema
		}
		c.Schema = strings.Join([]string{baseSchema, name}, "_")
		c.Postgres.Schema = c.Schema
		return c
	}
}

func WithMigrations(fs *embed.FS) OptsFunc {
	return func(c Config) Config {
		c.MigrationsFS = fs
		return c
	}
}
