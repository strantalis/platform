-- +goose Up
-- +goose StatementBegin

-- SQLite does not support altering FK constraints. The existing FK from
-- key_access_server_keys.provider_config_id to provider_config.id already
-- defaults to RESTRICT/NO ACTION, so no changes are required.

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- No-op for SQLite.

-- +goose StatementEnd
