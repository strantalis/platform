-- +goose Up
-- +goose StatementBegin

-- Add manager column to provider_config table.
ALTER TABLE provider_config
ADD COLUMN manager VARCHAR(255) NOT NULL DEFAULT 'opentdf.io/unspecified';

-- Backfill default manager for existing rows.
UPDATE provider_config
SET manager = 'opentdf.io/unspecified'
WHERE manager IS NULL;

-- Replace unique constraint on provider_name with composite uniqueness.
DROP INDEX IF EXISTS idx_provider_config_provider_name;

CREATE UNIQUE INDEX IF NOT EXISTS idx_provider_config_provider_name_manager
  ON provider_config (provider_name, manager);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_provider_config_provider_name_manager;

CREATE UNIQUE INDEX IF NOT EXISTS idx_provider_config_provider_name
  ON provider_config (provider_name);

ALTER TABLE provider_config
DROP COLUMN manager;

-- +goose StatementEnd
