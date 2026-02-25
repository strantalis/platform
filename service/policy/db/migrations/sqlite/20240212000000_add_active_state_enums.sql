-- +goose Up
-- +goose StatementBegin

ALTER TABLE attribute_namespaces ADD COLUMN active bool NOT NULL DEFAULT true;
ALTER TABLE attribute_definitions ADD COLUMN active bool NOT NULL DEFAULT true;
ALTER TABLE attribute_values ADD COLUMN active bool NOT NULL DEFAULT true;

-- SQLite does not support plpgsql functions; cascade behavior is handled in application code.

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- If rolling back, all deactivations should be hard deletes.

ALTER TABLE attribute_namespaces DROP COLUMN active;
ALTER TABLE attribute_definitions DROP COLUMN active;
ALTER TABLE attribute_values DROP COLUMN active;

-- +goose StatementEnd
