-- +goose Up
-- +goose StatementBegin

ALTER TABLE sym_key ADD COLUMN expiration TIMESTAMP;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE sym_key DROP COLUMN expiration;

-- +goose StatementEnd
