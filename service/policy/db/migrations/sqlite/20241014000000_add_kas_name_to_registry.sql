-- +goose Up
-- +goose StatementBegin

ALTER TABLE key_access_servers
  ADD COLUMN name VARCHAR;

CREATE UNIQUE INDEX IF NOT EXISTS idx_key_access_servers_name
  ON key_access_servers (name);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin


DROP INDEX IF EXISTS idx_key_access_servers_name;

ALTER TABLE key_access_servers
  DROP COLUMN name;

-- +goose StatementEnd
