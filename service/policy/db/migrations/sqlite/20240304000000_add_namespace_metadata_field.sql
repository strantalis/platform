-- +goose Up
-- +goose StatementBegin

ALTER TABLE attribute_namespaces ADD COLUMN metadata JSON;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE attribute_namespaces DROP COLUMN metadata;

-- +goose StatementEnd
