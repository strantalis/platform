-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS actions (
    id UUID PRIMARY KEY,
    name VARCHAR NOT NULL,
    is_standard BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT actions_name_unique UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS subject_mapping_actions (
    subject_mapping_id UUID NOT NULL REFERENCES subject_mappings(id) ON DELETE CASCADE,
    action_id UUID NOT NULL REFERENCES actions(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (subject_mapping_id, action_id)
);

CREATE INDEX IF NOT EXISTS idx_subject_mapping_actions_subject ON subject_mapping_actions(subject_mapping_id);
CREATE INDEX IF NOT EXISTS idx_subject_mapping_actions_action ON subject_mapping_actions(action_id);

-- Insert standard actions.
INSERT OR IGNORE INTO actions (id, name, is_standard) VALUES
    ('00000000-0000-0000-0000-000000000001', 'create', TRUE),
    ('00000000-0000-0000-0000-000000000002', 'read', TRUE),
    ('00000000-0000-0000-0000-000000000003', 'update', TRUE),
    ('00000000-0000-0000-0000-000000000004', 'delete', TRUE);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS subject_mapping_actions;
DROP TABLE IF EXISTS actions;

-- +goose StatementEnd
