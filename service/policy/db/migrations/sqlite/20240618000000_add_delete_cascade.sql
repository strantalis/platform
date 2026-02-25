-- +goose Up
-- +goose StatementBegin

-- SQLite does not support altering foreign key constraints. Emulate
-- ON DELETE CASCADE with triggers for the relationships defined in the
-- Postgres migration.

-- Delete Attribute Definitions and FQNs when their parent Namespace is deleted.
CREATE TRIGGER IF NOT EXISTS trg_attr_namespaces_delete_children
BEFORE DELETE ON attribute_namespaces
FOR EACH ROW
BEGIN
  DELETE FROM attribute_definitions WHERE namespace_id = OLD.id;
  DELETE FROM attribute_fqns WHERE namespace_id = OLD.id;
END;

-- Delete Attribute Values, FQNs, and grants when their parent Definition is deleted.
CREATE TRIGGER IF NOT EXISTS trg_attr_definitions_delete_children
BEFORE DELETE ON attribute_definitions
FOR EACH ROW
BEGIN
  DELETE FROM attribute_values WHERE attribute_definition_id = OLD.id;
  DELETE FROM attribute_fqns WHERE attribute_id = OLD.id;
  DELETE FROM attribute_definition_key_access_grants WHERE attribute_definition_id = OLD.id;
END;

-- Delete child mappings, FQNs, grants, and members when their parent Value is deleted.
CREATE TRIGGER IF NOT EXISTS trg_attr_values_delete_children
BEFORE DELETE ON attribute_values
FOR EACH ROW
BEGIN
  DELETE FROM resource_mappings WHERE attribute_value_id = OLD.id;
  DELETE FROM subject_mappings WHERE attribute_value_id = OLD.id;
  DELETE FROM attribute_fqns WHERE value_id = OLD.id;
  DELETE FROM attribute_value_key_access_grants WHERE attribute_value_id = OLD.id;
  DELETE FROM attribute_value_members WHERE value_id = OLD.id;
  DELETE FROM attribute_value_members WHERE member_id = OLD.id;
END;

-- Delete grants when their parent KAS is deleted.
CREATE TRIGGER IF NOT EXISTS trg_key_access_servers_delete_grants
BEFORE DELETE ON key_access_servers
FOR EACH ROW
BEGIN
  DELETE FROM attribute_definition_key_access_grants WHERE key_access_server_id = OLD.id;
  DELETE FROM attribute_value_key_access_grants WHERE key_access_server_id = OLD.id;
END;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS trg_key_access_servers_delete_grants;
DROP TRIGGER IF EXISTS trg_attr_values_delete_children;
DROP TRIGGER IF EXISTS trg_attr_definitions_delete_children;
DROP TRIGGER IF EXISTS trg_attr_namespaces_delete_children;

-- +goose StatementEnd
