-- +goose Up
-- +goose StatementBegin

DROP TABLE IF EXISTS attribute_namespace_certificates;
DROP TABLE IF EXISTS certificates;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS certificates
(
    id UUID PRIMARY KEY,
    pem TEXT NOT NULL,
    metadata JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS attribute_namespace_certificates
(
    namespace_id UUID NOT NULL REFERENCES attribute_namespaces(id) ON DELETE CASCADE,
    certificate_id UUID NOT NULL REFERENCES certificates(id) ON DELETE CASCADE,
    PRIMARY KEY (namespace_id, certificate_id)
);


-- SQLite does not support plpgsql triggers/functions; updated_at is managed by application code.

-- +goose StatementEnd
