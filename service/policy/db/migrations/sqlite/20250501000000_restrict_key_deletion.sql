-- +goose Up
-- +goose StatementBegin

-- SQLite does not support altering FK constraints. Emulate ON DELETE RESTRICT
-- with a trigger that aborts deletion when dependent keys exist.
CREATE TRIGGER IF NOT EXISTS trg_key_access_servers_restrict_delete
BEFORE DELETE ON key_access_servers
FOR EACH ROW
BEGIN
  SELECT RAISE(ABORT, 'key_access_server has keys')
  WHERE EXISTS (
    SELECT 1 FROM key_access_server_keys WHERE key_access_server_id = OLD.id
  );
END;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS trg_key_access_servers_restrict_delete;

-- +goose StatementEnd
