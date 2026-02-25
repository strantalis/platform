---------------------------------------------------------------- 
-- ATTRIBUTE FQN
----------------------------------------------------------------

-- name: UpsertAttributeValueFqn :many
WITH new_fqns_cte AS (
    -- get attribute value fqns
    SELECT
        ns.id AS namespace_id,
        ad.id AS attribute_id,
        av.id AS value_id,
        CONCAT('https://', ns.name, '/attr/', ad.name, '/value/', av.value) AS fqn
    FROM attribute_values av
    INNER JOIN attribute_definitions AS ad ON av.attribute_definition_id = ad.id
    INNER JOIN attribute_namespaces AS ns ON ad.namespace_id = ns.id
    WHERE av.id = sqlc.arg('value_id') 
)

INSERT OR REPLACE INTO attribute_fqns (id, namespace_id, attribute_id, value_id, fqn)
SELECT
    COALESCE(
        (
            SELECT af.id
            FROM attribute_fqns af
            WHERE af.namespace_id IS nf.namespace_id
              AND af.attribute_id IS nf.attribute_id
              AND af.value_id IS nf.value_id
        ),
        gen_random_uuid()
    ),
    nf.namespace_id,
    nf.attribute_id,
    nf.value_id,
    nf.fqn
FROM new_fqns_cte nf
RETURNING
    COALESCE(namespace_id, '') AS namespace_id,
    COALESCE(attribute_id, '') AS attribute_id,
    COALESCE(value_id, '') AS value_id,
    fqn;

-- name: UpsertAttributeDefinitionFqn :many
WITH new_fqns_cte AS (
    -- get attribute definition fqns
    SELECT
        ns.id AS namespace_id,
        ad.id AS attribute_id,
        NULL AS value_id,
        CONCAT('https://', ns.name, '/attr/', ad.name) AS fqn
    FROM attribute_definitions ad
    JOIN attribute_namespaces ns ON ad.namespace_id = ns.id
    WHERE ad.id = sqlc.arg('attribute_id') 
    UNION
    -- get attribute value fqns
    SELECT
        ns.id as namespace_id,
        ad.id as attribute_id,
        av.id as value_id,
        CONCAT('https://', ns.name, '/attr/', ad.name, '/value/', av.value) AS fqn
    FROM attribute_values av
    JOIN attribute_definitions ad on av.attribute_definition_id = ad.id
    JOIN attribute_namespaces ns on ad.namespace_id = ns.id
    WHERE ad.id = sqlc.arg('attribute_id') 
)
INSERT OR REPLACE INTO attribute_fqns (id, namespace_id, attribute_id, value_id, fqn)
SELECT
    COALESCE(
        (
            SELECT af.id
            FROM attribute_fqns af
            WHERE af.namespace_id IS nf.namespace_id
              AND af.attribute_id IS nf.attribute_id
              AND af.value_id IS nf.value_id
        ),
        gen_random_uuid()
    ),
    nf.namespace_id,
    nf.attribute_id,
    nf.value_id,
    nf.fqn
FROM new_fqns_cte nf
RETURNING
    COALESCE(namespace_id, '') AS namespace_id,
    COALESCE(attribute_id, '') AS attribute_id,
    COALESCE(value_id, '') AS value_id,
    fqn;

-- name: UpsertAttributeNamespaceFqn :many
WITH new_fqns_cte AS (
    -- get namespace fqns
    SELECT
        ns.id as namespace_id,
        NULL as attribute_id,
        NULL as value_id,
        CONCAT('https://', ns.name) AS fqn
    FROM attribute_namespaces ns
    WHERE ns.id = sqlc.arg('namespace_id') 
    UNION
    -- get attribute definition fqns
    SELECT
        ns.id as namespace_id,
        ad.id as attribute_id,
        NULL as value_id,
        CONCAT('https://', ns.name, '/attr/', ad.name) AS fqn
    FROM attribute_definitions ad
    JOIN attribute_namespaces ns on ad.namespace_id = ns.id
    WHERE ns.id = sqlc.arg('namespace_id') 
    UNION
    -- get attribute value fqns
    SELECT
        ns.id as namespace_id,
        ad.id as attribute_id,
        av.id as value_id,
        CONCAT('https://', ns.name, '/attr/', ad.name, '/value/', av.value) AS fqn
    FROM attribute_values av
    JOIN attribute_definitions ad on av.attribute_definition_id = ad.id
    JOIN attribute_namespaces ns on ad.namespace_id = ns.id
    WHERE ns.id = sqlc.arg('namespace_id') 
)
INSERT OR REPLACE INTO attribute_fqns (id, namespace_id, attribute_id, value_id, fqn)
SELECT
    COALESCE(
        (
            SELECT af.id
            FROM attribute_fqns af
            WHERE af.namespace_id IS nf.namespace_id
              AND af.attribute_id IS nf.attribute_id
              AND af.value_id IS nf.value_id
        ),
        gen_random_uuid()
    ),
    nf.namespace_id,
    nf.attribute_id,
    nf.value_id,
    nf.fqn
FROM new_fqns_cte nf
RETURNING
    COALESCE(namespace_id, '') AS namespace_id,
    COALESCE(attribute_id, '') AS attribute_id,
    COALESCE(value_id, '') AS value_id,
    fqn;
