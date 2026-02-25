----------------------------------------------------------------
-- ACTIONS
----------------------------------------------------------------

-- name: ListActions :many
WITH counted AS (
    SELECT COUNT(id) AS total FROM actions
)
SELECT 
    a.id,
    a.name,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT(
        'labels', json_extract(a.metadata, '$.labels'), 
        'created_at', a.created_at, 
        'updated_at', a.updated_at
    )) as metadata,
    a.is_standard,
    counted.total
FROM actions a
CROSS JOIN counted
ORDER BY a.created_at DESC
LIMIT sqlc.arg('limit_') 
OFFSET sqlc.arg('offset_');

-- name: GetAction :one
SELECT 
    a.id,
    a.name,
    a.is_standard,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(a.metadata, '$.labels'), 'created_at', a.created_at, 'updated_at', a.updated_at)) AS metadata
FROM actions a
WHERE 
  (sqlc.narg('id') IS NULL OR a.id = sqlc.narg('id'))
  AND (sqlc.narg('name') IS NULL OR a.name = sqlc.narg('name'));

-- name: CreateOrListActionsByName :many
WITH input_actions AS (
    SELECT input.value AS name FROM json_each(sqlc.arg('action_names')) AS input(value)
),
new_actions AS (
    INSERT INTO actions (id, name, is_standard, created_at, updated_at)
    SELECT 
        gen_random_uuid(),
        input.name,
        FALSE, -- custom actions
        STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'),
        STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
    FROM input_actions input
    WHERE NOT EXISTS (
        SELECT 1 FROM actions a WHERE LOWER(a.name) = LOWER(input.name)
    )
    RETURNING id, name, is_standard, created_at
),
all_actions AS (
    -- Get existing actions that match input names
    SELECT a.id, a.name, a.is_standard, a.created_at, 
           TRUE AS pre_existing
    FROM actions a
    JOIN input_actions input ON LOWER(a.name) = LOWER(input.name)
    
    UNION ALL
    
    -- Include newly created actions
    SELECT id, name, is_standard, created_at,
           FALSE AS pre_existing
    FROM new_actions
)
SELECT 
    id,
    name,
    is_standard,
    created_at,
    pre_existing
FROM all_actions
ORDER BY name;

-- name: CreateCustomAction :one
INSERT INTO actions (id, name, metadata, is_standard, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    FALSE,
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'),
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
)
RETURNING id;

-- name: UpdateCustomAction :execrows
UPDATE actions
SET
    name = COALESCE(sqlc.narg('name'), name),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_at = STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now', '+0.001 seconds')
WHERE id = $1
  AND is_standard = FALSE;

-- name: DeleteCustomAction :execrows
DELETE FROM actions
WHERE id = $1
  AND is_standard = FALSE;
