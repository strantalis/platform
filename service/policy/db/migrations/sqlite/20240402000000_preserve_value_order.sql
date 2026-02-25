-- +goose Up
-- +goose StatementBegin

ALTER TABLE attribute_definitions ADD COLUMN values_order JSON DEFAULT '[]';
-- SQLite does not support plpgsql; values_order is managed by application logic.

-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin

ALTER TABLE attribute_definitions DROP COLUMN values_order;

-- +goose StatementEnd
