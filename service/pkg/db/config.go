package db

import (
	"embed"
	"log/slog"
	"strings"
)

const (
	DriverPostgres = "postgres"
	DriverSQLite   = "sqlite"
)

// PoolConfig holds all connection pool related configuration.
type PoolConfig struct {
	MaxConns          int32 `mapstructure:"max_connection_count" json:"max_connection_count" default:"4"`
	MinConns          int32 `mapstructure:"min_connection_count" json:"min_connection_count" default:"0"`
	MinIdleConns      int32 `mapstructure:"min_idle_connections_count" json:"min_idle_connections_count" default:"0"`
	MaxConnLifetime   int   `mapstructure:"max_connection_lifetime_seconds" json:"max_connection_lifetime_seconds" default:"3600"`
	MaxConnIdleTime   int   `mapstructure:"max_connection_idle_seconds" json:"max_connection_idle_seconds" default:"1800"`
	HealthCheckPeriod int   `mapstructure:"health_check_period_seconds" json:"health_check_period_seconds" default:"60"`
}

// PostgresConfig holds Postgres connection configuration.
type PostgresConfig struct {
	Host           string     `mapstructure:"host" json:"host" default:"localhost"`
	Port           int        `mapstructure:"port" json:"port" default:"5432"`
	Database       string     `mapstructure:"database" json:"database" default:"opentdf"`
	User           string     `mapstructure:"user" json:"user" default:"postgres"`
	Password       string     `mapstructure:"password" json:"password" default:"changeme"`
	SSLMode        string     `mapstructure:"sslmode" json:"sslmode" default:"prefer"`
	Schema         string     `mapstructure:"schema" json:"schema" default:"opentdf"`
	ConnectTimeout int        `mapstructure:"connect_timeout_seconds" json:"connect_timeout_seconds" default:"15"`
	Pool           PoolConfig `mapstructure:"pool" json:"pool"`
}

// SQLiteConfig holds SQLite connection configuration.
type SQLiteConfig struct {
	Path          string `mapstructure:"path" json:"path" default:"opentdf.db"`
	InMemory      bool   `mapstructure:"in_memory" json:"in_memory" default:"false"`
	Cache         string `mapstructure:"cache" json:"cache" default:"shared"`
	Mode          string `mapstructure:"mode" json:"mode" default:"rwc"`
	BusyTimeoutMS int    `mapstructure:"busy_timeout_ms" json:"busy_timeout_ms" default:"5000"`
	ForeignKeys   bool   `mapstructure:"foreign_keys" json:"foreign_keys" default:"true"`
	JournalMode   string `mapstructure:"journal_mode" json:"journal_mode" default:"WAL"`
}

// Config represents the configuration settings for the database.
type Config struct {
	Driver   string         `mapstructure:"driver" json:"driver" default:"postgres"`
	Postgres PostgresConfig `mapstructure:"postgres" json:"postgres"`
	SQLite   SQLiteConfig   `mapstructure:"sqlite" json:"sqlite"`

	// Legacy flat fields (deprecated). These are mapped into Postgres when Driver is unset.
	Host           string     `mapstructure:"host" json:"host" default:"localhost"`
	Port           int        `mapstructure:"port" json:"port" default:"5432"`
	Database       string     `mapstructure:"database" json:"database" default:"opentdf"`
	User           string     `mapstructure:"user" json:"user" default:"postgres"`
	Password       string     `mapstructure:"password" json:"password" default:"changeme"`
	SSLMode        string     `mapstructure:"sslmode" json:"sslmode" default:"prefer"`
	Schema         string     `mapstructure:"schema" json:"schema" default:"opentdf"`
	ConnectTimeout int        `mapstructure:"connect_timeout_seconds" json:"connect_timeout_seconds" default:"15"`
	Pool           PoolConfig `mapstructure:"pool" json:"pool"`

	RunMigrations    bool      `mapstructure:"runMigrations" json:"runMigrations" default:"true"`
	MigrationsFS     *embed.FS `mapstructure:"-" json:"-"`
	VerifyConnection bool      `mapstructure:"verifyConnection" json:"verifyConnection" default:"true"`
}

// Normalize applies defaults and legacy field mapping.
func (c *Config) Normalize() {
	if c.Driver == "" {
		c.Driver = DriverPostgres
	}
	c.Driver = strings.ToLower(c.Driver)

	// Map legacy flat fields into Postgres config when not explicitly set.
	if c.Postgres.Host == "" {
		c.Postgres.Host = c.Host
	}
	if c.Postgres.Port == 0 {
		c.Postgres.Port = c.Port
	}
	if c.Postgres.Database == "" {
		c.Postgres.Database = c.Database
	}
	if c.Postgres.User == "" {
		c.Postgres.User = c.User
	}
	if c.Postgres.Password == "" {
		c.Postgres.Password = c.Password
	}
	if c.Postgres.SSLMode == "" {
		c.Postgres.SSLMode = c.SSLMode
	}
	if c.Postgres.Schema == "" {
		c.Postgres.Schema = c.Schema
	}
	if c.Postgres.ConnectTimeout == 0 {
		c.Postgres.ConnectTimeout = c.ConnectTimeout
	}
	if c.Postgres.Pool == (PoolConfig{}) {
		c.Postgres.Pool = c.Pool
	}

	if c.SQLite.Path == "" {
		c.SQLite.Path = "opentdf.db"
	}
	if c.SQLite.Cache == "" {
		c.SQLite.Cache = "shared"
	}
	if c.SQLite.Mode == "" {
		c.SQLite.Mode = "rwc"
	}
	if c.SQLite.BusyTimeoutMS == 0 {
		c.SQLite.BusyTimeoutMS = 5000
	}
	if c.SQLite.JournalMode == "" {
		c.SQLite.JournalMode = "WAL"
	}
}

func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("driver", c.Driver),
		slog.Group("postgres",
			slog.String("host", c.Postgres.Host),
			slog.Int("port", c.Postgres.Port),
			slog.String("database", c.Postgres.Database),
			slog.String("user", c.Postgres.User),
			slog.String("password", "[REDACTED]"),
			slog.String("sslmode", c.Postgres.SSLMode),
			slog.String("schema", c.Postgres.Schema),
			slog.Int("connect_timeout_seconds", c.Postgres.ConnectTimeout),
			slog.Group("pool",
				slog.Int("max_connection_count", int(c.Postgres.Pool.MaxConns)),
				slog.Int("min_connection_count", int(c.Postgres.Pool.MinConns)),
				slog.Int("max_connection_lifetime_seconds", c.Postgres.Pool.MaxConnLifetime),
				slog.Int("max_connection_idle_seconds", c.Postgres.Pool.MaxConnIdleTime),
				slog.Int("health_check_period_seconds", c.Postgres.Pool.HealthCheckPeriod),
			),
		),
		slog.Group("sqlite",
			slog.String("path", c.SQLite.Path),
			slog.Bool("in_memory", c.SQLite.InMemory),
			slog.String("cache", c.SQLite.Cache),
			slog.String("mode", c.SQLite.Mode),
			slog.Int("busy_timeout_ms", c.SQLite.BusyTimeoutMS),
			slog.Bool("foreign_keys", c.SQLite.ForeignKeys),
			slog.String("journal_mode", c.SQLite.JournalMode),
		),
		slog.Bool("runMigrations", c.RunMigrations),
		slog.Bool("verifyConnection", c.VerifyConnection),
	)
}
