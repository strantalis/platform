----------------------------------------------------------------
-- REGISTERED RESOURCES
----------------------------------------------------------------

-- name: CreateRegisteredResource :one
INSERT INTO registered_resources (id, name, metadata, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'),
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
)
RETURNING id;

-- name: GetRegisteredResource :one
SELECT
    r.id,
    r.name,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(r.metadata, '$.labels'), 'created_at', r.created_at, 'updated_at', r.updated_at)) as metadata,
    JSON_AGG(
        CASE
            WHEN v.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                'id', v.id,
                'value', v.value
            )
        END
    ) as values
FROM registered_resources r
LEFT JOIN registered_resource_values v ON v.registered_resource_id = r.id
WHERE
    (sqlc.narg('id') IS NULL OR r.id = sqlc.narg('id')) AND
    (sqlc.narg('name') IS NULL OR r.name = sqlc.narg('name'))
GROUP BY r.id;

-- name: ListRegisteredResources :many
WITH counted AS (
    SELECT COUNT(id) AS total
    FROM registered_resources
)
SELECT
    r.id,
    r.name,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(r.metadata, '$.labels'), 'created_at', r.created_at, 'updated_at', r.updated_at)) as metadata,
    -- Aggregate all values for this resource into a JSON array, filtering NULL entries
    JSON_AGG(
        CASE
            WHEN v.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                'id', v.id,
                'value', v.value,
                'action_attribute_values', action_attrs.values
            )
        END
    ) as values,
    counted.total
FROM registered_resources r
CROSS JOIN counted
LEFT JOIN registered_resource_values v ON v.registered_resource_id = r.id
LEFT JOIN (
    SELECT
        rav.registered_resource_value_id,
        JSON_AGG(
            JSON_BUILD_OBJECT(
                'action', JSON_BUILD_OBJECT(
                    'id', a.id,
                    'name', a.name
                ),
                'attribute_value', JSON_BUILD_OBJECT(
                    'id', av.id,
                    'value', av.value,
                    'fqn', fqns.fqn
                )
            )
        ) AS values
    FROM registered_resource_action_attribute_values rav
    LEFT JOIN actions a on rav.action_id = a.id
    LEFT JOIN attribute_values av on rav.attribute_value_id = av.id
    LEFT JOIN attribute_fqns fqns on av.id = fqns.value_id
    GROUP BY rav.registered_resource_value_id
) action_attrs ON action_attrs.registered_resource_value_id = v.id
GROUP BY r.id, counted.total
ORDER BY r.created_at DESC
LIMIT sqlc.arg('limit_') 
OFFSET sqlc.arg('offset_');

-- name: UpdateRegisteredResource :execrows
UPDATE registered_resources
SET
    name = COALESCE(sqlc.narg('name'), name),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_at = STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now', '+0.001 seconds')
WHERE id = $1;

-- name: DeleteRegisteredResource :execrows
DELETE FROM registered_resources WHERE id = $1;


----------------------------------------------------------------
-- REGISTERED RESOURCE VALUES
----------------------------------------------------------------

-- name: CreateRegisteredResourceValue :one
INSERT INTO registered_resource_values (id, registered_resource_id, value, metadata, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'),
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
)
RETURNING id;

-- name: GetRegisteredResourceValue :one
SELECT
    v.id,
    v.registered_resource_id,
    v.value,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(v.metadata, '$.labels'), 'created_at', v.created_at, 'updated_at', v.updated_at)) as metadata,
    JSON_AGG(
        CASE
            WHEN rav.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                'action', JSON_BUILD_OBJECT(
                    'id', a.id,
                    'name', a.name
                ),
                'attribute_value', JSON_BUILD_OBJECT(
                    'id', av.id,
                    'value', av.value,
                    'fqn', fqns.fqn
                )
            )
        END
    ) as action_attribute_values
FROM registered_resource_values v
JOIN registered_resources r ON v.registered_resource_id = r.id
LEFT JOIN registered_resource_action_attribute_values rav ON v.id = rav.registered_resource_value_id
LEFT JOIN actions a on rav.action_id = a.id
LEFT JOIN attribute_values av on rav.attribute_value_id = av.id
LEFT JOIN attribute_fqns fqns on av.id = fqns.value_id
WHERE
    (sqlc.narg('id') IS NULL OR v.id = sqlc.narg('id')) AND
    (sqlc.narg('name') IS NULL OR r.name = sqlc.narg('name')) AND
    (sqlc.narg('value') IS NULL OR v.value = sqlc.narg('value'))
GROUP BY v.id;

-- name: ListRegisteredResourceValues :many
WITH counted AS (
    SELECT COUNT(v.id) AS total
    FROM registered_resource_values v
    WHERE sqlc.narg('registered_resource_id') IS NULL OR v.registered_resource_id = sqlc.narg('registered_resource_id')
)
SELECT
    v.id,
    v.registered_resource_id,
    v.value,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(v.metadata, '$.labels'), 'created_at', v.created_at, 'updated_at', v.updated_at)) as metadata,
    JSON_AGG(
        CASE
            WHEN rav.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                'action', JSON_BUILD_OBJECT(
                    'id', a.id,
                    'name', a.name
                ),
                'attribute_value', JSON_BUILD_OBJECT(
                    'id', av.id,
                    'value', av.value,
                    'fqn', fqns.fqn
                )
            )
        END
    ) as action_attribute_values,
    counted.total
FROM registered_resource_values v
JOIN registered_resources r ON v.registered_resource_id = r.id
LEFT JOIN registered_resource_action_attribute_values rav ON v.id = rav.registered_resource_value_id
LEFT JOIN actions a on rav.action_id = a.id
LEFT JOIN attribute_values av on rav.attribute_value_id = av.id
LEFT JOIN attribute_fqns fqns on av.id = fqns.value_id  
CROSS JOIN counted
WHERE
    sqlc.narg('registered_resource_id') IS NULL OR v.registered_resource_id = sqlc.narg('registered_resource_id')
GROUP BY v.id, counted.total
ORDER BY v.created_at DESC
LIMIT sqlc.arg('limit_')
OFFSET sqlc.arg('offset_');

-- name: UpdateRegisteredResourceValue :execrows
UPDATE registered_resource_values
SET
    value = COALESCE(sqlc.narg('value'), value),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_at = STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now', '+0.001 seconds')
WHERE id = $1;

-- name: DeleteRegisteredResourceValue :execrows
DELETE FROM registered_resource_values WHERE id = $1;

---------------------------------------------------------------- 
-- Registered Resource Action Attribute Values
----------------------------------------------------------------

-- name: CreateRegisteredResourceActionAttributeValues :exec
INSERT INTO registered_resource_action_attribute_values (
    id,
    registered_resource_value_id,
    action_id,
    attribute_value_id,
    created_at,
    updated_at
)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'),
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
);

-- name: DeleteRegisteredResourceActionAttributeValues :execrows
DELETE FROM registered_resource_action_attribute_values
WHERE registered_resource_value_id = $1;
