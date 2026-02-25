-- +goose Up
-- +goose StatementBegin


CREATE TABLE IF NOT EXISTS attribute_namespaces
(
    -- generate on create
    id UUID PRIMARY KEY,
    name VARCHAR NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS attribute_definitions
(
    id UUID PRIMARY KEY,
    namespace_id UUID NOT NULL REFERENCES attribute_namespaces(id),
    name VARCHAR NOT NULL,
    rule attribute_definition_rule NOT NULL,
    metadata JSON,
    UNIQUE (namespace_id, name)
);

CREATE TABLE IF NOT EXISTS attribute_values
(
    id UUID PRIMARY KEY,
    attribute_definition_id UUID NOT NULL REFERENCES attribute_definitions(id),
    value VARCHAR NOT NULL,
    members TEXT,
    metadata JSON,
    UNIQUE (attribute_definition_id, value)
);

CREATE TABLE IF NOT EXISTS key_access_servers
(
    id UUID PRIMARY KEY,
    uri VARCHAR NOT NULL UNIQUE,
    public_key JSON NOT NULL,
    metadata JSON
);

CREATE TABLE IF NOT EXISTS attribute_definition_key_access_grants
(
    attribute_definition_id UUID NOT NULL REFERENCES attribute_definitions(id),
    key_access_server_id UUID NOT NULL REFERENCES key_access_servers(id),
    PRIMARY KEY (attribute_definition_id, key_access_server_id)
);

CREATE TABLE IF NOT EXISTS attribute_value_key_access_grants
(
    attribute_value_id UUID NOT NULL REFERENCES attribute_values(id),
    key_access_server_id UUID NOT NULL REFERENCES key_access_servers(id),
    PRIMARY KEY (attribute_value_id, key_access_server_id)
);

CREATE TABLE IF NOT EXISTS resource_mappings
(
    id UUID PRIMARY KEY,
    attribute_value_id UUID NOT NULL REFERENCES attribute_values(id),
    terms TEXT,
    metadata JSON
);

CREATE TABLE IF NOT EXISTS subject_mappings
(
    id UUID PRIMARY KEY,
    attribute_value_id UUID NOT NULL REFERENCES attribute_values(id),
    operator subject_mappings_operator NOT NULL,
    subject_attribute VARCHAR NOT NULL,
    subject_attribute_values TEXT,
    metadata JSON
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS subject_mappings;
DROP TABLE IF EXISTS resource_mappings;
DROP TABLE IF EXISTS attribute_value_key_access_grants;
DROP TABLE IF EXISTS attribute_definition_key_access_grants;
DROP TABLE IF EXISTS key_access_servers;
DROP TABLE IF EXISTS attribute_values;
DROP TABLE IF EXISTS attribute_definitions;
DROP TABLE IF EXISTS attribute_namespaces;

-- +goose StatementEnd
