-- +goose Up
-- +goose StatementBegin

DROP TRIGGER IF EXISTS trg_attr_values_delete_children;

DROP TABLE IF EXISTS attribute_value_members;

ALTER TABLE attribute_values DROP COLUMN members;

CREATE TRIGGER IF NOT EXISTS trg_attr_values_delete_children
BEFORE DELETE ON attribute_values
FOR EACH ROW
BEGIN
  DELETE FROM resource_mappings WHERE attribute_value_id = OLD.id;
  DELETE FROM subject_mappings WHERE attribute_value_id = OLD.id;
  DELETE FROM attribute_fqns WHERE value_id = OLD.id;
  DELETE FROM attribute_value_key_access_grants WHERE attribute_value_id = OLD.id;
END;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS attribute_value_members
(
    id UUID PRIMARY KEY,
    value_id UUID NOT NULL REFERENCES attribute_values(id) ON DELETE CASCADE,
    member_id UUID NOT NULL REFERENCES attribute_values(id) ON DELETE CASCADE,
    UNIQUE (value_id, member_id)
);

ALTER TABLE attribute_values ADD COLUMN members TEXT;

DROP TRIGGER IF EXISTS trg_attr_values_delete_children;

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

-- +goose StatementEnd
