package db

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"github.com/opentdf/platform/service/logger"
	"github.com/pressly/goose/v3"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const sqliteDriverName = "sqlite3_opentdf"

func init() {
	sql.Register(sqliteDriverName, &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			if err := conn.RegisterFunc("gen_random_uuid", func() (string, error) {
				return uuid.NewString(), nil
			}, false); err != nil {
				return err
			}

			if err := registerSQLiteJSONFunctions(conn); err != nil {
				return err
			}
			if err := registerSQLiteAggregates(conn); err != nil {
				return err
			}

			return nil
		},
	})
}

type SQLiteClient struct {
	Logger        *logger.Logger
	config        Config
	sqliteConfig  SQLiteConfig
	ranMigrations bool
	SQLDB         *sql.DB
	tracer        trace.Tracer
}

func newSQLiteClient(ctx context.Context, config Config, logCfg logger.Config, tracer *trace.Tracer) (*SQLiteClient, error) {
	c := SQLiteClient{
		config:       config,
		sqliteConfig: config.SQLite,
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
	c.Logger = l.With("driver", DriverSQLite)

	dsn := buildSQLiteDSN(c.sqliteConfig)
	db, err := sql.Open(sqliteDriverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}
	c.SQLDB = db

	applySQLitePoolDefaults(db, c.sqliteConfig)

	if c.config.VerifyConnection {
		if err := db.PingContext(ctx); err != nil {
			return nil, fmt.Errorf("failed to connect to sqlite database: %w", err)
		}
	}

	return &c, nil
}

func (c *SQLiteClient) Driver() string { return DriverSQLite }

func (c *SQLiteClient) DB() *sql.DB { return c.SQLDB }

func (c *SQLiteClient) Tracer() trace.Tracer { return c.tracer }

func (c *SQLiteClient) Close() {
	if c.SQLDB != nil {
		_ = c.SQLDB.Close()
	}
}

func (c *SQLiteClient) RunMigrations(ctx context.Context, migrations fs.FS) (int, error) {
	migrationFS, err := selectMigrationsFS(migrations, DriverSQLite)
	if err != nil {
		return 0, err
	}
	slog.Info("running sqlite migration up")

	provider, version, closeProvider, err := migrationInit(ctx, goose.DialectSQLite3, c.SQLDB, migrationFS, c.config.RunMigrations, nil)
	if err != nil {
		slog.Error("failed to create goose provider", slog.Any("err", err))
		return 0, err
	}
	defer closeProvider()

	res, err := provider.Up(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to run migrations: %w", err)
	}

	if len(res) != 0 {
		version = res[len(res)-1].Source.Version
	}

	applied := 0
	for _, r := range res {
		if r.Error != nil {
			return applied, fmt.Errorf("failed to run migrations: %w", r.Error)
		}
		if !r.Empty {
			applied++
		}
	}
	c.ranMigrations = true
	slog.Info("sqlite migration up complete", slog.Int64("post_op_version", version))
	return applied, nil
}

func (c *SQLiteClient) MigrationStatus(ctx context.Context) ([]*MigrationStatus, error) {
	slog.Info("running sqlite migrations status")
	provider, _, closeProvider, err := migrationInit(ctx, goose.DialectSQLite3, c.SQLDB, nil, c.config.RunMigrations, nil)
	if err != nil {
		slog.Error("failed to create goose provider", slog.Any("err", err))
		return nil, err
	}
	defer closeProvider()

	return provider.Status(ctx)
}

func (c *SQLiteClient) MigrationDown(ctx context.Context, migrations fs.FS) error {
	slog.Info("running sqlite migration down")
	migrationFS, err := selectMigrationsFS(migrations, DriverSQLite)
	if err != nil {
		return err
	}
	provider, _, closeProvider, err := migrationInit(ctx, goose.DialectSQLite3, c.SQLDB, migrationFS, c.config.RunMigrations, nil)
	if err != nil {
		slog.Error("failed to create goose provider", slog.Any("err", err))
		return err
	}
	defer closeProvider()

	res, err := provider.Down(ctx)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	if res.Error != nil {
		return fmt.Errorf("failed to run migrations: %w", res.Error)
	}

	slog.Info("sqlite migration down complete", slog.Int64("post_op_version", res.Source.Version))
	return nil
}

func (c *SQLiteClient) MigrationsEnabled() bool { return c.config.RunMigrations }

func (c *SQLiteClient) RanMigrations() bool { return c.ranMigrations }

func buildSQLiteDSN(cfg SQLiteConfig) string {
	if cfg.InMemory {
		v := url.Values{}
		v.Set("mode", "memory")
		if cfg.Cache != "" {
			v.Set("cache", cfg.Cache)
		}
		if cfg.BusyTimeoutMS > 0 {
			v.Set("_busy_timeout", strconv.Itoa(cfg.BusyTimeoutMS))
		}
		if cfg.ForeignKeys {
			v.Set("_foreign_keys", "1")
		}
		if cfg.JournalMode != "" {
			v.Set("_journal_mode", cfg.JournalMode)
		}
		name := cfg.Path
		if name == "" {
			name = "opentdf"
		}
		return "file:" + url.PathEscape(name) + "?" + v.Encode()
	}

	v := url.Values{}
	if cfg.Cache != "" {
		v.Set("cache", cfg.Cache)
	}
	if cfg.Mode != "" {
		v.Set("mode", cfg.Mode)
	}
	if cfg.BusyTimeoutMS > 0 {
		v.Set("_busy_timeout", strconv.Itoa(cfg.BusyTimeoutMS))
	}
	if cfg.ForeignKeys {
		v.Set("_foreign_keys", "1")
	}
	if cfg.JournalMode != "" {
		v.Set("_journal_mode", cfg.JournalMode)
	}
	return fmt.Sprintf("file:%s?%s", cfg.Path, v.Encode())
}

func applySQLitePoolDefaults(db *sql.DB, cfg SQLiteConfig) {
	if cfg.InMemory {
		// Keep at least one connection alive for shared in-memory DB.
		db.SetMaxIdleConns(1)
		db.SetConnMaxLifetime(0)
	}
	// SQLite is sensitive to high concurrency; keep it conservative by default.
	db.SetMaxOpenConns(1)
}

func registerSQLiteJSONFunctions(conn *sqlite3.SQLiteConn) error {
	if err := conn.RegisterFunc("JSON_BUILD_OBJECT", jsonBuildObject, true); err != nil {
		return err
	}
	if err := conn.RegisterFunc("JSONB_BUILD_OBJECT", jsonBuildObject, true); err != nil {
		return err
	}
	if err := conn.RegisterFunc("JSON_BUILD_ARRAY", jsonBuildArray, true); err != nil {
		return err
	}
	if err := conn.RegisterFunc("JSONB_BUILD_ARRAY", jsonBuildArray, true); err != nil {
		return err
	}
	if err := conn.RegisterFunc("JSON_STRIP_NULLS", jsonStripNulls, true); err != nil {
		return err
	}
	if err := conn.RegisterFunc("REGEXP_REPLACE", regexReplace, true); err != nil {
		return err
	}
	if err := conn.RegisterFunc("DECODE", decodeBase64, true); err != nil {
		return err
	}
	if err := conn.RegisterFunc("CONVERT_FROM", convertFrom, true); err != nil {
		return err
	}
	if err := conn.RegisterFunc("ARRAY_POSITION", arrayPosition, true); err != nil {
		return err
	}
	return nil
}

func registerSQLiteAggregates(conn *sqlite3.SQLiteConn) error {
	if err := conn.RegisterAggregator("JSON_AGG", func() *jsonAgg { return &jsonAgg{} }, true); err != nil {
		return err
	}
	if err := conn.RegisterAggregator("JSONB_AGG", func() *jsonAgg { return &jsonAgg{} }, true); err != nil {
		return err
	}
	return nil
}

func jsonBuildObject(args ...interface{}) ([]byte, error) {
	if len(args)%2 != 0 {
		return []byte("{}"), nil
	}
	obj := map[string]interface{}{}
	for i := 0; i < len(args); i += 2 {
		key := toString(args[i])
		value := normalizeJSONValue(args[i+1])
		if isBooleanJSONKey(key) {
			value = normalizeBoolValue(value)
		}
		if key == "created_at" || key == "updated_at" {
			value = normalizeTimestampValue(value)
		}
		if key == "labels" && isEmptyStringValue(value) {
			value = nil
		}
		obj[key] = value
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return []byte("{}"), err
	}
	return b, nil
}

func jsonBuildArray(args ...interface{}) ([]byte, error) {
	items := make([]interface{}, 0, len(args))
	for _, arg := range args {
		items = append(items, normalizeJSONValue(arg))
	}
	b, err := json.Marshal(items)
	if err != nil {
		return []byte("[]"), err
	}
	return b, nil
}

func jsonStripNulls(v interface{}) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	raw := toString(v)
	if raw == "" {
		return []byte("null"), nil
	}
	var data interface{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		//nolint:nilerr // best-effort normalization; fall back to raw JSON on parse errors
		return []byte(raw), nil
	}
	clean := stripNulls(data)
	b, err := json.Marshal(clean)
	if err != nil {
		//nolint:nilerr // best-effort normalization; fall back to raw JSON on marshal errors
		return []byte(raw), nil
	}
	return b, nil
}

func normalizeJSONValue(v interface{}) interface{} {
	switch t := v.(type) {
	case nil:
		return nil
	case []byte:
		if json.Valid(t) {
			var out interface{}
			if err := json.Unmarshal(t, &out); err == nil {
				return out
			}
		}
		return string(t)
	case string:
		if json.Valid([]byte(t)) {
			var out interface{}
			if err := json.Unmarshal([]byte(t), &out); err == nil {
				return out
			}
		}
		return t
	default:
		return v
	}
}

func normalizeTimestampValue(v interface{}) interface{} {
	switch t := v.(type) {
	case time.Time:
		return t.UTC().Format(time.RFC3339Nano)
	case string:
		if formatted, ok := formatTimestampString(t); ok {
			return formatted
		}
		return t
	default:
		return v
	}
}

func formatTimestampString(input string) (string, bool) {
	if input == "" {
		return input, true
	}
	if ts, err := time.Parse(time.RFC3339Nano, input); err == nil {
		return ts.UTC().Format(time.RFC3339Nano), true
	}
	if ts, err := time.Parse(time.RFC3339, input); err == nil {
		return ts.UTC().Format(time.RFC3339Nano), true
	}
	if ts, err := time.ParseInLocation("2006-01-02 15:04:05", input, time.UTC); err == nil {
		return ts.UTC().Format(time.RFC3339Nano), true
	}
	if ts, err := time.ParseInLocation("2006-01-02T15:04:05", input, time.UTC); err == nil {
		return ts.UTC().Format(time.RFC3339Nano), true
	}
	return input, false
}

func isBooleanJSONKey(key string) bool {
	switch key {
	case "active", "allow_traversal", "is_standard":
		return true
	default:
		return false
	}
}

func normalizeBoolValue(v interface{}) interface{} {
	switch t := v.(type) {
	case bool:
		return t
	case int:
		return t != 0
	case int32:
		return t != 0
	case int64:
		return t != 0
	case float32:
		return t != 0
	case float64:
		return t != 0
	case string:
		lower := strings.ToLower(strings.TrimSpace(t))
		if lower == "1" || lower == "true" {
			return true
		}
		if lower == "0" || lower == "false" || lower == "" {
			return false
		}
		return t
	case []byte:
		return normalizeBoolValue(string(t))
	default:
		return v
	}
}

func isEmptyStringValue(v interface{}) bool {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t) == ""
	case []byte:
		return len(bytes.TrimSpace(t)) == 0
	default:
		return false
	}
}

func stripNulls(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		out := map[string]interface{}{}
		for k, val := range t {
			if val == nil {
				continue
			}
			out[k] = stripNulls(val)
		}
		return out
	case []interface{}:
		out := make([]interface{}, 0, len(t))
		for _, val := range t {
			out = append(out, stripNulls(val))
		}
		return out
	default:
		return v
	}
}

func regexReplace(input, pattern, replacement interface{}) (string, error) {
	re, err := regexp.Compile(toString(pattern))
	if err != nil {
		return "", err
	}
	return re.ReplaceAllString(toString(input), toString(replacement)), nil
}

func decodeBase64(input, encoding interface{}) ([]byte, error) {
	if input == nil {
		return nil, nil
	}
	enc := strings.ToLower(toString(encoding))
	if enc == "" || enc == "base64" {
		return base64.StdEncoding.DecodeString(toString(input))
	}
	return nil, fmt.Errorf("unsupported decode encoding: %s", enc)
}

func convertFrom(input, encoding interface{}) (string, error) {
	if input == nil {
		return "", nil
	}
	enc := strings.ToLower(toString(encoding))
	if enc == "" || enc == "utf8" || enc == "utf-8" {
		switch v := input.(type) {
		case []byte:
			return string(v), nil
		default:
			return toString(input), nil
		}
	}
	return "", fmt.Errorf("unsupported convert_from encoding: %s", enc)
}

func arrayPosition(arrayJSON, value interface{}) (interface{}, error) {
	raw := toString(arrayJSON)
	if raw == "" {
		//nolint:nilnil // represent SQL NULL when input array is empty
		return nil, nil
	}
	var items []interface{}
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		//nolint:nilerr,nilnil // invalid JSON input yields SQL NULL (best-effort behavior)
		return nil, nil
	}
	target := toString(value)
	for i, item := range items {
		if toString(item) == target {
			return i + 1, nil
		}
	}
	//nolint:nilnil // SQL NULL when value not found
	return nil, nil
}

type jsonAgg struct {
	items []json.RawMessage
}

func (a *jsonAgg) Step(value interface{}) error {
	if value == nil {
		return nil
	}
	raw := toString(value)
	if raw == "" {
		return nil
	}
	if json.Valid([]byte(raw)) {
		a.items = append(a.items, json.RawMessage(raw))
		return nil
	}
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	a.items = append(a.items, json.RawMessage(b))
	return nil
}

func (a *jsonAgg) Done() ([]byte, error) {
	if len(a.items) == 0 {
		return []byte("[]"), nil
	}
	b, err := json.Marshal(a.items)
	if err != nil {
		return []byte("[]"), err
	}
	return b, nil
}

func toString(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return fmt.Sprint(v)
	}
}
