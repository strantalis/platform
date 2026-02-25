---------------------------------------------------------------- 
-- ATTRIBUTE VALUES
----------------------------------------------------------------

-- name: ListAttributeValues :many
SELECT
    COUNT(*) OVER() AS total,
    av.id,
    av.value,
    av.active,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(av.metadata, '$.labels'), 'created_at', av.created_at, 'updated_at', av.updated_at)) as metadata,
    av.attribute_definition_id,
    fqns.fqn
FROM attribute_values av
LEFT JOIN attribute_fqns fqns ON av.id = fqns.value_id
WHERE (
    (sqlc.narg('active') IS NULL OR av.active = sqlc.narg('active')) AND
    (sqlc.narg('attribute_definition_id') IS NULL OR av.attribute_definition_id = sqlc.narg('attribute_definition_id')) 
)
ORDER BY av.created_at DESC
LIMIT sqlc.arg('limit_') 
OFFSET sqlc.arg('offset_'); 

-- name: GetAttributeValue :one
WITH obligation_triggers_agg AS (
    SELECT
        ot.obligation_value_id,
        JSON_AGG(
            DISTINCT JSON_BUILD_OBJECT(
                'id', ot.id,
                'action', JSON_BUILD_OBJECT(
                    'id', a.id,
                    'name', a.name
                ),
                'attribute_value', JSON_BUILD_OBJECT(
                    'id', av.id,
                    'fqn', av_fqns.fqn
                ),
                'context', CASE
                    WHEN ot.client_id IS NOT NULL THEN JSON_BUILD_ARRAY(
                        JSON_BUILD_OBJECT(
                            'pep', JSON_BUILD_OBJECT(
                                'client_id', ot.client_id
                            )
                        )
                    )
                    ELSE '[]'
                END
            )
        ) AS triggers
    FROM obligation_triggers ot
    JOIN actions a ON ot.action_id = a.id
    JOIN attribute_values av ON ot.attribute_value_id = av.id
    LEFT JOIN attribute_fqns av_fqns ON av.id = av_fqns.value_id
    GROUP BY ot.obligation_value_id
),
obligation_values_agg AS (
    SELECT
        ov.obligation_definition_id,
        JSON_AGG(
            DISTINCT JSON_BUILD_OBJECT(
                'id', ov.id,
                'value', ov.value,
                'fqn', ns_fqns.fqn || '/obl/' || od.name || '/value/' || ov.value,
                'triggers', COALESCE(ota.triggers, '[]')
            )
        ) AS "values"
    FROM obligation_values_standard ov
    LEFT JOIN obligation_triggers_agg ota ON ov.id = ota.obligation_value_id
    JOIN obligation_definitions od ON ov.obligation_definition_id = od.id
    JOIN attribute_namespaces n ON od.namespace_id = n.id
    LEFT JOIN attribute_fqns ns_fqns ON n.id = ns_fqns.namespace_id AND ns_fqns.attribute_id IS NULL AND ns_fqns.value_id IS NULL
    GROUP BY ov.obligation_definition_id
),
attribute_obligations AS (
    SELECT
        ot.attribute_value_id,
        JSON_AGG(
            DISTINCT JSON_BUILD_OBJECT(
                'id', od.id,
                'name', od.name,
                'fqn', ns_fqns.fqn || '/obl/' || od.name,
                'namespace', JSON_BUILD_OBJECT(
                    'id', n.id,
                    'name', n.name,
                    'fqn', ns_fqns.fqn
                ),
                'values', COALESCE(ova."values", '[]')
            )
        ) AS obligations
    FROM obligation_triggers ot
    JOIN obligation_values_standard ov ON ot.obligation_value_id = ov.id
    JOIN obligation_definitions od ON ov.obligation_definition_id = od.id
    JOIN attribute_namespaces n ON od.namespace_id = n.id
    LEFT JOIN attribute_fqns ns_fqns ON n.id = ns_fqns.namespace_id AND ns_fqns.attribute_id IS NULL AND ns_fqns.value_id IS NULL
    LEFT JOIN obligation_values_agg ova ON od.id = ova.obligation_definition_id
    GROUP BY ot.attribute_value_id
)
SELECT
    av.id,
    av.value,
    av.active,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(av.metadata, '$.labels'), 'created_at', av.created_at, 'updated_at', av.updated_at)) as metadata,
    av.attribute_definition_id,
    fqns.fqn,
    COALESCE(JSON_AGG(
        DISTINCT CASE
            WHEN avkag.attribute_value_id IS NOT NULL THEN JSON_BUILD_OBJECT(
                'id', kas.id,
                'uri', kas.uri,
                'name', kas.name,
                'public_key', kas.public_key
            )
        END
    ), JSON_BUILD_ARRAY()) AS grants,
    COALESCE(value_keys.keys, JSON_BUILD_ARRAY()) as keys,
    ao.obligations
FROM attribute_values av
LEFT JOIN attribute_fqns fqns ON av.id = fqns.value_id
LEFT JOIN attribute_value_key_access_grants avkag ON av.id = avkag.attribute_value_id
LEFT JOIN key_access_servers kas ON avkag.key_access_server_id = kas.id
LEFT JOIN (
    SELECT
        k.value_id,
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
    FROM attribute_value_public_key_map k
    INNER JOIN key_access_server_keys kask ON k.key_access_server_key_id = kask.id
    INNER JOIN key_access_servers kas ON kas.id = kask.key_access_server_id
    GROUP BY k.value_id
) value_keys ON av.id = value_keys.value_id
LEFT JOIN attribute_obligations ao ON av.id = ao.attribute_value_id
WHERE (sqlc.narg('id') IS NULL OR av.id = sqlc.narg('id'))
  AND (sqlc.narg('fqn') IS NULL OR REGEXP_REPLACE(fqns.fqn, '^https://', '') = REGEXP_REPLACE(sqlc.narg('fqn'), '^https://', ''))
GROUP BY av.id, fqns.fqn, value_keys.keys, ao.obligations;

-- name: CreateAttributeValue :one
INSERT INTO attribute_values (id, attribute_definition_id, value, metadata, created_at, updated_at)
VALUES (gen_random_uuid(), sqlc.arg('attribute_definition_id'), sqlc.arg('value'), sqlc.arg('metadata'), STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'), STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')) 
RETURNING id;

-- updateAttributeValue: Safe and Unsafe Updates both
-- name: UpdateAttributeValue :execrows
UPDATE attribute_values
SET
    value = COALESCE(sqlc.narg('value'), value),
    active = COALESCE(sqlc.narg('active'), active),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_at = STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE id = $1;

-- name: DeleteAttributeValue :execrows
DELETE FROM attribute_values WHERE id = $1;

-- name: RemoveKeyAccessServerFromAttributeValue :execrows
DELETE FROM attribute_value_key_access_grants
WHERE attribute_value_id = $1 AND key_access_server_id = $2;

-- name: AssignPublicKeyToAttributeValue :one
INSERT INTO attribute_value_public_key_map (value_id, key_access_server_key_id)
VALUES ($1, $2)
RETURNING *;

-- name: RemovePublicKeyFromAttributeValue :execrows
DELETE FROM attribute_value_public_key_map
WHERE value_id = $1 AND key_access_server_key_id = $2;

-- name: RotatePublicKeyForAttributeValue :many
UPDATE attribute_value_public_key_map
SET key_access_server_key_id = sqlc.arg('new_key_id')
WHERE (key_access_server_key_id = sqlc.arg('old_key_id'))
RETURNING value_id;
