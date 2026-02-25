-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS registered_resource_action_attribute_values (
  id UUID PRIMARY KEY,
  registered_resource_value_id UUID NOT NULL REFERENCES registered_resource_values(id) ON DELETE CASCADE,
  action_id UUID NOT NULL REFERENCES actions(id) ON DELETE CASCADE,
  attribute_value_id UUID NOT NULL REFERENCES attribute_values(id) ON DELETE CASCADE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(registered_resource_value_id, action_id, attribute_value_id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS registered_resource_action_attribute_values;

-- +goose StatementEnd
