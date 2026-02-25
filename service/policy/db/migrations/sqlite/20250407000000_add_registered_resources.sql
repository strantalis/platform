-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS registered_resources (
  id UUID PRIMARY KEY,
  name VARCHAR NOT NULL UNIQUE,
  metadata JSON,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS registered_resource_values (
  id UUID PRIMARY KEY,
  registered_resource_id UUID NOT NULL REFERENCES registered_resources(id) ON DELETE CASCADE,
  value VARCHAR NOT NULL,
  metadata JSON,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(registered_resource_id, value)
);

-- SQLite does not support plpgsql triggers/functions; updated_at is managed by application code.

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS registered_resource_values;
DROP TABLE IF EXISTS registered_resources;

-- +goose StatementEnd
