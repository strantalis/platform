-- +goose Up
-- +goose StatementBegin
ALTER TABLE key_access_servers ADD COLUMN identifier VARCHAR NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
