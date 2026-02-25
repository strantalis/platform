-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS resources
(
    id INTEGER PRIMARY KEY,
    name VARCHAR NOT NULL,
    namespace VARCHAR NOT NULL,
    version INTEGER NOT NULL,
    fqn VARCHAR,
    labels JSON,
    description VARCHAR,
    policytype VARCHAR NOT NULL,
    resource JSON
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS resources;
-- +goose StatementEnd
