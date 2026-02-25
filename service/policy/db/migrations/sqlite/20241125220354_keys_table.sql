-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS public_keys (
    id uuid PRIMARY KEY,
    is_active boolean NOT NULL DEFAULT FALSE,
    was_mapped boolean NOT NULL DEFAULT FALSE,
    key_access_server_id uuid NOT NULL REFERENCES key_access_servers(id),
    key_id varchar(36) NOT NULL,
    alg varchar(50) NOT NULL,
    public_key text NOT NULL,
    metadata JSON,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (key_access_server_id, key_id, alg)
);

CREATE TABLE IF NOT EXISTS attribute_namespace_public_key_map (
    namespace_id uuid NOT NULL REFERENCES attribute_namespaces(id) ON DELETE CASCADE,
    key_id uuid NOT NULL REFERENCES public_keys(id) ON DELETE CASCADE,
    PRIMARY KEY (namespace_id, key_id)
);

CREATE TABLE IF NOT EXISTS attribute_definition_public_key_map (
    definition_id uuid NOT NULL REFERENCES attribute_definitions(id) ON DELETE CASCADE,
    key_id uuid NOT NULL REFERENCES public_keys(id) ON DELETE CASCADE,
    PRIMARY KEY (definition_id, key_id)
);

CREATE TABLE IF NOT EXISTS attribute_value_public_key_map (
    value_id uuid NOT NULL REFERENCES attribute_values(id) ON DELETE CASCADE,
    key_id uuid NOT NULL REFERENCES public_keys(id) ON DELETE CASCADE,
    PRIMARY KEY (value_id, key_id)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS attribute_value_public_key_map;
DROP TABLE IF EXISTS attribute_definition_public_key_map;
DROP TABLE IF EXISTS attribute_namespace_public_key_map;
DROP TABLE IF EXISTS public_keys;

-- +goose StatementEnd
