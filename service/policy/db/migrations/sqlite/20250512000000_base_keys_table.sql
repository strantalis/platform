-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS base_keys (
  id UUID CONSTRAINT base_key_pkey PRIMARY KEY,
  key_access_server_key_id UUID CONSTRAINT key_access_server_key_id_fkey REFERENCES key_access_server_keys(id) ON DELETE RESTRICT
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS base_keys;

-- +goose StatementEnd
