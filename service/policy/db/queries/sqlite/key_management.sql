---------------------------------------------------------------- 
-- Provider Config
----------------------------------------------------------------

-- name: CreateProviderConfig :one
WITH inserted AS (
  INSERT INTO provider_config (id, provider_name, manager, config, metadata, created_at, updated_at)
  VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    $4,
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'),
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
  )
  RETURNING *
)
SELECT 
  id,
  provider_name,
  manager,
  config,
  JSON_STRIP_NULLS(
    JSON_BUILD_OBJECT(
      'labels', json_extract(metadata, '$.labels'),         
      'created_at', created_at,               
      'updated_at', updated_at                
    )
  ) AS metadata
FROM inserted;

-- name: GetProviderConfig :one
SELECT 
    pc.id,
    pc.provider_name,
    pc.manager,
    pc.config,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(pc.metadata, '$.labels'), 'created_at', pc.created_at, 'updated_at', pc.updated_at)) AS metadata
FROM provider_config AS pc
WHERE (sqlc.narg('id') IS NULL OR pc.id = sqlc.narg('id'))
  AND (sqlc.narg('name') IS NULL OR pc.provider_name = sqlc.narg('name'))
  AND (sqlc.narg('manager') IS NULL OR pc.manager = sqlc.narg('manager'));


-- name: ListProviderConfigs :many
WITH counted AS (
    SELECT COUNT(pc.id) AS total 
    FROM provider_config pc
)
SELECT 
    pc.id,
    pc.provider_name,
    pc.manager,
    pc.config,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(pc.metadata, '$.labels'), 'created_at', pc.created_at, 'updated_at', pc.updated_at)) AS metadata,
    counted.total
FROM provider_config AS pc
CROSS JOIN counted
ORDER BY pc.created_at DESC
LIMIT sqlc.arg('limit_') 
OFFSET sqlc.arg('offset_');

-- name: UpdateProviderConfig :execrows
UPDATE provider_config
SET
    provider_name = COALESCE(sqlc.narg('provider_name'), provider_name),
    manager = COALESCE(sqlc.narg('manager'), manager),
    config = COALESCE(sqlc.narg('config'), config),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_at = STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now', '+0.001 seconds')
WHERE id = $1;

-- name: DeleteProviderConfig :execrows
DELETE FROM provider_config 
WHERE id = $1;
