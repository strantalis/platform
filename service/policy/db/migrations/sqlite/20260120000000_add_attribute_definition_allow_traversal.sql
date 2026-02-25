-- +goose Up
-- +goose StatementBegin
ALTER TABLE attribute_definitions
    ADD COLUMN allow_traversal BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE attribute_definitions
    DROP COLUMN allow_traversal;
-- +goose StatementEnd
