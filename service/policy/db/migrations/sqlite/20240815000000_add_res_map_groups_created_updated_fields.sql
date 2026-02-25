-- +goose Up
-- +goose StatementBegin

ALTER TABLE resource_mapping_groups ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE resource_mapping_groups ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- SQLite does not support plpgsql triggers/functions; updated_at is managed by application code.

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE resource_mapping_groups DROP COLUMN created_at;
ALTER TABLE resource_mapping_groups DROP COLUMN updated_at;

-- +goose StatementEnd
