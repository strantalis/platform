-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS attribute_namespace_key_access_grants
(
    namespace_id UUID NOT NULL REFERENCES attribute_namespaces(id) ON DELETE CASCADE,
    key_access_server_id UUID NOT NULL REFERENCES key_access_servers(id) ON DELETE CASCADE,
    PRIMARY KEY (namespace_id, key_access_server_id)
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS attribute_namespace_key_access_grants;

-- +goose StatementEnd
