-- +goose Up
-- +goose StatementBegin

ALTER TABLE key_access_servers
ADD COLUMN source_type VARCHAR;

-- Replace legacy public_keys tables with the new key_access_server_keys model.
DROP TABLE IF EXISTS attribute_value_public_key_map;
DROP TABLE IF EXISTS attribute_definition_public_key_map;
DROP TABLE IF EXISTS attribute_namespace_public_key_map;
DROP TABLE IF EXISTS public_keys;

CREATE TABLE IF NOT EXISTS provider_config (
    id UUID CONSTRAINT provider_config_pkey PRIMARY KEY,
    provider_name VARCHAR(255) NOT NULL,
    config JSON NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSON
);

CREATE TABLE IF NOT EXISTS asym_key (
    id UUID CONSTRAINT asym_key_pkey PRIMARY KEY,
    key_id VARCHAR(36) NOT NULL UNIQUE,
    key_algorithm INT NOT NULL,
    key_status INT NOT NULL,
    key_mode INT NOT NULL,
    public_key_ctx JSON,
    private_key_ctx JSON,
    expiration TIMESTAMP,
    provider_config_id UUID CONSTRAINT asym_key_provider_config_fk REFERENCES provider_config(id),
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sym_key (
    id UUID CONSTRAINT sym_key_pkey PRIMARY KEY,
    key_id VARCHAR(36) NOT NULL UNIQUE,
    key_status INT NOT NULL,
    key_mode INT NOT NULL,
    key_value BLOB,
    provider_config_id UUID CONSTRAINT sym_key_provider_config_fk REFERENCES provider_config(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    metadata JSON
);

CREATE TABLE IF NOT EXISTS key_access_server_keys (
    id UUID CONSTRAINT key_access_server_keys_pkey PRIMARY KEY,
    key_access_server_id UUID NOT NULL CONSTRAINT key_access_server_fk REFERENCES key_access_servers(id) ON DELETE CASCADE,
    key_algorithm INT NOT NULL,
    key_id VARCHAR(36) NOT NULL,
    key_status INT NOT NULL,
    key_mode INT NOT NULL,
    public_key_ctx JSON,
    private_key_ctx JSON,
    expiration TIMESTAMP,
    provider_config_id UUID CONSTRAINT key_access_server_keys_provider_config_fk REFERENCES provider_config(id),
    metadata JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (key_access_server_id, key_id)
);

CREATE TABLE IF NOT EXISTS attribute_namespace_public_key_map (
    namespace_id UUID NOT NULL CONSTRAINT namespace_fk REFERENCES attribute_namespaces(id) ON DELETE CASCADE,
    key_access_server_key_id UUID NOT NULL CONSTRAINT key_access_server_keys_fk REFERENCES key_access_server_keys(id) ON DELETE CASCADE,
    PRIMARY KEY (namespace_id, key_access_server_key_id)
);

CREATE TABLE IF NOT EXISTS attribute_definition_public_key_map (
    definition_id UUID NOT NULL CONSTRAINT definition_fk REFERENCES attribute_definitions(id) ON DELETE CASCADE,
    key_access_server_key_id UUID NOT NULL CONSTRAINT key_access_server_keys_fk REFERENCES key_access_server_keys(id) ON DELETE CASCADE,
    PRIMARY KEY (definition_id, key_access_server_key_id)
);

CREATE TABLE IF NOT EXISTS attribute_value_public_key_map (
    value_id UUID NOT NULL CONSTRAINT value_fk REFERENCES attribute_values(id) ON DELETE CASCADE,
    key_access_server_key_id UUID NOT NULL CONSTRAINT key_access_server_keys_fk REFERENCES key_access_server_keys(id) ON DELETE CASCADE,
    PRIMARY KEY (value_id, key_access_server_key_id)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS attribute_value_public_key_map;
DROP TABLE IF EXISTS attribute_definition_public_key_map;
DROP TABLE IF EXISTS attribute_namespace_public_key_map;
DROP TABLE IF EXISTS key_access_server_keys;
DROP TABLE IF EXISTS sym_key;
DROP TABLE IF EXISTS asym_key;
DROP TABLE IF EXISTS provider_config;

ALTER TABLE key_access_servers
DROP COLUMN source_type;

-- +goose StatementEnd
