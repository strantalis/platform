----------------------------------------------------------------
-- NAMESPACES
----------------------------------------------------------------

-- name: ListNamespaces :many
SELECT
    COUNT(*) OVER() AS total,
    ns.id,
    ns.name,
    ns.active,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(ns.metadata, '$.labels'), 'created_at', ns.created_at, 'updated_at', ns.updated_at)) as metadata,
    fqns.fqn
FROM attribute_namespaces ns
LEFT JOIN attribute_fqns fqns ON ns.id = fqns.namespace_id AND fqns.attribute_id IS NULL
WHERE (sqlc.narg('active') IS NULL OR ns.active = sqlc.narg('active'))
ORDER BY ns.created_at DESC
LIMIT sqlc.arg('limit_')
OFFSET sqlc.arg('offset_');

-- name: GetNamespace :one
SELECT
    ns.id,
    ns.name,
    ns.active,
    fqns.fqn,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(ns.metadata, '$.labels'), 'created_at', ns.created_at, 'updated_at', ns.updated_at)) as metadata,
    COALESCE(
        JSON_AGG(
            DISTINCT CASE
                WHEN kas_ns_grants.namespace_id IS NOT NULL THEN JSON_BUILD_OBJECT(
                    'id', kas.id,
                    'uri', kas.uri,
                    'name', kas.name,
                    'public_key', kas.public_key
                )
            END
        ),
        JSON_BUILD_ARRAY()
    ) as grants,
    COALESCE(nmp_keys.keys, JSON_BUILD_ARRAY()) as keys
FROM attribute_namespaces ns
LEFT JOIN attribute_namespace_key_access_grants kas_ns_grants ON kas_ns_grants.namespace_id = ns.id
LEFT JOIN key_access_servers kas ON kas.id = kas_ns_grants.key_access_server_id
LEFT JOIN attribute_fqns fqns ON fqns.namespace_id = ns.id
LEFT JOIN (
    SELECT
        k.namespace_id,
        JSON_AGG(
            DISTINCT CASE
                WHEN kask.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                    'kas_uri', kas.uri,
                    'kas_id', kas.id,
                    'public_key', JSON_BUILD_OBJECT(
                         'algorithm', kask.key_algorithm,
                         'kid', kask.key_id,
                         'pem', CONVERT_FROM(DECODE(json_extract(kask.public_key_ctx, '$.pem'), 'base64'), 'UTF8')
                    )
                )
            END
        ) AS keys
    FROM attribute_namespace_public_key_map k
    INNER JOIN key_access_server_keys kask ON k.key_access_server_key_id = kask.id
    INNER JOIN key_access_servers kas ON kask.key_access_server_id = kas.id
    GROUP BY k.namespace_id
) nmp_keys ON ns.id = nmp_keys.namespace_id
WHERE fqns.attribute_id IS NULL AND fqns.value_id IS NULL
  AND (sqlc.narg('id') IS NULL OR ns.id = sqlc.narg('id'))
  AND (sqlc.narg('name') IS NULL OR ns.name = REGEXP_REPLACE(sqlc.narg('name'), '^https://', ''))
GROUP BY ns.id, fqns.fqn, nmp_keys.keys;

-- name: CreateNamespace :one
INSERT INTO attribute_namespaces (id, name, metadata, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'), STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'))
RETURNING id;

-- updateNamespace: both Safe and Unsafe Updates
-- name: UpdateNamespace :execrows
UPDATE attribute_namespaces
SET
    name = COALESCE(sqlc.narg('name'), name),
    active = COALESCE(sqlc.narg('active'), active),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_at = STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE id = $1;

-- name: DeleteNamespace :execrows
DELETE FROM attribute_namespaces WHERE id = $1;

-- name: RemoveKeyAccessServerFromNamespace :execrows
DELETE FROM attribute_namespace_key_access_grants
WHERE namespace_id = $1 AND key_access_server_id = $2;

-- name: AssignPublicKeyToNamespace :one
INSERT INTO attribute_namespace_public_key_map (namespace_id, key_access_server_key_id)
VALUES ($1, $2)
RETURNING namespace_id, key_access_server_key_id;

-- name: RemovePublicKeyFromNamespace :execrows
DELETE FROM attribute_namespace_public_key_map
WHERE namespace_id = $1 AND key_access_server_key_id = $2;

-- name: RotatePublicKeyForNamespace :many
UPDATE attribute_namespace_public_key_map
SET key_access_server_key_id = sqlc.arg('new_key_id')
WHERE (key_access_server_key_id = sqlc.arg('old_key_id'))
RETURNING namespace_id;
