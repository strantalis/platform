package db

import (
	"context"
	"database/sql"
	"regexp"
	"strconv"
)

var (
	sqliteArgPattern        = regexp.MustCompile(`\$\d+`)
	sqliteAliasValuePattern = regexp.MustCompile(`(?i)\bAS\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(\s*value\s*\)`)
)

type sqliteDBWrapper struct {
	db *sql.DB
}

func (w sqliteDBWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	normalized := normalizeSQLiteQuery(query)
	return w.db.ExecContext(ctx, normalized, reorderSQLiteArgs(normalized, args)...)
}

func (w sqliteDBWrapper) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return w.db.PrepareContext(ctx, normalizeSQLiteQuery(query))
}

func (w sqliteDBWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	normalized := normalizeSQLiteQuery(query)
	return w.db.QueryContext(ctx, normalized, reorderSQLiteArgs(normalized, args)...)
}

func (w sqliteDBWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	normalized := normalizeSQLiteQuery(query)
	return w.db.QueryRowContext(ctx, normalized, reorderSQLiteArgs(normalized, args)...)
}

type sqliteTxWrapper struct {
	tx *sql.Tx
}

func (w sqliteTxWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	normalized := normalizeSQLiteQuery(query)
	return w.tx.ExecContext(ctx, normalized, reorderSQLiteArgs(normalized, args)...)
}

func (w sqliteTxWrapper) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return w.tx.PrepareContext(ctx, normalizeSQLiteQuery(query))
}

func (w sqliteTxWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	normalized := normalizeSQLiteQuery(query)
	return w.tx.QueryContext(ctx, normalized, reorderSQLiteArgs(normalized, args)...)
}

func (w sqliteTxWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	normalized := normalizeSQLiteQuery(query)
	return w.tx.QueryRowContext(ctx, normalized, reorderSQLiteArgs(normalized, args)...)
}

func wrapSQLiteDB(db *sql.DB) sqliteDBWrapper {
	return sqliteDBWrapper{db: db}
}

func wrapSQLiteTx(tx *sql.Tx) sqliteTxWrapper {
	return sqliteTxWrapper{tx: tx}
}

func reorderSQLiteArgs(query string, args []interface{}) []interface{} {
	if len(args) == 0 {
		return args
	}
	matches := sqliteArgPattern.FindAllStringIndex(query, -1)
	if len(matches) == 0 {
		return args
	}
	seen := make(map[string]struct{}, len(matches))
	order := make([]string, 0, len(matches))
	for _, match := range matches {
		placeholder := query[match[0]:match[1]]
		if _, ok := seen[placeholder]; ok {
			continue
		}
		seen[placeholder] = struct{}{}
		order = append(order, placeholder)
	}
	ordered := make([]interface{}, 0, len(order))
	for _, placeholder := range order {
		index, err := strconv.Atoi(placeholder[1:])
		if err != nil || index < 1 || index > len(args) {
			ordered = append(ordered, nil)
			continue
		}
		ordered = append(ordered, args[index-1])
	}
	return ordered
}

func normalizeSQLiteQuery(query string) string {
	return sqliteAliasValuePattern.ReplaceAllString(query, "AS $1")
}
