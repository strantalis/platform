---------------------------------------------------------------- 
-- SUBJECT CONDITION SETS
----------------------------------------------------------------

-- name: ListSubjectConditionSets :many
WITH counted AS (
    SELECT COUNT(scs.id) AS total
    FROM subject_condition_set scs
)
SELECT
    scs.id,
    scs.condition,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(scs.metadata, '$.labels'), 'created_at', scs.created_at, 'updated_at', scs.updated_at)) as metadata,
    counted.total
FROM subject_condition_set scs
CROSS JOIN counted
ORDER BY scs.created_at DESC
LIMIT sqlc.arg('limit_') 
OFFSET sqlc.arg('offset_'); 

-- name: GetSubjectConditionSet :one
SELECT
    id,
    condition,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(metadata, '$.labels'), 'created_at', created_at, 'updated_at', updated_at)) as metadata
FROM subject_condition_set
WHERE id = $1;

-- name: CreateSubjectConditionSet :one
INSERT INTO subject_condition_set (id, condition, metadata, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'),
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
)
RETURNING id;

-- name: UpdateSubjectConditionSet :execrows
UPDATE subject_condition_set
SET
    condition = COALESCE(sqlc.narg('condition'), condition),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_at = STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now', '+0.001 seconds')
WHERE id = $1;

-- name: DeleteSubjectConditionSet :execrows
DELETE FROM subject_condition_set WHERE id = $1;

-- name: DeleteAllUnmappedSubjectConditionSets :many
DELETE FROM subject_condition_set
WHERE id NOT IN (SELECT DISTINCT sm.subject_condition_set_id FROM subject_mappings sm)
RETURNING id;

---------------------------------------------------------------- 
-- SUBJECT MAPPINGS
----------------------------------------------------------------

-- name: ListSubjectMappings :many
WITH subject_actions AS (
    SELECT
        sma.subject_mapping_id,
        COALESCE(
            JSON_AGG(CASE WHEN a.is_standard = TRUE THEN JSON_BUILD_OBJECT('id', a.id, 'name', a.name) END),
            '[]'
        ) AS standard_actions,
        COALESCE(
            JSON_AGG(CASE WHEN a.is_standard = FALSE THEN JSON_BUILD_OBJECT('id', a.id, 'name', a.name) END),
            '[]'
        ) AS custom_actions
    FROM subject_mapping_actions sma
    JOIN actions a ON sma.action_id = a.id
    GROUP BY sma.subject_mapping_id
), counted AS (
    SELECT COUNT(sm.id) AS total
    FROM subject_mappings sm
)
SELECT
    sm.id,
    sa.standard_actions,
    sa.custom_actions,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(sm.metadata, '$.labels'), 'created_at', sm.created_at, 'updated_at', sm.updated_at)) AS metadata,
    JSON_BUILD_OBJECT(
        'id', scs.id,
        'metadata', JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(scs.metadata, '$.labels'), 'created_at', scs.created_at, 'updated_at', scs.updated_at)),
        'subject_sets', scs.condition
    ) AS subject_condition_set,
    JSON_BUILD_OBJECT(
        'id', av.id,
        'value', av.value,
        'active', av.active,
        'fqn', fqns.fqn
    ) AS attribute_value,
    counted.total
FROM subject_mappings sm
CROSS JOIN counted
LEFT JOIN subject_actions sa ON sm.id = sa.subject_mapping_id
LEFT JOIN attribute_values av ON sm.attribute_value_id = av.id
LEFT JOIN attribute_fqns fqns ON av.id = fqns.value_id
LEFT JOIN subject_condition_set scs ON scs.id = sm.subject_condition_set_id
GROUP BY
    sm.id,
    sa.standard_actions,
    sa.custom_actions,
    sm.metadata, sm.created_at, sm.updated_at, -- for metadata object
    scs.id, scs.metadata, scs.created_at, scs.updated_at, scs.condition, -- for subject_condition_set object
    av.id, av.value, av.active, -- for attribute_value object
    fqns.fqn,
    counted.total
ORDER BY sm.created_at DESC
LIMIT sqlc.arg('limit_')
OFFSET sqlc.arg('offset_');

-- name: GetSubjectMapping :one
SELECT
    sm.id,
    (
        SELECT JSON_AGG(JSON_BUILD_OBJECT('id', a.id, 'name', a.name))
        FROM actions a
        JOIN subject_mapping_actions sma ON sma.action_id = a.id
        WHERE sma.subject_mapping_id = sm.id AND a.is_standard = TRUE
    ) AS standard_actions,
    (
        SELECT JSON_AGG(JSON_BUILD_OBJECT('id', a.id, 'name', a.name))
        FROM actions a
        JOIN subject_mapping_actions sma ON sma.action_id = a.id
        WHERE sma.subject_mapping_id = sm.id AND a.is_standard = FALSE
    ) AS custom_actions,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(sm.metadata, '$.labels'), 'created_at', sm.created_at, 'updated_at', sm.updated_at)) AS metadata,
    JSON_BUILD_OBJECT(
        'id', scs.id,
        'metadata', JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(scs.metadata, '$.labels'), 'created_at', scs.created_at, 'updated_at', scs.updated_at)),
        'subject_sets', scs.condition
    ) AS subject_condition_set,
    JSON_BUILD_OBJECT('id', av.id,'value', av.value,'active', av.active) AS attribute_value
FROM subject_mappings sm
LEFT JOIN attribute_values av ON sm.attribute_value_id = av.id
LEFT JOIN subject_condition_set scs ON scs.id = sm.subject_condition_set_id
WHERE sm.id = $1
GROUP BY av.id, sm.id, scs.id;

-- name: MatchSubjectMappings :many
WITH subject_actions AS (
    SELECT
        sma.subject_mapping_id,
        COALESCE(
            JSON_AGG(CASE WHEN a.is_standard = TRUE THEN JSON_BUILD_OBJECT('id', a.id, 'name', a.name) END),
            '[]'
        ) AS standard_actions,
        COALESCE(
            JSON_AGG(CASE WHEN a.is_standard = FALSE THEN JSON_BUILD_OBJECT('id', a.id, 'name', a.name) END),
            '[]'
        ) AS custom_actions
    FROM subject_mapping_actions sma
    JOIN actions a ON sma.action_id = a.id
    GROUP BY sma.subject_mapping_id
)
SELECT
    sm.id,
    sa.standard_actions,
    sa.custom_actions,
    JSON_BUILD_OBJECT(
        'id', scs.id,
        'subject_sets', scs.condition
    ) AS subject_condition_set,
    JSON_BUILD_OBJECT(
        'id', av.id,
        'value', av.value,
        'active', av.active,
        'fqn', fqns.fqn
    ) AS attribute_value
FROM subject_mappings sm
LEFT JOIN subject_actions sa ON sm.id = sa.subject_mapping_id
LEFT JOIN attribute_values av ON sm.attribute_value_id = av.id
LEFT JOIN attribute_definitions ad ON av.attribute_definition_id = ad.id
LEFT JOIN attribute_namespaces ns ON ad.namespace_id = ns.id
LEFT JOIN attribute_fqns fqns ON av.id = fqns.value_id
LEFT JOIN subject_condition_set scs ON scs.id = sm.subject_condition_set_id
WHERE
    ns.active = TRUE
    AND ad.active = TRUE
    AND av.active = TRUE
    AND EXISTS (
        SELECT 1
        FROM json_each(COALESCE(json_extract(scs.condition, '$.subject_sets'), scs.condition)) AS ss(value)
        CROSS JOIN json_each(json_extract(ss.value, '$.condition_groups')) AS cg(value)
        CROSS JOIN json_each(json_extract(cg.value, '$.conditions')) AS cond(value)
        WHERE json_extract(cond.value, '$.subject_external_selector_value') IN (
            SELECT selectors.value FROM json_each(sqlc.arg('selectors')) AS selectors(value)
        )
    )
GROUP BY
    sm.id,
    sa.standard_actions,
    sa.custom_actions,
    scs.id, scs.condition,
    av.id, av.value, av.active, fqns.fqn;

-- name: CreateSubjectMapping :one
WITH inserted_mapping AS (
    INSERT INTO subject_mappings (
        id,
        attribute_value_id,
        metadata,
        subject_condition_set_id,
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
    )
    RETURNING id
),
inserted_actions AS (
    INSERT INTO subject_mapping_actions (subject_mapping_id, action_id)
    SELECT 
        (SELECT id FROM inserted_mapping),
        action_ids.value
    FROM json_each(sqlc.arg('action_ids')) AS action_ids(value)
    RETURNING action_id
)
SELECT id FROM inserted_mapping;

-- name: UpdateSubjectMapping :execrows
WITH
    subject_mapping_update AS (
        UPDATE subject_mappings
        SET
            metadata = COALESCE(sqlc.narg('metadata'), metadata),
            subject_condition_set_id = COALESCE(sqlc.narg('subject_condition_set_id'), subject_condition_set_id),
            updated_at = STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now', '+0.001 seconds')
        WHERE id = sqlc.arg('id')
        RETURNING id
    ),
    -- Delete any actions that are NOT in the new list
    action_delete AS (
        DELETE FROM subject_mapping_actions
        WHERE
            subject_mapping_id = sqlc.arg('id')
            AND sqlc.narg('action_ids') IS NOT NULL
            AND action_id NOT IN (
                SELECT action_ids.value FROM json_each(sqlc.narg('action_ids')) AS action_ids(value)
            )
        RETURNING action_id
    ),
    -- Insert actions that are not already related to the mapping
    action_insert AS (
        INSERT INTO
            subject_mapping_actions (subject_mapping_id, action_id)
        SELECT
            sqlc.arg('id'),
            a.value
        FROM json_each(sqlc.narg('action_ids')) AS a(value)
        WHERE
            sqlc.narg('action_ids') IS NOT NULL
            AND NOT EXISTS (
                SELECT 1
                FROM subject_mapping_actions
                WHERE subject_mapping_id = sqlc.arg('id') AND action_id = a.value
            )
        RETURNING action_id
    ),
    update_count AS (
        SELECT COUNT(*) AS cnt
        FROM subject_mapping_update
    )
SELECT cnt
FROM update_count;

-- name: DeleteSubjectMapping :execrows
DELETE FROM subject_mappings WHERE id = $1;
