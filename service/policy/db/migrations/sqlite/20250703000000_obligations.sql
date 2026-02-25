-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS obligation_definitions
(
    id UUID PRIMARY KEY,
    namespace_id UUID NOT NULL REFERENCES attribute_namespaces(id) ON DELETE CASCADE,
    -- name is a unique identifier for the obligation definition within the namespace
    name VARCHAR NOT NULL,
    metadata JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- implicit index on unique (namespace_id, name) combo
    -- index name: obligation_definitions_namespace_id_name_key
    UNIQUE (namespace_id, name)
);

CREATE TABLE IF NOT EXISTS obligation_values_standard
(
    id UUID PRIMARY KEY,
    obligation_definition_id UUID NOT NULL REFERENCES obligation_definitions(id) ON DELETE CASCADE,
    -- value is a unique identifier for the obligation value within the definition
    value VARCHAR NOT NULL,
    metadata JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- implicit index on unique (obligation_definition_id, value) combo
    -- index name: obligation_values_standard_obligation_definition_id_value_key
    UNIQUE (obligation_definition_id, value)
);

CREATE TABLE IF NOT EXISTS obligation_triggers
(
    id UUID PRIMARY KEY,
    obligation_value_id UUID NOT NULL REFERENCES obligation_values_standard(id) ON DELETE CASCADE,
    action_id UUID NOT NULL REFERENCES actions(id) ON DELETE CASCADE,
    attribute_value_id UUID NOT NULL REFERENCES attribute_values(id) ON DELETE CASCADE,
    metadata JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(obligation_value_id, action_id, attribute_value_id)
);

CREATE TABLE IF NOT EXISTS obligation_fulfillers
(
    id UUID PRIMARY KEY,
    obligation_value_id UUID NOT NULL REFERENCES obligation_values_standard(id) ON DELETE CASCADE,
    conditionals JSON,
    metadata JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- SQLite does not support plpgsql triggers/functions; updated_at is managed by application code.

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS obligation_fulfillers;
DROP TABLE IF EXISTS obligation_triggers;
DROP TABLE IF EXISTS obligation_values_standard;
DROP TABLE IF EXISTS obligation_definitions;

-- +goose StatementEnd
