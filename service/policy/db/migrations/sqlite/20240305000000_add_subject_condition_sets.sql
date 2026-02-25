-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS subject_condition_set (
    id UUID PRIMARY KEY,
    condition JSON NOT NULL,
    metadata JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE subject_mappings ADD COLUMN subject_condition_set_id UUID REFERENCES subject_condition_set(id) ON DELETE CASCADE;

-- Remove legacy subject mapping columns (migrated to subject_condition_set).
ALTER TABLE subject_mappings DROP COLUMN operator;
ALTER TABLE subject_mappings DROP COLUMN subject_attribute;
ALTER TABLE subject_mappings DROP COLUMN subject_attribute_values;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore legacy columns (nullable) before removing condition set.
ALTER TABLE subject_mappings ADD COLUMN operator TEXT;
ALTER TABLE subject_mappings ADD COLUMN subject_attribute TEXT;
ALTER TABLE subject_mappings ADD COLUMN subject_attribute_values TEXT;

ALTER TABLE subject_mappings DROP COLUMN subject_condition_set_id;
DROP TABLE IF EXISTS subject_condition_set;

-- +goose StatementEnd
