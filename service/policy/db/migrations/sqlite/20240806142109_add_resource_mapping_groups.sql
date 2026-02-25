-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS resource_mapping_groups (
    id UUID PRIMARY KEY,
    namespace_id UUID NOT NULL REFERENCES attribute_namespaces(id) ON DELETE CASCADE,
    name VARCHAR NOT NULL,
    UNIQUE(namespace_id, name)
);


ALTER TABLE resource_mappings ADD COLUMN group_id UUID REFERENCES resource_mapping_groups(id) ON DELETE SET NULL;


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE resource_mappings DROP COLUMN group_id;

DROP TABLE IF EXISTS resource_mapping_groups;

-- +goose StatementEnd
