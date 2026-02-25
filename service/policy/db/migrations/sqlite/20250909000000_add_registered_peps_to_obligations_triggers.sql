-- +goose Up
-- +goose StatementBegin
-- Add client_id column to obligation_triggers table
ALTER TABLE obligation_triggers
ADD COLUMN client_id TEXT DEFAULT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop the client_id column
ALTER TABLE obligation_triggers
DROP COLUMN client_id;

-- +goose StatementEnd
