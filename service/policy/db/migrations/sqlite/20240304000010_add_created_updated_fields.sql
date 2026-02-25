-- +goose Up
-- +goose StatementBegin

-- NOTE: pre-1.0 not crucial to migrate existing timestamps stored in the metadata JSON column into the new columns.

-- Add new columns for created and updated fields for tables:
-- 1. attribute_namespaces
-- 2. attribute_definitions
-- 3. attribute_values
-- 4. key_access_servers
-- 5. resource_mappings
-- 6. subject_mappings

ALTER TABLE attribute_namespaces ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE attribute_namespaces ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE attribute_definitions ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE attribute_definitions ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE attribute_values ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE attribute_values ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE key_access_servers ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE key_access_servers ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE resource_mappings ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE resource_mappings ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE subject_mappings ADD COLUMN created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE subject_mappings ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- SQLite does not support plpgsql triggers/functions; updated_at is managed by application code.

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE attribute_namespaces DROP COLUMN created_at;
ALTER TABLE attribute_namespaces DROP COLUMN updated_at;

ALTER TABLE attribute_definitions DROP COLUMN created_at;
ALTER TABLE attribute_definitions DROP COLUMN updated_at;

ALTER TABLE attribute_values DROP COLUMN created_at;
ALTER TABLE attribute_values DROP COLUMN updated_at;

ALTER TABLE key_access_servers DROP COLUMN created_at;
ALTER TABLE key_access_servers DROP COLUMN updated_at;

ALTER TABLE resource_mappings DROP COLUMN created_at;
ALTER TABLE resource_mappings DROP COLUMN updated_at;

ALTER TABLE subject_mappings DROP COLUMN created_at;
ALTER TABLE subject_mappings DROP COLUMN updated_at;

-- +goose StatementEnd
