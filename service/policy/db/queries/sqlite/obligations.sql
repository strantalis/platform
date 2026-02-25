----------------------------------------------------------------
-- OBLIGATIONS
----------------------------------------------------------------

-- name: CreateObligation :one
WITH inserted_obligation AS (
    INSERT INTO obligation_definitions (id, namespace_id, name, metadata, created_at, updated_at)
    VALUES (
        gen_random_uuid(),
        COALESCE(
            sqlc.narg('namespace_id'),
            (SELECT fqns.namespace_id FROM attribute_fqns fqns
             WHERE fqns.fqn = sqlc.narg('namespace_fqn') AND sqlc.narg('namespace_id') IS NULL)
        ),
        sqlc.arg('name'),
        sqlc.arg('metadata'),
        STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'),
        STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
    )
    RETURNING id, namespace_id, name, metadata
),
inserted_values AS (
    INSERT INTO obligation_values_standard (id, obligation_definition_id, value, created_at, updated_at)
    SELECT
        gen_random_uuid(),
        io.id,
        vals.value,
        STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'),
        STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
    FROM inserted_obligation io
    CROSS JOIN json_each(sqlc.arg('values')) AS vals(value)
    WHERE sqlc.narg('values') IS NOT NULL
    RETURNING id, obligation_definition_id, value
)
SELECT
    io.id,
    io.name,
    io.metadata,
    JSON_BUILD_OBJECT(
        'id', n.id,
        'name', n.name,
        'fqn', fqns.fqn
    ) as namespace,
    COALESCE(
        JSON_AGG(
            CASE
                WHEN iv.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                    'id', iv.id,
                    'value', iv.value
                )
            END
        ),
        '[]'
    ) as "values"
FROM inserted_obligation io
JOIN attribute_namespaces n ON io.namespace_id = n.id
LEFT JOIN attribute_fqns fqns ON fqns.namespace_id = n.id AND fqns.attribute_id IS NULL AND fqns.value_id IS NULL
LEFT JOIN inserted_values iv ON iv.obligation_definition_id = io.id
GROUP BY io.id, io.name, io.metadata, n.id, fqns.fqn;

-- name: GetObligation :one
WITH obligation_triggers_agg AS (
    SELECT
        ot.obligation_value_id,
        JSON_AGG(
            JSON_BUILD_OBJECT(
                'id', ot.id,
                'action', JSON_BUILD_OBJECT(
                    'id', a.id,
                    'name', a.name
                ),
                'attribute_value', JSON_BUILD_OBJECT(
                    'id', av.id,
                    'value', av.value,
                    'fqn', COALESCE(av_fqns.fqn, '')
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
        ) as triggers
    FROM obligation_triggers ot
    JOIN actions a ON ot.action_id = a.id
    JOIN attribute_values av ON ot.attribute_value_id = av.id
    LEFT JOIN attribute_fqns av_fqns ON av_fqns.value_id = av.id
    GROUP BY ot.obligation_value_id
)
SELECT
    od.id,
    od.name,
    JSON_BUILD_OBJECT(
        'id', n.id,
        'name', n.name,
        'fqn', fqns.fqn
    ) as namespace,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(od.metadata, '$.labels'), 'created_at', od.created_at,'updated_at', od.updated_at)) as metadata,
    JSON_AGG(
        CASE
            WHEN ov.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                'id', ov.id,
                'value', ov.value,
                'triggers', COALESCE(ota.triggers, '[]')
            )
        END
    ) as "values"
FROM obligation_definitions od
JOIN attribute_namespaces n on od.namespace_id = n.id
LEFT JOIN attribute_fqns fqns ON fqns.namespace_id = n.id AND fqns.attribute_id IS NULL AND fqns.value_id IS NULL
LEFT JOIN obligation_values_standard ov on od.id = ov.obligation_definition_id
LEFT JOIN obligation_triggers_agg ota on ov.id = ota.obligation_value_id
WHERE
    -- lookup by obligation id OR by namespace fqn + obligation name
    (
        -- lookup by obligation id
        (sqlc.narg('id') IS NOT NULL AND od.id = sqlc.narg('id'))
        OR
        -- lookup by namespace fqn + obligation name
        (sqlc.narg('namespace_fqn') IS NOT NULL AND sqlc.narg('name') IS NOT NULL
         AND fqns.fqn = sqlc.narg('namespace_fqn') AND od.name = sqlc.narg('name'))
    )
GROUP BY od.id, n.id, fqns.fqn;

-- name: ListObligations :many
WITH counted AS (
    SELECT COUNT(od.id) AS total
    FROM obligation_definitions od
    LEFT JOIN attribute_namespaces n ON od.namespace_id = n.id
    LEFT JOIN attribute_fqns fqns ON fqns.namespace_id = n.id AND fqns.attribute_id IS NULL AND fqns.value_id IS NULL
    WHERE
        (sqlc.narg('namespace_id') IS NULL OR od.namespace_id = sqlc.narg('namespace_id')) AND
        (sqlc.narg('namespace_fqn') IS NULL OR fqns.fqn = sqlc.narg('namespace_fqn'))
),
obligation_triggers_agg AS (
    SELECT
        ot.obligation_value_id,
        JSON_AGG(
            JSON_BUILD_OBJECT(
                'id', ot.id,
                'action', JSON_BUILD_OBJECT(
                    'id', a.id,
                    'name', a.name
                ),
                'attribute_value', JSON_BUILD_OBJECT(
                    'id', av.id,
                    'value', av.value,
                    'fqn', COALESCE(av_fqns.fqn, '')
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
        ) as triggers
    FROM obligation_triggers ot
    JOIN actions a ON ot.action_id = a.id
    JOIN attribute_values av ON ot.attribute_value_id = av.id
    LEFT JOIN attribute_fqns av_fqns ON av_fqns.value_id = av.id
    GROUP BY ot.obligation_value_id
)
SELECT
    od.id,
    od.name,
    JSON_BUILD_OBJECT(
        'id', n.id,
        'name', n.name,
        'fqn', fqns.fqn
    ) as namespace,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(od.metadata, '$.labels'), 'created_at', od.created_at,'updated_at', od.updated_at)) as metadata,
    JSON_AGG(
        CASE
            WHEN ov.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                'id', ov.id,
                'value', ov.value,
                'triggers', COALESCE(ota.triggers, '[]')
            )
        END
    ) as "values",
    counted.total
FROM obligation_definitions od
JOIN attribute_namespaces n on od.namespace_id = n.id
LEFT JOIN attribute_fqns fqns ON fqns.namespace_id = n.id AND fqns.attribute_id IS NULL AND fqns.value_id IS NULL
CROSS JOIN counted
LEFT JOIN obligation_values_standard ov on od.id = ov.obligation_definition_id
LEFT JOIN obligation_triggers_agg ota on ov.id = ota.obligation_value_id
WHERE
    (sqlc.narg('namespace_id') IS NULL OR od.namespace_id = sqlc.narg('namespace_id')) AND
    (sqlc.narg('namespace_fqn') IS NULL OR fqns.fqn = sqlc.narg('namespace_fqn'))
GROUP BY od.id, n.id, fqns.fqn, counted.total
ORDER BY od.created_at DESC
LIMIT sqlc.arg('limit_')
OFFSET sqlc.arg('offset_');

-- name: UpdateObligation :execrows
UPDATE obligation_definitions
SET
    name = COALESCE(sqlc.narg('name'), name),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_at = STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now', '+0.001 seconds')
WHERE id = sqlc.arg('id');

-- name: DeleteObligation :one
DELETE FROM obligation_definitions 
WHERE id IN (
    SELECT od.id
    FROM obligation_definitions od
    LEFT JOIN attribute_namespaces n ON od.namespace_id = n.id
    LEFT JOIN attribute_fqns fqns ON fqns.namespace_id = n.id AND fqns.attribute_id IS NULL AND fqns.value_id IS NULL
    WHERE
        -- lookup by obligation id OR by namespace fqn + obligation name
        (
            -- lookup by obligation id
            (sqlc.narg('id') IS NOT NULL AND od.id = sqlc.narg('id'))
            OR
            -- lookup by namespace fqn + obligation name
            (sqlc.narg('namespace_fqn') IS NOT NULL AND sqlc.narg('name') IS NOT NULL 
             AND fqns.fqn = sqlc.narg('namespace_fqn') AND od.name = sqlc.narg('name'))
        )
)
RETURNING id;

-- name: GetObligationsByFQNs :many
WITH obligation_triggers_agg AS (
    SELECT
        ot.obligation_value_id,
        JSON_AGG(
            JSON_BUILD_OBJECT(
                'id', ot.id,
                'action', JSON_BUILD_OBJECT(
                    'id', a.id,
                    'name', a.name
                ),
                'attribute_value', JSON_BUILD_OBJECT(
                    'id', av.id,
                    'value', av.value,
                    'fqn', COALESCE(av_fqns.fqn, '')
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
        ) as triggers
    FROM obligation_triggers ot
    JOIN actions a ON ot.action_id = a.id
    JOIN attribute_values av ON ot.attribute_value_id = av.id
    LEFT JOIN attribute_fqns av_fqns ON av_fqns.value_id = av.id
    GROUP BY ot.obligation_value_id
)
SELECT
    od.id,
    od.name,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(od.metadata, '$.labels'), 'created_at', od.created_at,'updated_at', od.updated_at)) as metadata,
    JSON_BUILD_OBJECT(
        'id', n.id,
        'name', n.name,
        'fqn', fqns.fqn
    ) as namespace,
    COALESCE(
        JSON_AGG(
            CASE
                WHEN ov.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                    'id', ov.id,
                    'value', ov.value,
                    'triggers', COALESCE(ota.triggers, '[]')
                )
            END
        ),
        '[]'
    ) as "values"
FROM
    obligation_definitions od
JOIN
    attribute_namespaces n on od.namespace_id = n.id
JOIN
    attribute_fqns fqns ON fqns.namespace_id = n.id AND fqns.attribute_id IS NULL AND fqns.value_id IS NULL
JOIN
    (SELECT ns.value as ns_fqn, nm.value as obl_name
     FROM json_each(sqlc.arg('namespace_fqns')) AS ns(key, value)
     JOIN json_each(sqlc.arg('names')) AS nm(key, value) ON ns.key = nm.key) as fqn_pairs
ON
    fqns.fqn = fqn_pairs.ns_fqn AND od.name = fqn_pairs.obl_name
LEFT JOIN
    obligation_values_standard ov on od.id = ov.obligation_definition_id
LEFT JOIN
    obligation_triggers_agg ota on ov.id = ota.obligation_value_id
GROUP BY
    od.id, n.id, fqns.fqn;

----------------------------------------------------------------
-- OBLIGATION VALUES
----------------------------------------------------------------

-- name: CreateObligationValue :one
WITH obligation_lookup AS (
    SELECT od.id, od.name, od.metadata
    FROM obligation_definitions od
    LEFT JOIN attribute_namespaces n ON od.namespace_id = n.id
    LEFT JOIN attribute_fqns fqns ON fqns.namespace_id = n.id AND fqns.attribute_id IS NULL AND fqns.value_id IS NULL
    WHERE
        -- lookup by obligation id OR by namespace fqn + obligation name
        (
            -- lookup by obligation id
            (sqlc.narg('id') IS NOT NULL AND od.id = sqlc.narg('id'))
            OR
            -- lookup by namespace fqn + obligation name
            (sqlc.narg('namespace_fqn') IS NOT NULL AND sqlc.narg('name') IS NOT NULL 
             AND fqns.fqn = sqlc.narg('namespace_fqn') AND od.name = sqlc.narg('name'))
        )
),
inserted_value AS (
    INSERT INTO obligation_values_standard (id, obligation_definition_id, value, metadata, created_at, updated_at)
    SELECT
        gen_random_uuid(),
        ol.id,
        sqlc.arg('value'),
        sqlc.arg('metadata'),
        STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'),
        STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
    FROM obligation_lookup ol
    RETURNING id, obligation_definition_id, value, metadata
)
SELECT
    iv.id,
    ol.name,
    ol.id as obligation_id,
    JSON_BUILD_OBJECT(
        'id', n.id,
        'name', n.name,
        'fqn', fqns.fqn
    ) as namespace,
    iv.metadata as metadata
FROM inserted_value iv
JOIN obligation_lookup ol ON ol.id = iv.obligation_definition_id
JOIN obligation_definitions od ON od.id = ol.id
JOIN attribute_namespaces n ON od.namespace_id = n.id
LEFT JOIN attribute_fqns fqns ON fqns.namespace_id = n.id AND fqns.attribute_id IS NULL AND fqns.value_id IS NULL;

-- name: GetObligationValue :one
WITH obligation_triggers_agg AS (
    SELECT
        ot.obligation_value_id,
        JSON_AGG(
            JSON_BUILD_OBJECT(
                'id', ot.id,
                'action', JSON_BUILD_OBJECT(
                    'id', a.id,
                    'name', a.name
                ),
                'attribute_value', JSON_BUILD_OBJECT(
                    'id', av.id,
                    'value', av.value,
                    'fqn', COALESCE(av_fqns.fqn, '')
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
        ) as triggers
    FROM obligation_triggers ot
    JOIN actions a ON ot.action_id = a.id
    JOIN attribute_values av ON ot.attribute_value_id = av.id
    LEFT JOIN attribute_fqns av_fqns ON av_fqns.value_id = av.id
    GROUP BY ot.obligation_value_id
)
SELECT
    ov.id,
    ov.value,
    od.id as obligation_id,
    od.name,
    JSON_BUILD_OBJECT(
        'id', n.id,
        'name', n.name,
        'fqn', fqns.fqn
    ) as namespace,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(ov.metadata, '$.labels'), 'created_at', ov.created_at,'updated_at', ov.updated_at)) as metadata,
    COALESCE(ota.triggers, '[]') as triggers
FROM obligation_values_standard ov
JOIN obligation_definitions od ON ov.obligation_definition_id = od.id
JOIN attribute_namespaces n ON od.namespace_id = n.id
LEFT JOIN attribute_fqns fqns ON fqns.namespace_id = n.id AND fqns.attribute_id IS NULL AND fqns.value_id IS NULL
LEFT JOIN obligation_triggers_agg ota on ov.id = ota.obligation_value_id
WHERE
    -- lookup by value id OR by namespace fqn + obligation name + value name
    (
        -- lookup by value id
        (sqlc.narg('id') IS NOT NULL AND ov.id = sqlc.narg('id'))
        OR
        -- lookup by namespace fqn + obligation name + value name
        (sqlc.narg('namespace_fqn') IS NOT NULL AND sqlc.narg('name') IS NOT NULL AND sqlc.narg('value') IS NOT NULL
         AND fqns.fqn = sqlc.narg('namespace_fqn') AND od.name = sqlc.narg('name') AND ov.value = sqlc.narg('value'))
    );

-- name: UpdateObligationValue :execrows
UPDATE obligation_values_standard
SET
    value = COALESCE(sqlc.narg('value'), value),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_at = STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now', '+0.001 seconds')
WHERE id = sqlc.arg('id');

-- name: GetObligationValuesByFQNs :many
WITH obligation_triggers_agg AS (
    SELECT
        ot.obligation_value_id,
        JSON_AGG(
            JSON_BUILD_OBJECT(
                'id', ot.id,
                'action', JSON_BUILD_OBJECT(
                    'id', a.id,
                    'name', a.name
                ),
                'attribute_value', JSON_BUILD_OBJECT(
                    'id', av.id,
                    'value', av.value,
                    'fqn', COALESCE(av_fqns.fqn, '')
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
        ) as triggers
    FROM obligation_triggers ot
    JOIN actions a ON ot.action_id = a.id
    JOIN attribute_values av ON ot.attribute_value_id = av.id
    LEFT JOIN attribute_fqns av_fqns ON av_fqns.value_id = av.id
    GROUP BY ot.obligation_value_id
)
SELECT
    ov.id,
    ov.value,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(ov.metadata, '$.labels'), 'created_at', ov.created_at,'updated_at', ov.updated_at)) as metadata,
    od.id as obligation_id,
    od.name as name,
    JSON_BUILD_OBJECT(
        'id', n.id,
        'name', n.name,
        'fqn', fqns.fqn
    ) as namespace,
    COALESCE(ota.triggers, '[]') as triggers
FROM
    obligation_values_standard ov
JOIN
    obligation_definitions od ON ov.obligation_definition_id = od.id
JOIN
    attribute_namespaces n ON od.namespace_id = n.id
JOIN
    attribute_fqns fqns ON fqns.namespace_id = n.id AND fqns.attribute_id IS NULL AND fqns.value_id IS NULL
JOIN
    (SELECT ns.value as ns_fqn, nm.value as obl_name, v.value as value
     FROM json_each(sqlc.arg('namespace_fqns')) AS ns(key, value)
     JOIN json_each(sqlc.arg('names')) AS nm(key, value) ON ns.key = nm.key
     JOIN json_each(sqlc.arg('values')) AS v(key, value) ON ns.key = v.key) as fqn_pairs
ON
    fqns.fqn = fqn_pairs.ns_fqn AND od.name = fqn_pairs.obl_name AND ov.value = fqn_pairs.value
LEFT JOIN
    obligation_triggers_agg ota on ov.id = ota.obligation_value_id;

-- name: DeleteObligationValue :one
DELETE FROM obligation_values_standard
WHERE id IN (
    SELECT ov.id
    FROM obligation_values_standard ov
    JOIN obligation_definitions od ON ov.obligation_definition_id = od.id
    LEFT JOIN attribute_namespaces n ON od.namespace_id = n.id
    LEFT JOIN attribute_fqns fqns ON fqns.namespace_id = n.id AND fqns.attribute_id IS NULL AND fqns.value_id IS NULL
    WHERE
        -- lookup by value id OR by namespace fqn + obligation name + value name
        (
            -- lookup by value id
            (sqlc.narg('id') IS NOT NULL AND ov.id = sqlc.narg('id'))
            OR
            -- lookup by namespace fqn + obligation name + value
            (sqlc.narg('namespace_fqn') IS NOT NULL AND sqlc.narg('name') IS NOT NULL AND sqlc.narg('value') IS NOT NULL
             AND fqns.fqn = sqlc.narg('namespace_fqn') AND od.name = sqlc.narg('name') AND ov.value = sqlc.narg('value'))
        )
)
RETURNING id;

----------------------------------------------------------------
-- OBLIGATION TRIGGERS
----------------------------------------------------------------

-- name: CreateObligationTrigger :one
WITH ov_id AS (
    SELECT ov.id, od.namespace_id
    FROM obligation_values_standard ov
    JOIN obligation_definitions od ON ov.obligation_definition_id = od.id
    WHERE sqlc.narg('obligation_value_id') IS NOT NULL AND ov.id = sqlc.narg('obligation_value_id')
),
a_id AS (
    SELECT a.id
    FROM actions a
    WHERE
        (sqlc.narg('action_id') IS NOT NULL AND a.id = sqlc.narg('action_id'))
        OR
        (sqlc.narg('action_name') IS NOT NULL AND a.name = sqlc.narg('action_name'))
),
-- Gets the attribute value, but also ensures that the attribute value belongs to the same namespace as the obligation, to which the obligation value belongs
av_id AS (
    SELECT av.id
    FROM attribute_values av
    JOIN attribute_definitions ad ON av.attribute_definition_id = ad.id
    LEFT JOIN attribute_fqns fqns ON fqns.value_id = av.id
    WHERE
        ((sqlc.narg('attribute_value_id') IS NOT NULL AND av.id = sqlc.narg('attribute_value_id'))
        OR
        (sqlc.narg('attribute_value_fqn') IS NOT NULL AND fqns.fqn = sqlc.narg('attribute_value_fqn')))
        AND ad.namespace_id = (SELECT namespace_id FROM ov_id)
),
inserted AS (
    INSERT INTO obligation_triggers (id, obligation_value_id, action_id, attribute_value_id, metadata, client_id, created_at, updated_at)
    SELECT
        gen_random_uuid(),
        (SELECT id FROM ov_id),
        (SELECT id FROM a_id),
        (SELECT id FROM av_id),
        sqlc.arg('metadata'),
        sqlc.narg('client_id'),
        STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'),
        STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
    RETURNING id, obligation_value_id, action_id, attribute_value_id, metadata, created_at, updated_at, client_id
)
SELECT
    JSON_STRIP_NULLS(
        JSON_BUILD_OBJECT(
            'labels', json_extract(i.metadata, '$.labels'),
            'created_at', i.created_at,
            'updated_at', i.updated_at
        )
    ) AS metadata,
    JSON_STRIP_NULLS(
        JSON_BUILD_OBJECT(
            'id', i.id,
            'obligation_value', JSON_BUILD_OBJECT(
                'id', ov.id,
                'value', ov.value,
                'obligation', JSON_BUILD_OBJECT(
                    'id', od.id,
                    'name', od.name,
                    'namespace', JSON_BUILD_OBJECT(
                        'id', n.id,
                        'name', n.name,
                        'fqn', COALESCE(ns_fqns.fqn, '')
                    )
                )
            ),
            'action', JSON_BUILD_OBJECT(
                'id', a.id,
                'name', a.name
            ),
            'attribute_value', JSON_BUILD_OBJECT(
                'id', av.id,
                'value', av.value,
                'fqn', COALESCE(av_fqns.fqn, '')
            ),
            'context', CASE
                WHEN i.client_id IS NOT NULL THEN JSON_BUILD_ARRAY(
                    JSON_BUILD_OBJECT(
                        'pep', JSON_BUILD_OBJECT(
                            'client_id', i.client_id
                        )
                    ))
                ELSE '[]'
            END
        )
    ) as "trigger"
FROM inserted i
JOIN obligation_values_standard ov ON i.obligation_value_id = ov.id
JOIN obligation_definitions od ON ov.obligation_definition_id = od.id
JOIN attribute_namespaces n ON od.namespace_id = n.id
LEFT JOIN attribute_fqns ns_fqns ON ns_fqns.namespace_id = n.id AND ns_fqns.attribute_id IS NULL AND ns_fqns.value_id IS NULL
JOIN actions a ON i.action_id = a.id
JOIN attribute_values av ON i.attribute_value_id = av.id
LEFT JOIN attribute_fqns av_fqns ON av_fqns.value_id = av.id;

-- name: DeleteAllObligationTriggersForValue :execrows
DELETE FROM obligation_triggers
WHERE obligation_value_id = $1;


-- name: DeleteObligationTrigger :one
DELETE FROM obligation_triggers
WHERE id = $1
RETURNING id;

-- name: ListObligationTriggers :many
SELECT
    JSON_STRIP_NULLS(
        JSON_BUILD_OBJECT(
            'id', ot.id,
            'obligation_value', JSON_BUILD_OBJECT(
                'id', ov.id,
                'value', ov.value,
                'obligation', JSON_BUILD_OBJECT(
                    'id', od.id,
                    'name', od.name,
                    'namespace', JSON_BUILD_OBJECT(
                        'id', n.id,
                        'name', n.name,
                        'fqn', COALESCE(ns_fqns.fqn, '')
                    )
                )
            ),
            'action', JSON_BUILD_OBJECT(
                'id', a.id,
                'name', a.name
            ),
            'attribute_value', JSON_BUILD_OBJECT(
                'id', av.id,
                'value', av.value,
                'fqn', COALESCE(av_fqns.fqn, '')
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
    ) as "trigger",
    JSON_STRIP_NULLS(
        JSON_BUILD_OBJECT(
            'labels', json_extract(ot.metadata, '$.labels'),
            'created_at', ot.created_at,
            'updated_at', ot.updated_at
        )
    ) as metadata,
    COUNT(*) OVER() as total
FROM obligation_triggers ot
JOIN obligation_values_standard ov ON ot.obligation_value_id = ov.id
JOIN obligation_definitions od ON ov.obligation_definition_id = od.id
JOIN attribute_namespaces n ON od.namespace_id = n.id
LEFT JOIN attribute_fqns ns_fqns ON ns_fqns.namespace_id = n.id AND ns_fqns.attribute_id IS NULL AND ns_fqns.value_id IS NULL
JOIN actions a ON ot.action_id = a.id
JOIN attribute_values av ON ot.attribute_value_id = av.id
LEFT JOIN attribute_fqns av_fqns ON av_fqns.value_id = av.id
WHERE
    (sqlc.narg('namespace_id') IS NULL OR od.namespace_id = sqlc.narg('namespace_id')) AND
    (sqlc.narg('namespace_fqn') IS NULL OR ns_fqns.fqn = sqlc.narg('namespace_fqn'))
ORDER BY ot.created_at DESC
LIMIT sqlc.arg('limit_')
OFFSET sqlc.arg('offset_');
