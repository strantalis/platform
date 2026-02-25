CREATE TABLE IF NOT EXISTS attribute_namespaces (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    metadata TEXT,
    active BOOLEAN,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS attribute_definitions (
    id TEXT PRIMARY KEY,
    namespace_id TEXT,
    name TEXT NOT NULL,
    rule TEXT,
    metadata TEXT,
    active BOOLEAN,
    values_order TEXT,
    allow_traversal BOOLEAN,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS attribute_values (
    id TEXT PRIMARY KEY,
    attribute_definition_id TEXT,
    value TEXT NOT NULL,
    metadata TEXT,
    active BOOLEAN,
    members TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS attribute_fqns (
    id TEXT PRIMARY KEY,
    fqn TEXT NOT NULL,
    namespace_id TEXT,
    attribute_id TEXT,
    value_id TEXT
);

CREATE TABLE IF NOT EXISTS key_access_servers (
    id TEXT PRIMARY KEY,
    uri TEXT NOT NULL,
    public_key TEXT,
    metadata TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    name TEXT,
    source_type TEXT
);

CREATE TABLE IF NOT EXISTS attribute_definition_key_access_grants (
    attribute_definition_id TEXT,
    key_access_server_id TEXT
);

CREATE TABLE IF NOT EXISTS attribute_value_key_access_grants (
    attribute_value_id TEXT,
    key_access_server_id TEXT
);

CREATE TABLE IF NOT EXISTS attribute_namespace_key_access_grants (
    namespace_id TEXT,
    key_access_server_id TEXT
);

CREATE TABLE IF NOT EXISTS resource_mapping_groups (
    id TEXT PRIMARY KEY,
    namespace_id TEXT,
    name TEXT,
    metadata TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS resource_mappings (
    id TEXT PRIMARY KEY,
    attribute_value_id TEXT,
    terms TEXT,
    metadata TEXT,
    group_id TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS subject_condition_set (
    id TEXT PRIMARY KEY,
    condition TEXT,
    metadata TEXT,
    selector_values TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS subject_mappings (
    id TEXT PRIMARY KEY,
    attribute_value_id TEXT,
    subject_condition_set_id TEXT,
    metadata TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS actions (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    is_standard BOOLEAN,
    metadata TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS subject_mapping_actions (
    subject_mapping_id TEXT,
    action_id TEXT,
    created_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS registered_resources (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    metadata TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS registered_resource_values (
    id TEXT PRIMARY KEY,
    registered_resource_id TEXT,
    value TEXT NOT NULL,
    metadata TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS registered_resource_action_attribute_values (
    id TEXT PRIMARY KEY,
    registered_resource_value_id TEXT,
    action_id TEXT,
    attribute_value_id TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS provider_config (
    id TEXT PRIMARY KEY,
    provider_name TEXT NOT NULL,
    manager TEXT,
    config TEXT,
    metadata TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS key_access_server_keys (
    id TEXT PRIMARY KEY,
    key_access_server_id TEXT,
    key_algorithm INTEGER,
    key_id TEXT,
    key_status INTEGER,
    key_mode INTEGER,
    public_key_ctx TEXT,
    private_key_ctx TEXT,
    expiration TIMESTAMP,
    provider_config_id TEXT,
    metadata TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    legacy BOOLEAN
);

CREATE TABLE IF NOT EXISTS attribute_namespace_public_key_map (
    namespace_id TEXT,
    key_access_server_key_id TEXT
);

CREATE TABLE IF NOT EXISTS attribute_definition_public_key_map (
    definition_id TEXT,
    key_access_server_key_id TEXT
);

CREATE TABLE IF NOT EXISTS attribute_value_public_key_map (
    value_id TEXT,
    key_access_server_key_id TEXT
);

CREATE TABLE IF NOT EXISTS obligation_definitions (
    id TEXT PRIMARY KEY,
    namespace_id TEXT,
    name TEXT NOT NULL,
    metadata TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS obligation_values_standard (
    id TEXT PRIMARY KEY,
    obligation_definition_id TEXT,
    value TEXT NOT NULL,
    metadata TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS obligation_triggers (
    id TEXT PRIMARY KEY,
    obligation_value_id TEXT,
    action_id TEXT,
    attribute_value_id TEXT,
    client_id TEXT,
    metadata TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS obligation_fulfillers (
    id TEXT PRIMARY KEY,
    obligation_value_id TEXT,
    conditionals TEXT,
    metadata TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS base_keys (
    id TEXT PRIMARY KEY,
    key_access_server_key_id TEXT
);
