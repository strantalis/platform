-- +goose Up
-- +goose StatementBegin

ALTER TABLE subject_condition_set ADD COLUMN selector_values JSON;

CREATE INDEX IF NOT EXISTS idx_subject_condition_set_selector_values
ON subject_condition_set(selector_values);

CREATE INDEX IF NOT EXISTS idx_subject_mappings_attribute_value_id
ON subject_mappings(attribute_value_id);

CREATE INDEX IF NOT EXISTS idx_subject_mappings_subject_condition_set_id
ON subject_mappings(subject_condition_set_id);

CREATE INDEX IF NOT EXISTS idx_attribute_values_attribute_definition_id
ON attribute_values(attribute_definition_id);

CREATE INDEX IF NOT EXISTS idx_attribute_fqns_value_id
ON attribute_fqns(value_id);

CREATE INDEX IF NOT EXISTS idx_subject_mapping_actions_mapping_action
ON subject_mapping_actions(subject_mapping_id, action_id);

CREATE INDEX IF NOT EXISTS idx_attribute_namespaces_active
ON attribute_namespaces(active);

CREATE INDEX IF NOT EXISTS idx_attribute_definitions_active
ON attribute_definitions(active);

CREATE INDEX IF NOT EXISTS idx_attribute_values_active
ON attribute_values(active);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_subject_condition_set_selector_values;
DROP INDEX IF EXISTS idx_subject_mappings_attribute_value_id;
DROP INDEX IF EXISTS idx_subject_mappings_subject_condition_set_id;
DROP INDEX IF EXISTS idx_attribute_values_attribute_definition_id;
DROP INDEX IF EXISTS idx_attribute_fqns_value_id;
DROP INDEX IF EXISTS idx_subject_mapping_actions_mapping_action;
DROP INDEX IF EXISTS idx_attribute_namespaces_active;
DROP INDEX IF EXISTS idx_attribute_definitions_active;
DROP INDEX IF EXISTS idx_attribute_values_active;

ALTER TABLE subject_condition_set DROP COLUMN selector_values;

-- +goose StatementEnd
