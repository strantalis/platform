-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS idx_provider_config_provider_name
  ON provider_config (provider_name);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_provider_config_provider_name;
-- +goose StatementEnd
