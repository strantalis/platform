package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	pkgdb "github.com/opentdf/platform/service/pkg/db"
	dbsqlite "github.com/opentdf/platform/service/policy/db/sqlite"
)

type pgQueries struct {
	*Queries
}

func (p pgQueries) WithTx(tx policyTx) policyQueries {
	pgxTx, ok := tx.(pgx.Tx)
	if !ok {
		return p
	}
	return pgQueries{Queries: p.Queries.WithTx(pgxTx)}
}

type sqliteQueries struct {
	q  *dbsqlite.Queries
	db dbsqlite.DBTX
}

func isValidAttributeDefinitionRule(rule AttributeDefinitionRule) bool {
	switch rule {
	case AttributeDefinitionRuleUNSPECIFIED,
		AttributeDefinitionRuleALLOF,
		AttributeDefinitionRuleANYOF,
		AttributeDefinitionRuleHIERARCHY:
		return true
	default:
		return false
	}
}

func (s sqliteQueries) WithTx(tx policyTx) policyQueries {
	sqlTx, ok := tx.(*sql.Tx)
	if !ok {
		return s
	}
	wrapper := wrapSQLiteTx(sqlTx)
	return sqliteQueries{q: dbsqlite.New(wrapper), db: wrapper}
}

func (s sqliteQueries) withSQLiteTx(ctx context.Context, fn func(d dbsqlite.DBTX) error) error {
	switch db := s.db.(type) {
	case sqliteTxWrapper:
		return fn(db)
	case sqliteDBWrapper:
		tx, err := db.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		wrapper := wrapSQLiteTx(tx)
		if err := fn(wrapper); err != nil {
			_ = tx.Rollback()
			return err
		}
		return tx.Commit()
	default:
		return fn(s.db)
	}
}

func (s sqliteQueries) assignPublicKeyToAttributeDefinition(ctx context.Context, arg assignPublicKeyToAttributeDefinitionParams) (AttributeDefinitionPublicKeyMap, error) {
	sqliteArg, err := convertStruct[dbsqlite.AssignPublicKeyToAttributeDefinitionParams](arg)
	if err != nil {
		return AttributeDefinitionPublicKeyMap{}, err
	}
	res, err := s.q.AssignPublicKeyToAttributeDefinition(ctx, sqliteArg)
	if err != nil {
		return AttributeDefinitionPublicKeyMap{}, err
	}
	return convertStruct[AttributeDefinitionPublicKeyMap](res)
}

func (s sqliteQueries) assignPublicKeyToAttributeValue(ctx context.Context, arg assignPublicKeyToAttributeValueParams) (AttributeValuePublicKeyMap, error) {
	sqliteArg, err := convertStruct[dbsqlite.AssignPublicKeyToAttributeValueParams](arg)
	if err != nil {
		return AttributeValuePublicKeyMap{}, err
	}
	res, err := s.q.AssignPublicKeyToAttributeValue(ctx, sqliteArg)
	if err != nil {
		return AttributeValuePublicKeyMap{}, err
	}
	return convertStruct[AttributeValuePublicKeyMap](res)
}

func (s sqliteQueries) assignPublicKeyToNamespace(ctx context.Context, arg assignPublicKeyToNamespaceParams) (AttributeNamespacePublicKeyMap, error) {
	sqliteArg, err := convertStruct[dbsqlite.AssignPublicKeyToNamespaceParams](arg)
	if err != nil {
		return AttributeNamespacePublicKeyMap{}, err
	}
	res, err := s.q.AssignPublicKeyToNamespace(ctx, sqliteArg)
	if err != nil {
		return AttributeNamespacePublicKeyMap{}, err
	}
	return convertStruct[AttributeNamespacePublicKeyMap](res)
}

func (s sqliteQueries) createAttribute(ctx context.Context, arg createAttributeParams) (string, error) {
	if !isValidAttributeDefinitionRule(arg.Rule) {
		return "", pkgdb.ErrEnumValueInvalid
	}
	sqliteArg, err := convertStruct[dbsqlite.CreateAttributeParams](arg)
	if err != nil {
		return "", err
	}
	res, err := s.q.CreateAttribute(ctx, sqliteArg)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s sqliteQueries) createAttributeValue(ctx context.Context, arg createAttributeValueParams) (string, error) {
	sqliteArg, err := convertStruct[dbsqlite.CreateAttributeValueParams](arg)
	if err != nil {
		return "", err
	}
	res, err := s.q.CreateAttributeValue(ctx, sqliteArg)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s sqliteQueries) createCustomAction(ctx context.Context, arg createCustomActionParams) (string, error) {
	sqliteArg, err := convertStruct[dbsqlite.CreateCustomActionParams](arg)
	if err != nil {
		return "", err
	}
	res, err := s.q.CreateCustomAction(ctx, sqliteArg)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s sqliteQueries) createKey(ctx context.Context, arg createKeyParams) (string, error) {
	sqliteArg, err := convertStruct[dbsqlite.CreateKeyParams](arg)
	if err != nil {
		return "", err
	}
	res, err := s.q.CreateKey(ctx, sqliteArg)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s sqliteQueries) createKeyAccessServer(ctx context.Context, arg createKeyAccessServerParams) (string, error) {
	sqliteArg, err := convertStruct[dbsqlite.CreateKeyAccessServerParams](arg)
	if err != nil {
		return "", err
	}
	res, err := s.q.CreateKeyAccessServer(ctx, sqliteArg)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s sqliteQueries) createNamespace(ctx context.Context, arg createNamespaceParams) (string, error) {
	sqliteArg, err := convertStruct[dbsqlite.CreateNamespaceParams](arg)
	if err != nil {
		return "", err
	}
	res, err := s.q.CreateNamespace(ctx, sqliteArg)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s sqliteQueries) createObligation(ctx context.Context, arg createObligationParams) (createObligationRow, error) {
	var namespaceID string
	if arg.NamespaceID.Valid {
		namespaceID = UUIDToString(arg.NamespaceID)
	} else if arg.NamespaceFqn.Valid {
		if err := s.db.QueryRowContext(ctx, `
SELECT namespace_id
FROM attribute_fqns
WHERE fqn = $1
  AND attribute_id IS NULL
  AND value_id IS NULL
`, arg.NamespaceFqn.String).Scan(&namespaceID); err != nil {
			return createObligationRow{}, err
		}
	}
	if namespaceID == "" {
		return createObligationRow{}, sql.ErrNoRows
	}

	var out createObligationRow
	err := s.withSQLiteTx(ctx, func(db dbsqlite.DBTX) error {
		var obligationID string
		if err := db.QueryRowContext(ctx, `
INSERT INTO obligation_definitions (id, namespace_id, name, metadata, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'),
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
)
RETURNING id
`, namespaceID, arg.Name, arg.Metadata).Scan(&obligationID); err != nil {
			return err
		}

		if len(arg.Values) > 0 {
			for _, value := range arg.Values {
				if _, err := db.ExecContext(ctx, `
INSERT INTO obligation_values_standard (id, obligation_definition_id, value, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'),
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
)
`, obligationID, value); err != nil {
					return err
				}
			}
		}

		row := db.QueryRowContext(ctx, `
SELECT
    od.id,
    od.name,
    CAST(od.metadata AS BLOB) AS metadata,
    CAST(JSON_BUILD_OBJECT(
        'id', n.id,
        'name', n.name,
        'fqn', fqns.fqn
    ) AS BLOB) AS namespace,
    CAST(COALESCE(
        JSON_AGG(
            CASE
                WHEN ov.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                    'id', ov.id,
                    'value', ov.value
                )
            END
        ),
        '[]'
    ) AS BLOB) AS obligation_values
FROM obligation_definitions od
JOIN attribute_namespaces n ON od.namespace_id = n.id
LEFT JOIN attribute_fqns fqns ON fqns.namespace_id = n.id
    AND fqns.attribute_id IS NULL
    AND fqns.value_id IS NULL
LEFT JOIN obligation_values_standard ov ON ov.obligation_definition_id = od.id
WHERE od.id = $1
GROUP BY od.id, od.name, od.metadata, n.id, fqns.fqn
`, obligationID)
		if err := row.Scan(&out.ID, &out.Name, &out.Metadata, &out.Namespace, &out.Values); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return createObligationRow{}, err
	}
	return out, nil
}

func (s sqliteQueries) createObligationTrigger(ctx context.Context, arg createObligationTriggerParams) (createObligationTriggerRow, error) {
	if !arg.ObligationValueID.Valid {
		return createObligationTriggerRow{}, pkgdb.ErrInvalidOblTriParam
	}

	var out createObligationTriggerRow
	err := s.withSQLiteTx(ctx, func(db dbsqlite.DBTX) error {
		var namespaceID string
		if err := db.QueryRowContext(ctx, `
SELECT od.namespace_id
FROM obligation_values_standard ov
JOIN obligation_definitions od ON ov.obligation_definition_id = od.id
WHERE ov.id = $1
`, UUIDToString(arg.ObligationValueID)).Scan(&namespaceID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return pkgdb.ErrInvalidOblTriParam
			}
			return err
		}

		var actionID string
		switch {
		case arg.ActionID.Valid:
			if err := db.QueryRowContext(ctx, `
SELECT id
FROM actions
WHERE id = $1
`, UUIDToString(arg.ActionID)).Scan(&actionID); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return pkgdb.ErrInvalidOblTriParam
				}
				return err
			}
		case arg.ActionName.Valid:
			if err := db.QueryRowContext(ctx, `
SELECT id
FROM actions
WHERE name = $1
`, arg.ActionName.String).Scan(&actionID); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return pkgdb.ErrInvalidOblTriParam
				}
				return err
			}
		default:
			return pkgdb.ErrInvalidOblTriParam
		}

		var attributeValueID string
		switch {
		case arg.AttributeValueID.Valid:
			var attributeNamespaceID string
			if err := db.QueryRowContext(ctx, `
SELECT ad.namespace_id
FROM attribute_values av
JOIN attribute_definitions ad ON av.attribute_definition_id = ad.id
WHERE av.id = $1
`, UUIDToString(arg.AttributeValueID)).Scan(&attributeNamespaceID); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return pkgdb.ErrNotNullViolation
				}
				return err
			}
			if attributeNamespaceID != namespaceID {
				return pkgdb.ErrInvalidOblTriParam
			}
			attributeValueID = UUIDToString(arg.AttributeValueID)
		case arg.AttributeValueFqn.Valid:
			var attributeNamespaceID string
			if err := db.QueryRowContext(ctx, `
SELECT av.id, ad.namespace_id
FROM attribute_values av
JOIN attribute_definitions ad ON av.attribute_definition_id = ad.id
LEFT JOIN attribute_fqns fqns ON fqns.value_id = av.id
WHERE fqns.fqn = $1
`, arg.AttributeValueFqn.String).Scan(&attributeValueID, &attributeNamespaceID); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return pkgdb.ErrNotNullViolation
				}
				return err
			}
			if attributeNamespaceID != namespaceID {
				return pkgdb.ErrInvalidOblTriParam
			}
		default:
			return pkgdb.ErrInvalidOblTriParam
		}

		var triggerID string
		var clientID any
		if arg.ClientID.Valid {
			clientID = arg.ClientID.String
		}
		if err := db.QueryRowContext(ctx, `
INSERT INTO obligation_triggers (
    id,
    obligation_value_id,
    action_id,
    attribute_value_id,
    metadata,
    client_id,
    created_at,
    updated_at
)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    $4,
    $5,
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'),
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
)
RETURNING id
`, UUIDToString(arg.ObligationValueID), actionID, attributeValueID, arg.Metadata, clientID).Scan(&triggerID); err != nil {
			return err
		}

		row := db.QueryRowContext(ctx, `
SELECT
    CAST(JSON_STRIP_NULLS(
        JSON_BUILD_OBJECT(
            'labels', json_extract(i.metadata, '$.labels'),
            'created_at', i.created_at,
            'updated_at', i.updated_at
        )
    ) AS BLOB) AS metadata,
    CAST(JSON_STRIP_NULLS(
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
    ) AS BLOB) AS "trigger"
FROM obligation_triggers i
JOIN obligation_values_standard ov ON i.obligation_value_id = ov.id
JOIN obligation_definitions od ON ov.obligation_definition_id = od.id
JOIN attribute_namespaces n ON od.namespace_id = n.id
LEFT JOIN attribute_fqns ns_fqns ON ns_fqns.namespace_id = n.id
    AND ns_fqns.attribute_id IS NULL
    AND ns_fqns.value_id IS NULL
JOIN actions a ON i.action_id = a.id
JOIN attribute_values av ON i.attribute_value_id = av.id
LEFT JOIN attribute_fqns av_fqns ON av_fqns.value_id = av.id
WHERE i.id = $1
`, triggerID)
		return row.Scan(&out.Metadata, &out.Trigger)
	})
	if err != nil {
		return createObligationTriggerRow{}, err
	}
	return out, nil
}

func (s sqliteQueries) createObligationValue(ctx context.Context, arg createObligationValueParams) (createObligationValueRow, error) {
	var obligationID string
	if arg.ID.Valid {
		if err := s.db.QueryRowContext(ctx, `
SELECT id
FROM obligation_definitions
WHERE id = $1
`, UUIDToString(arg.ID)).Scan(&obligationID); err != nil {
			return createObligationValueRow{}, err
		}
	} else if arg.NamespaceFqn.Valid && arg.Name.Valid {
		if err := s.db.QueryRowContext(ctx, `
SELECT od.id
FROM obligation_definitions od
JOIN attribute_namespaces n ON od.namespace_id = n.id
LEFT JOIN attribute_fqns fqns ON fqns.namespace_id = n.id
    AND fqns.attribute_id IS NULL
    AND fqns.value_id IS NULL
WHERE fqns.fqn = $1 AND od.name = $2
`, arg.NamespaceFqn.String, arg.Name.String).Scan(&obligationID); err != nil {
			return createObligationValueRow{}, err
		}
	}
	if obligationID == "" {
		return createObligationValueRow{}, sql.ErrNoRows
	}

	var out createObligationValueRow
	err := s.withSQLiteTx(ctx, func(db dbsqlite.DBTX) error {
		var valueID string
		if err := db.QueryRowContext(ctx, `
INSERT INTO obligation_values_standard (id, obligation_definition_id, value, metadata, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    $1,
    $2,
    $3,
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'),
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
)
RETURNING id
`, obligationID, arg.Value, arg.Metadata).Scan(&valueID); err != nil {
			return err
		}

		row := db.QueryRowContext(ctx, `
SELECT
    ov.id,
    od.name,
    od.id as obligation_id,
    CAST(JSON_BUILD_OBJECT(
        'id', n.id,
        'name', n.name,
        'fqn', fqns.fqn
    ) AS BLOB) as namespace,
    CAST(ov.metadata AS BLOB) as metadata
FROM obligation_values_standard ov
JOIN obligation_definitions od ON ov.obligation_definition_id = od.id
JOIN attribute_namespaces n ON od.namespace_id = n.id
LEFT JOIN attribute_fqns fqns ON fqns.namespace_id = n.id
    AND fqns.attribute_id IS NULL
    AND fqns.value_id IS NULL
WHERE ov.id = $1
`, valueID)
		return row.Scan(&out.ID, &out.Name, &out.ObligationID, &out.Namespace, &out.Metadata)
	})
	if err != nil {
		return createObligationValueRow{}, err
	}
	return out, nil
}

func (s sqliteQueries) createOrListActionsByName(ctx context.Context, actionNames []string) ([]createOrListActionsByNameRow, error) {
	if len(actionNames) == 0 {
		return []createOrListActionsByNameRow{}, nil
	}

	payloadBytes, err := json.Marshal(actionNames)
	if err != nil {
		return nil, err
	}
	payload := string(payloadBytes)

	preExisting := make(map[string]bool, len(actionNames))
	rows := make([]createOrListActionsByNameRow, 0, len(actionNames))

	parseSQLiteTimestamp := func(value string) pgtype.Timestamptz {
		value = strings.TrimSpace(value)
		if value == "" {
			return pgtype.Timestamptz{}
		}
		layouts := []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02 15:04:05.999999999",
			"2006-01-02 15:04:05.999999",
			"2006-01-02 15:04:05.999",
			"2006-01-02 15:04:05",
		}
		for _, layout := range layouts {
			if ts, parseErr := time.Parse(layout, value); parseErr == nil {
				return pgtype.Timestamptz{Time: ts, Valid: true}
			}
		}
		return pgtype.Timestamptz{}
	}

	err = s.withSQLiteTx(ctx, func(db dbsqlite.DBTX) error {
		existingRows, err := db.QueryContext(ctx, `
SELECT LOWER(name)
FROM actions
WHERE LOWER(name) IN (SELECT LOWER(value) FROM json_each(?))
`, payload)
		if err != nil {
			return err
		}
		defer existingRows.Close()
		for existingRows.Next() {
			var name string
			if err := existingRows.Scan(&name); err != nil {
				return err
			}
			preExisting[name] = true
		}
		if err := existingRows.Err(); err != nil {
			return err
		}

		_, err = db.ExecContext(ctx, `
INSERT INTO actions (id, name, is_standard, created_at, updated_at)
SELECT
    gen_random_uuid(),
    input.value,
    FALSE,
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now'),
    STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now')
FROM json_each(?) AS input
WHERE NOT EXISTS (
    SELECT 1 FROM actions a WHERE LOWER(a.name) = LOWER(input.value)
)
`, payload)
		if err != nil {
			return err
		}

		allRows, err := db.QueryContext(ctx, `
SELECT id, name, is_standard, created_at
FROM actions
WHERE LOWER(name) IN (SELECT LOWER(value) FROM json_each(?))
ORDER BY name
`, payload)
		if err != nil {
			return err
		}
		defer allRows.Close()

		for allRows.Next() {
			var (
				id          string
				name        string
				isStandard  sql.NullBool
				createdAt   sql.NullString
				createdTime pgtype.Timestamptz
			)
			if err := allRows.Scan(&id, &name, &isStandard, &createdAt); err != nil {
				return err
			}
			if createdAt.Valid {
				createdTime = parseSQLiteTimestamp(createdAt.String)
			}
			rows = append(rows, createOrListActionsByNameRow{
				ID:          id,
				Name:        name,
				IsStandard:  isStandard.Valid && isStandard.Bool,
				CreatedAt:   createdTime,
				PreExisting: preExisting[strings.ToLower(name)],
			})
		}
		if err := allRows.Err(); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s sqliteQueries) createProviderConfig(ctx context.Context, arg createProviderConfigParams) (createProviderConfigRow, error) {
	var out createProviderConfigRow
	if len(arg.Config) == 0 {
		return createProviderConfigRow{}, pkgdb.ErrNotNullViolation
	}
	if !json.Valid(arg.Config) {
		return createProviderConfigRow{}, pkgdb.ErrEnumValueInvalid
	}
	err := s.withSQLiteTx(ctx, func(db dbsqlite.DBTX) error {
		var id string
		if err := db.QueryRowContext(ctx, `
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
RETURNING id
`, arg.ProviderName, arg.Manager, arg.Config, arg.Metadata).Scan(&id); err != nil {
			return err
		}

		row := db.QueryRowContext(ctx, `
SELECT
    pc.id,
    pc.provider_name,
    pc.manager,
    CAST(pc.config AS BLOB) AS config,
    CAST(JSON_STRIP_NULLS(
        JSON_BUILD_OBJECT(
            'labels', json_extract(pc.metadata, '$.labels'),
            'created_at', pc.created_at,
            'updated_at', pc.updated_at
        )
    ) AS BLOB) AS metadata
FROM provider_config AS pc
WHERE pc.id = $1
`, id)
		return row.Scan(&out.ID, &out.ProviderName, &out.Manager, &out.Config, &out.Metadata)
	})
	if err != nil {
		return createProviderConfigRow{}, err
	}
	return out, nil
}

func (s sqliteQueries) createRegisteredResource(ctx context.Context, arg createRegisteredResourceParams) (string, error) {
	sqliteArg, err := convertStruct[dbsqlite.CreateRegisteredResourceParams](arg)
	if err != nil {
		return "", err
	}
	res, err := s.q.CreateRegisteredResource(ctx, sqliteArg)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s sqliteQueries) createRegisteredResourceActionAttributeValues(ctx context.Context, arg []createRegisteredResourceActionAttributeValuesParams) (int64, error) {
	var count int64
	for _, row := range arg {
		sqliteArg, err := convertStruct[dbsqlite.CreateRegisteredResourceActionAttributeValuesParams](row)
		if err != nil {
			return count, err
		}
		if err := s.q.CreateRegisteredResourceActionAttributeValues(ctx, sqliteArg); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (s sqliteQueries) createRegisteredResourceValue(ctx context.Context, arg createRegisteredResourceValueParams) (string, error) {
	sqliteArg, err := convertStruct[dbsqlite.CreateRegisteredResourceValueParams](arg)
	if err != nil {
		return "", err
	}
	res, err := s.q.CreateRegisteredResourceValue(ctx, sqliteArg)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s sqliteQueries) createResourceMapping(ctx context.Context, arg createResourceMappingParams) (string, error) {
	sqliteArg, err := convertStruct[dbsqlite.CreateResourceMappingParams](arg)
	if err != nil {
		return "", err
	}
	res, err := s.q.CreateResourceMapping(ctx, sqliteArg)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s sqliteQueries) createResourceMappingGroup(ctx context.Context, arg createResourceMappingGroupParams) (string, error) {
	sqliteArg, err := convertStruct[dbsqlite.CreateResourceMappingGroupParams](arg)
	if err != nil {
		return "", err
	}
	res, err := s.q.CreateResourceMappingGroup(ctx, sqliteArg)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s sqliteQueries) createSubjectConditionSet(ctx context.Context, arg createSubjectConditionSetParams) (string, error) {
	sqliteArg, err := convertStruct[dbsqlite.CreateSubjectConditionSetParams](arg)
	if err != nil {
		return "", err
	}
	res, err := s.q.CreateSubjectConditionSet(ctx, sqliteArg)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s sqliteQueries) createSubjectMapping(ctx context.Context, arg createSubjectMappingParams) (string, error) {
	sqliteArg, err := convertStruct[dbsqlite.CreateSubjectMappingParams](arg)
	if err != nil {
		return "", err
	}
	var createdID string
	err = s.withSQLiteTx(ctx, func(d dbsqlite.DBTX) error {
		row := d.QueryRowContext(ctx, `
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
`, sqliteArg.AttributeValueID, sqliteArg.Metadata, sqliteArg.SubjectConditionSetID)
		if scanErr := row.Scan(&createdID); scanErr != nil {
			return fmt.Errorf("insert subject mapping: %w", scanErr)
		}

		if len(arg.ActionIds) == 0 {
			return nil
		}

		_, execErr := d.ExecContext(ctx, `
INSERT INTO subject_mapping_actions (subject_mapping_id, action_id)
SELECT $1, value FROM json_each($2)
`, createdID, sqliteArg.ActionIds)
		if execErr != nil {
			return fmt.Errorf("insert subject mapping actions: %w", execErr)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return createdID, nil
}

func (s sqliteQueries) deleteAllObligationTriggersForValue(ctx context.Context, obligationValueID string) (int64, error) {
	res, err := s.q.DeleteAllObligationTriggersForValue(ctx, sql.NullString{String: obligationValueID, Valid: obligationValueID != ""})
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) deleteAllUnmappedSubjectConditionSets(ctx context.Context) ([]string, error) {
	res, err := s.q.DeleteAllUnmappedSubjectConditionSets(ctx)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s sqliteQueries) deleteAttribute(ctx context.Context, id string) (int64, error) {
	res, err := s.q.DeleteAttribute(ctx, id)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) deleteAttributeValue(ctx context.Context, id string) (int64, error) {
	res, err := s.q.DeleteAttributeValue(ctx, id)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) deleteCustomAction(ctx context.Context, id string) (int64, error) {
	res, err := s.q.DeleteCustomAction(ctx, id)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) deleteKey(ctx context.Context, id string) (int64, error) {
	res, err := s.q.DeleteKey(ctx, id)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) deleteKeyAccessServer(ctx context.Context, id string) (int64, error) {
	res, err := s.q.DeleteKeyAccessServer(ctx, id)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) deleteNamespace(ctx context.Context, id string) (int64, error) {
	res, err := s.q.DeleteNamespace(ctx, id)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) deleteObligation(ctx context.Context, arg deleteObligationParams) (string, error) {
	sqliteArg, err := convertStruct[dbsqlite.DeleteObligationParams](arg)
	if err != nil {
		return "", err
	}
	res, err := s.q.DeleteObligation(ctx, sqliteArg)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s sqliteQueries) deleteObligationTrigger(ctx context.Context, id string) (string, error) {
	res, err := s.q.DeleteObligationTrigger(ctx, id)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s sqliteQueries) deleteObligationValue(ctx context.Context, arg deleteObligationValueParams) (string, error) {
	sqliteArg, err := convertStruct[dbsqlite.DeleteObligationValueParams](arg)
	if err != nil {
		return "", err
	}
	res, err := s.q.DeleteObligationValue(ctx, sqliteArg)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (s sqliteQueries) deleteProviderConfig(ctx context.Context, id string) (int64, error) {
	res, err := s.q.DeleteProviderConfig(ctx, id)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) deleteRegisteredResource(ctx context.Context, id string) (int64, error) {
	res, err := s.q.DeleteRegisteredResource(ctx, id)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) deleteRegisteredResourceActionAttributeValues(ctx context.Context, registeredResourceValueID string) (int64, error) {
	res, err := s.q.DeleteRegisteredResourceActionAttributeValues(ctx, sql.NullString{String: registeredResourceValueID, Valid: registeredResourceValueID != ""})
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) deleteRegisteredResourceValue(ctx context.Context, id string) (int64, error) {
	res, err := s.q.DeleteRegisteredResourceValue(ctx, id)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) deleteResourceMapping(ctx context.Context, id string) (int64, error) {
	res, err := s.q.DeleteResourceMapping(ctx, id)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) deleteResourceMappingGroup(ctx context.Context, id string) (int64, error) {
	res, err := s.q.DeleteResourceMappingGroup(ctx, id)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) deleteSubjectConditionSet(ctx context.Context, id string) (int64, error) {
	res, err := s.q.DeleteSubjectConditionSet(ctx, id)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) deleteSubjectMapping(ctx context.Context, id string) (int64, error) {
	res, err := s.q.DeleteSubjectMapping(ctx, id)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) getAction(ctx context.Context, arg getActionParams) (getActionRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.GetActionParams](arg)
	if err != nil {
		return getActionRow{}, err
	}
	res, err := s.q.GetAction(ctx, sqliteArg)
	if err != nil {
		return getActionRow{}, err
	}
	return convertStruct[getActionRow](res)
}

func (s sqliteQueries) getAttribute(ctx context.Context, arg getAttributeParams) (getAttributeRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.GetAttributeParams](arg)
	if err != nil {
		return getAttributeRow{}, err
	}
	res, err := s.q.GetAttribute(ctx, sqliteArg)
	if err != nil {
		return getAttributeRow{}, err
	}
	return convertStruct[getAttributeRow](res)
}

func (s sqliteQueries) getAttributeValue(ctx context.Context, arg getAttributeValueParams) (getAttributeValueRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.GetAttributeValueParams](arg)
	if err != nil {
		return getAttributeValueRow{}, err
	}
	res, err := s.q.GetAttributeValue(ctx, sqliteArg)
	if err != nil {
		return getAttributeValueRow{}, err
	}
	return convertStruct[getAttributeValueRow](res)
}

func (s sqliteQueries) getBaseKey(ctx context.Context) ([]byte, error) {
	res, err := s.q.GetBaseKey(ctx)
	if err != nil {
		return nil, err
	}
	bytes, err := toBytes(reflect.ValueOf(res))
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (s sqliteQueries) getKey(ctx context.Context, arg getKeyParams) (getKeyRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.GetKeyParams](arg)
	if err != nil {
		return getKeyRow{}, err
	}
	res, err := s.q.GetKey(ctx, sqliteArg)
	if err != nil {
		return getKeyRow{}, err
	}
	return convertStruct[getKeyRow](res)
}

func (s sqliteQueries) getKeyAccessServer(ctx context.Context, arg getKeyAccessServerParams) (getKeyAccessServerRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.GetKeyAccessServerParams](arg)
	if err != nil {
		return getKeyAccessServerRow{}, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT
    kas.id,
    kas.uri,
    kas.public_key,
    kas.name,
    kas.source_type,
    CAST(JSON_STRIP_NULLS(
        JSON_BUILD_OBJECT(
            'labels', json_extract(kas.metadata, '$.labels'),
            'created_at', kas.created_at,
            'updated_at', kas.updated_at
        )
    ) AS BLOB) AS metadata,
    CAST(COALESCE(kask_keys.keys, JSON('[]')) AS BLOB) AS keys
FROM key_access_servers AS kas
LEFT JOIN (
        SELECT
            kask.key_access_server_id,
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
        FROM key_access_server_keys kask
        INNER JOIN key_access_servers kas ON kask.key_access_server_id = kas.id
        GROUP BY kask.key_access_server_id
    ) kask_keys ON kas.id = kask_keys.key_access_server_id
WHERE ($1 IS NULL OR kas.id = $1)
  AND ($2 IS NULL OR kas.name = $2)
  AND ($3 IS NULL OR kas.uri = $3)
`, sqliteArg.ID, sqliteArg.Name, sqliteArg.Uri)
	var res dbsqlite.GetKeyAccessServerRow
	if err := row.Scan(
		&res.ID,
		&res.Uri,
		&res.PublicKey,
		&res.Name,
		&res.SourceType,
		&res.Metadata,
		&res.Keys,
	); err != nil {
		return getKeyAccessServerRow{}, err
	}
	return convertStruct[getKeyAccessServerRow](res)
}

func (s sqliteQueries) getNamespace(ctx context.Context, arg getNamespaceParams) (getNamespaceRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.GetNamespaceParams](arg)
	if err != nil {
		return getNamespaceRow{}, err
	}
	res, err := s.q.GetNamespace(ctx, sqliteArg)
	if err != nil {
		return getNamespaceRow{}, err
	}
	return convertStruct[getNamespaceRow](res)
}

func (s sqliteQueries) getObligation(ctx context.Context, arg getObligationParams) (getObligationRow, error) {
	var id interface{}
	if arg.ID.Valid {
		id = UUIDToString(arg.ID)
	}
	var namespaceFqn interface{}
	if arg.NamespaceFqn.Valid {
		namespaceFqn = arg.NamespaceFqn.String
	}
	var name interface{}
	if arg.Name.Valid {
		name = arg.Name.String
	}

	row := s.db.QueryRowContext(ctx, `
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
    CAST(JSON_BUILD_OBJECT(
        'id', n.id,
        'name', n.name,
        'fqn', fqns.fqn
    ) AS BLOB) as namespace,
    CAST(JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(od.metadata, '$.labels'), 'created_at', od.created_at,'updated_at', od.updated_at)) AS BLOB) as metadata,
    CAST(JSON_AGG(
        CASE
            WHEN ov.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                'id', ov.id,
                'value', ov.value,
                'triggers', COALESCE(ota.triggers, '[]')
            )
        END
    ) AS BLOB) as obligation_values
FROM obligation_definitions od
JOIN attribute_namespaces n on od.namespace_id = n.id
LEFT JOIN attribute_fqns fqns ON fqns.namespace_id = n.id AND fqns.attribute_id IS NULL AND fqns.value_id IS NULL
LEFT JOIN obligation_values_standard ov on od.id = ov.obligation_definition_id
LEFT JOIN obligation_triggers_agg ota on ov.id = ota.obligation_value_id
WHERE
    (
        ($1 IS NOT NULL AND od.id = $1)
        OR
        ($2 IS NOT NULL AND $3 IS NOT NULL
         AND fqns.fqn = $2 AND od.name = $3)
    )
GROUP BY od.id, n.id, fqns.fqn
`, id, namespaceFqn, name)

	var out getObligationRow
	if err := row.Scan(&out.ID, &out.Name, &out.Namespace, &out.Metadata, &out.Values); err != nil {
		return getObligationRow{}, err
	}
	return out, nil
}

func (s sqliteQueries) getObligationValue(ctx context.Context, arg getObligationValueParams) (getObligationValueRow, error) {
	var id interface{}
	if arg.ID.Valid {
		id = UUIDToString(arg.ID)
	}
	var namespaceFqn interface{}
	if arg.NamespaceFqn.Valid {
		namespaceFqn = arg.NamespaceFqn.String
	}
	var name interface{}
	if arg.Name.Valid {
		name = arg.Name.String
	}
	var value interface{}
	if arg.Value.Valid {
		value = arg.Value.String
	}

	row := s.db.QueryRowContext(ctx, `
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
    CAST(JSON_BUILD_OBJECT(
        'id', n.id,
        'name', n.name,
        'fqn', fqns.fqn
    ) AS BLOB) as namespace,
    CAST(JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(ov.metadata, '$.labels'), 'created_at', ov.created_at,'updated_at', ov.updated_at)) AS BLOB) as metadata,
    CAST(COALESCE(ota.triggers, '[]') AS BLOB) as triggers
FROM obligation_values_standard ov
JOIN obligation_definitions od ON ov.obligation_definition_id = od.id
JOIN attribute_namespaces n ON od.namespace_id = n.id
LEFT JOIN attribute_fqns fqns ON fqns.namespace_id = n.id AND fqns.attribute_id IS NULL AND fqns.value_id IS NULL
LEFT JOIN obligation_triggers_agg ota on ov.id = ota.obligation_value_id
WHERE
    (
        ($1 IS NOT NULL AND ov.id = $1)
        OR
        ($2 IS NOT NULL AND $3 IS NOT NULL AND $4 IS NOT NULL
         AND fqns.fqn = $2 AND od.name = $3 AND ov.value = $4)
    )
`, id, namespaceFqn, name, value)

	var out getObligationValueRow
	if err := row.Scan(&out.ID, &out.Value, &out.ObligationID, &out.Name, &out.Namespace, &out.Metadata, &out.Triggers); err != nil {
		return getObligationValueRow{}, err
	}
	return out, nil
}

func (s sqliteQueries) getObligationValuesByFQNs(ctx context.Context, arg getObligationValuesByFQNsParams) ([]getObligationValuesByFQNsRow, error) {
	if len(arg.NamespaceFqns) == 0 || len(arg.Names) == 0 || len(arg.Values) == 0 {
		return []getObligationValuesByFQNsRow{}, nil
	}
	namespaceFqnsJSON, err := json.Marshal(arg.NamespaceFqns)
	if err != nil {
		return nil, err
	}
	namesJSON, err := json.Marshal(arg.Names)
	if err != nil {
		return nil, err
	}
	valuesJSON, err := json.Marshal(arg.Values)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
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
    CAST(JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(ov.metadata, '$.labels'), 'created_at', ov.created_at,'updated_at', ov.updated_at)) AS BLOB) as metadata,
    od.id as obligation_id,
    od.name as name,
    CAST(JSON_BUILD_OBJECT(
        'id', n.id,
        'name', n.name,
        'fqn', fqns.fqn
    ) AS BLOB) as namespace,
    CAST(COALESCE(ota.triggers, '[]') AS BLOB) as triggers
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
     FROM json_each($1) AS ns
     JOIN json_each($2) AS nm ON ns.key = nm.key
     JOIN json_each($3) AS v ON ns.key = v.key) as fqn_pairs
ON
    fqns.fqn = fqn_pairs.ns_fqn AND od.name = fqn_pairs.obl_name AND ov.value = fqn_pairs.value
LEFT JOIN
    obligation_triggers_agg ota on ov.id = ota.obligation_value_id
`, string(namespaceFqnsJSON), string(namesJSON), string(valuesJSON))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []getObligationValuesByFQNsRow{}
	for rows.Next() {
		var item getObligationValuesByFQNsRow
		if err := rows.Scan(&item.ID, &item.Value, &item.Metadata, &item.ObligationID, &item.Name, &item.Namespace, &item.Triggers); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s sqliteQueries) getObligationsByFQNs(ctx context.Context, arg getObligationsByFQNsParams) ([]getObligationsByFQNsRow, error) {
	if len(arg.NamespaceFqns) == 0 || len(arg.Names) == 0 {
		return []getObligationsByFQNsRow{}, nil
	}
	namespaceFqnsJSON, err := json.Marshal(arg.NamespaceFqns)
	if err != nil {
		return nil, err
	}
	namesJSON, err := json.Marshal(arg.Names)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, `
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
    CAST(JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(od.metadata, '$.labels'), 'created_at', od.created_at,'updated_at', od.updated_at)) AS BLOB) as metadata,
    CAST(JSON_BUILD_OBJECT(
        'id', n.id,
        'name', n.name,
        'fqn', fqns.fqn
    ) AS BLOB) as namespace,
    CAST(
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
        ) AS BLOB
    ) as obligation_values
FROM
    obligation_definitions od
JOIN
    attribute_namespaces n on od.namespace_id = n.id
JOIN
    attribute_fqns fqns ON fqns.namespace_id = n.id AND fqns.attribute_id IS NULL AND fqns.value_id IS NULL
JOIN
    (SELECT ns.value as ns_fqn, nm.value as obl_name
     FROM json_each($1) AS ns
     JOIN json_each($2) AS nm ON ns.key = nm.key) as fqn_pairs
ON
    fqns.fqn = fqn_pairs.ns_fqn AND od.name = fqn_pairs.obl_name
LEFT JOIN
    obligation_values_standard ov on od.id = ov.obligation_definition_id
LEFT JOIN
    obligation_triggers_agg ota on ov.id = ota.obligation_value_id
GROUP BY
    od.id, n.id, fqns.fqn
`, string(namespaceFqnsJSON), string(namesJSON))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []getObligationsByFQNsRow{}
	for rows.Next() {
		var item getObligationsByFQNsRow
		if err := rows.Scan(&item.ID, &item.Name, &item.Metadata, &item.Namespace, &item.Values); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s sqliteQueries) getProviderConfig(ctx context.Context, arg getProviderConfigParams) (getProviderConfigRow, error) {
	var id interface{}
	if arg.ID.Valid {
		id = UUIDToString(arg.ID)
	}
	var name interface{}
	if arg.Name.Valid {
		name = arg.Name.String
	}
	var manager interface{}
	if arg.Manager.Valid {
		manager = arg.Manager.String
	}

	row := s.db.QueryRowContext(ctx, `
SELECT
    pc.id,
    pc.provider_name,
    pc.manager,
    CAST(pc.config AS BLOB) AS config,
    CAST(JSON_STRIP_NULLS(
        JSON_BUILD_OBJECT(
            'labels', json_extract(pc.metadata, '$.labels'),
            'created_at', pc.created_at,
            'updated_at', pc.updated_at
        )
    ) AS BLOB) AS metadata
FROM provider_config AS pc
WHERE ($1 IS NULL OR pc.id = $1)
  AND ($2 IS NULL OR pc.provider_name = $2)
  AND ($3 IS NULL OR pc.manager = $3)
`, id, name, manager)

	var out getProviderConfigRow
	if err := row.Scan(&out.ID, &out.ProviderName, &out.Manager, &out.Config, &out.Metadata); err != nil {
		return getProviderConfigRow{}, err
	}
	return out, nil
}

func (s sqliteQueries) getRegisteredResource(ctx context.Context, arg getRegisteredResourceParams) (getRegisteredResourceRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.GetRegisteredResourceParams](arg)
	if err != nil {
		return getRegisteredResourceRow{}, err
	}
	row := s.db.QueryRowContext(ctx, `
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
    ) as resource_values
FROM registered_resources r
LEFT JOIN registered_resource_values v ON v.registered_resource_id = r.id
WHERE
    ($1 IS NULL OR r.id = $1) AND
    ($2 IS NULL OR r.name = $2)
GROUP BY r.id
`, sqliteArg.ID, sqliteArg.Name)

	var res dbsqlite.GetRegisteredResourceRow
	if err := row.Scan(&res.ID, &res.Name, &res.Metadata, &res.Values); err != nil {
		return getRegisteredResourceRow{}, err
	}
	return convertStruct[getRegisteredResourceRow](res)
}

func (s sqliteQueries) getRegisteredResourceValue(ctx context.Context, arg getRegisteredResourceValueParams) (getRegisteredResourceValueRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.GetRegisteredResourceValueParams](arg)
	if err != nil {
		return getRegisteredResourceValueRow{}, err
	}
	res, err := s.q.GetRegisteredResourceValue(ctx, sqliteArg)
	if err != nil {
		return getRegisteredResourceValueRow{}, err
	}
	return convertStruct[getRegisteredResourceValueRow](res)
}

func (s sqliteQueries) getResourceMapping(ctx context.Context, id string) (getResourceMappingRow, error) {
	res, err := s.q.GetResourceMapping(ctx, id)
	if err != nil {
		return getResourceMappingRow{}, err
	}
	return convertStruct[getResourceMappingRow](res)
}

func (s sqliteQueries) getResourceMappingGroup(ctx context.Context, id string) (getResourceMappingGroupRow, error) {
	res, err := s.q.GetResourceMappingGroup(ctx, id)
	if err != nil {
		return getResourceMappingGroupRow{}, err
	}
	return convertStruct[getResourceMappingGroupRow](res)
}

func (s sqliteQueries) getSubjectConditionSet(ctx context.Context, id string) (getSubjectConditionSetRow, error) {
	res, err := s.q.GetSubjectConditionSet(ctx, id)
	if err != nil {
		return getSubjectConditionSetRow{}, err
	}
	return convertStruct[getSubjectConditionSetRow](res)
}

func (s sqliteQueries) getSubjectMapping(ctx context.Context, id string) (getSubjectMappingRow, error) {
	res, err := s.q.GetSubjectMapping(ctx, id)
	if err != nil {
		return getSubjectMappingRow{}, err
	}
	return convertStruct[getSubjectMappingRow](res)
}

func (s sqliteQueries) keyAccessServerExists(ctx context.Context, arg keyAccessServerExistsParams) (bool, error) {
	sqliteArg, err := convertStruct[dbsqlite.KeyAccessServerExistsParams](arg)
	if err != nil {
		return false, err
	}
	res, err := s.q.KeyAccessServerExists(ctx, sqliteArg)
	if err != nil {
		return false, err
	}
	return res, nil
}

func (s sqliteQueries) listActions(ctx context.Context, arg listActionsParams) ([]listActionsRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListActionsParams](arg)
	if err != nil {
		return nil, err
	}
	res, err := s.q.ListActions(ctx, sqliteArg)
	if err != nil {
		return nil, err
	}
	return convertSlice[listActionsRow](res)
}

func (s sqliteQueries) listAttributeValues(ctx context.Context, arg listAttributeValuesParams) ([]listAttributeValuesRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListAttributeValuesParams](arg)
	if err != nil {
		return nil, err
	}
	res, err := s.q.ListAttributeValues(ctx, sqliteArg)
	if err != nil {
		return nil, err
	}
	return convertSlice[listAttributeValuesRow](res)
}

func (s sqliteQueries) listAttributesByDefOrValueFqns(ctx context.Context, arg listAttributesByDefOrValueFqnsParams) ([]listAttributesByDefOrValueFqnsRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListAttributesByDefOrValueFqnsParams](arg)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
WITH target_definition AS (
    SELECT DISTINCT
        ad.id,
        ad.namespace_id,
        ad.name,
        ad.rule,
        ad.allow_traversal,
        ad.active,
        ad.values_order,
        ad.created_at,
        COALESCE(
            JSON_AGG(
                DISTINCT CASE
                    WHEN kas.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                        'id', kas.id,
                        'uri', kas.uri,
                        'name', kas.name,
                        'public_key', kas.public_key
                    )
                END
            ),
            '[]'
        ) AS grants,
        COALESCE(defk.keys, '[]') AS keys
    FROM attribute_fqns fqns
    INNER JOIN attribute_definitions ad ON fqns.attribute_id = ad.id
    LEFT JOIN attribute_definition_key_access_grants adkag ON ad.id = adkag.attribute_definition_id
    LEFT JOIN key_access_servers kas ON adkag.key_access_server_id = kas.id
    LEFT JOIN (
        SELECT
            k.definition_id,
            COALESCE(
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
                ),
                '[]'
            ) AS keys
        FROM attribute_definition_public_key_map k
        INNER JOIN key_access_server_keys kask ON k.key_access_server_key_id = kask.id
        INNER JOIN key_access_servers kas ON kask.key_access_server_id = kas.id
        GROUP BY k.definition_id
    ) defk ON ad.id = defk.definition_id
    WHERE fqns.fqn IN (SELECT fqns_vals.value FROM json_each($1) AS fqns_vals(value))
        AND ad.active = TRUE
    GROUP BY ad.id, ad.created_at, defk.keys
),
namespaces AS (
	SELECT
		n.id,
		JSON_BUILD_OBJECT(
			'id', n.id,
			'name', n.name,
			'active', n.active,
	        'fqn', fqns.fqn,
            'grants', COALESCE(
                JSON_AGG(
                    DISTINCT CASE
                        WHEN kas.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                            'id', kas.id,
                            'uri', kas.uri,
                            'name', kas.name,
                            'public_key', kas.public_key
                        )
                    END
                ),
                '[]'
            ),
            'kas_keys', COALESCE(nmp_keys.keys, '[]')
    	) AS namespace
	FROM target_definition td
	INNER JOIN attribute_namespaces n ON td.namespace_id = n.id
	INNER JOIN attribute_fqns fqns ON n.id = fqns.namespace_id
    LEFT JOIN attribute_namespace_key_access_grants ankag ON n.id = ankag.namespace_id
	LEFT JOIN key_access_servers kas ON ankag.key_access_server_id = kas.id
    LEFT JOIN (
        SELECT
            k.namespace_id,
            COALESCE(
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
                ),
                '[]'
            ) AS keys
        FROM attribute_namespace_public_key_map k
        INNER JOIN key_access_server_keys kask ON k.key_access_server_key_id = kask.id
        INNER JOIN key_access_servers kas ON kask.key_access_server_id = kas.id
        GROUP BY k.namespace_id
    ) nmp_keys ON n.id = nmp_keys.namespace_id
	WHERE n.active = TRUE
		AND (fqns.attribute_id IS NULL AND fqns.value_id IS NULL)
	GROUP BY n.id, fqns.fqn, nmp_keys.keys
),
value_grants AS (
	SELECT
		av.id,
        COALESCE(
            JSON_AGG(
                DISTINCT CASE
                    WHEN kas.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                        'id', kas.id,
                        'uri', kas.uri,
                        'name', kas.name,
                        'public_key', kas.public_key
                    )
                END
            ),
            '[]'
        ) AS grants
	FROM target_definition td
	LEFT JOIN attribute_values av on td.id = av.attribute_definition_id
	LEFT JOIN attribute_value_key_access_grants avkag ON av.id = avkag.attribute_value_id
	LEFT JOIN key_access_servers kas ON avkag.key_access_server_id = kas.id
	GROUP BY av.id
),
value_subject_mappings AS (
	SELECT
		av.id,
        COALESCE((
            SELECT JSON_AGG(subject_map)
            FROM (
                SELECT
                    JSON_BUILD_OBJECT(
                        'id', sm.id,
                        'actions', (
                            SELECT COALESCE(
                                JSON_AGG(
                                    CASE
                                        WHEN a.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                                            'id', a.id,
                                            'name', a.name
                                        )
                                    END
                                ),
                                '[]'
                            )
                            FROM subject_mapping_actions sma
                            LEFT JOIN actions a ON sma.action_id = a.id
                            WHERE sma.subject_mapping_id = sm.id
                        ),
                        'subject_condition_set', JSON_BUILD_OBJECT(
                            'id', scs.id,
                            'subject_sets', scs.condition
                        )
                    ) AS subject_map
                FROM subject_mappings sm
                LEFT JOIN subject_condition_set scs ON sm.subject_condition_set_id = scs.id
                WHERE sm.attribute_value_id = av.id
                ORDER BY sm.created_at, sm.rowid
            ) ordered_subject_mappings
        ), '[]') AS sub_maps
	FROM target_definition td
	LEFT JOIN attribute_values av ON td.id = av.attribute_definition_id
	GROUP BY av.id
),
value_resource_mappings AS (
    SELECT
        av.id,
        COALESCE(
            JSON_AGG(
                CASE
                    WHEN rm.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                        'id', rm.id,
                        'terms', rm.terms,
                        'group', CASE
                                    WHEN rm.group_id IS NULL THEN 'null'
                                    ELSE JSON_BUILD_OBJECT(
                                        'id', rmg.id,
                                        'name', rmg.name,
                                        'namespace_id', rmg.namespace_id
                                    )
                                 END
                    )
                END
            ),
            '[]'
        ) AS res_maps
    FROM target_definition td
    LEFT JOIN attribute_values av ON td.id = av.attribute_definition_id
    LEFT JOIN resource_mappings rm ON av.id = rm.attribute_value_id
    LEFT JOIN resource_mapping_groups rmg ON rm.group_id = rmg.id
    GROUP BY av.id
),
"values" AS (
    SELECT
		av.attribute_definition_id,
		JSON_AGG(
	        JSON_BUILD_OBJECT(
	            'id', av.id,
	            'value', av.value,
	            'active', av.active,
	            'fqn', fqns.fqn,
                'grants', COALESCE(avg.grants, '[]'),
	            'subject_mappings', COALESCE(avsm.sub_maps, '[]'),
                'resource_mappings', COALESCE(avrm.res_maps, '[]'),
                'kas_keys', COALESCE(value_keys.keys, '[]')
	        -- enforce order of values in response
	        ) ORDER BY COALESCE(CAST(vo.key AS INTEGER), 2147483647)
	    ) AS "values"
	FROM target_definition td
	LEFT JOIN attribute_values av ON td.id = av.attribute_definition_id
	LEFT JOIN attribute_fqns fqns ON av.id = fqns.value_id
    LEFT JOIN value_grants avg ON av.id = avg.id
	LEFT JOIN value_subject_mappings avsm ON av.id = avsm.id
    LEFT JOIN value_resource_mappings avrm ON av.id = avrm.id
    LEFT JOIN json_each(td.values_order) AS vo ON vo.value = av.id
    LEFT JOIN (
        SELECT
            k.value_id,
            COALESCE(
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
                ),
                '[]'
            ) AS keys
        FROM attribute_value_public_key_map k
        INNER JOIN key_access_server_keys kask ON k.key_access_server_key_id = kask.id
        INNER JOIN key_access_servers kas ON kask.key_access_server_id = kas.id
        GROUP BY k.value_id
    ) value_keys ON av.id = value_keys.value_id
	WHERE (av.active = TRUE OR $2 = TRUE)
	GROUP BY av.attribute_definition_id
)
SELECT
	td.id,
	td.name,
    td.rule,
    td.allow_traversal,
	td.active,
	n.namespace,
	fqns.fqn,
	"values"."values",
    CAST(COALESCE(td.grants, '[]') AS BLOB),
    CAST(COALESCE(td.keys, '[]') AS BLOB)
FROM target_definition td
INNER JOIN attribute_fqns fqns ON td.id = fqns.attribute_id
INNER JOIN namespaces n ON td.namespace_id = n.id
LEFT JOIN "values" ON td.id = "values".attribute_definition_id
WHERE fqns.value_id IS NULL
ORDER BY td.created_at DESC
`, sqliteArg.Fqns, sqliteArg.IncludeInactiveValues)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]dbsqlite.ListAttributesByDefOrValueFqnsRow, 0)
	for rows.Next() {
		var i dbsqlite.ListAttributesByDefOrValueFqnsRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Rule,
			&i.AllowTraversal,
			&i.Active,
			&i.Namespace,
			&i.Fqn,
			&i.Values,
			&i.Grants,
			&i.Keys,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return convertSlice[listAttributesByDefOrValueFqnsRow](items)
}

func (s sqliteQueries) listAttributesDetail(ctx context.Context, arg listAttributesDetailParams) ([]listAttributesDetailRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListAttributesDetailParams](arg)
	if err != nil {
		return nil, err
	}
	res, err := s.q.ListAttributesDetail(ctx, sqliteArg)
	if err != nil {
		return nil, err
	}
	return convertSlice[listAttributesDetailRow](res)
}

func (s sqliteQueries) listAttributesSummary(ctx context.Context, arg listAttributesSummaryParams) ([]listAttributesSummaryRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListAttributesSummaryParams](arg)
	if err != nil {
		return nil, err
	}
	res, err := s.q.ListAttributesSummary(ctx, sqliteArg)
	if err != nil {
		return nil, err
	}
	return convertSlice[listAttributesSummaryRow](res)
}

func (s sqliteQueries) listKeyAccessServerGrants(ctx context.Context, arg listKeyAccessServerGrantsParams) ([]listKeyAccessServerGrantsRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListKeyAccessServerGrantsParams](arg)
	if err != nil {
		return nil, err
	}
	res, err := s.q.ListKeyAccessServerGrants(ctx, sqliteArg)
	if err != nil {
		return nil, err
	}
	return convertSlice[listKeyAccessServerGrantsRow](res)
}

func (s sqliteQueries) listKeyAccessServers(ctx context.Context, arg listKeyAccessServersParams) ([]listKeyAccessServersRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListKeyAccessServersParams](arg)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
WITH counted AS (
    SELECT COUNT(kas.id) AS total
    FROM key_access_servers AS kas
)
SELECT kas.id,
    kas.uri,
    kas.public_key,
    kas.name AS kas_name,
    kas.source_type,
    CAST(JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(kas.metadata, '$.labels'), 'created_at', kas.created_at, 'updated_at', kas.updated_at)) AS BLOB) AS metadata,
    CAST(COALESCE(kask_keys.keys, JSON('[]')) AS BLOB) AS keys,
    counted.total
FROM key_access_servers AS kas
CROSS JOIN counted
LEFT JOIN (
        SELECT
            kask.key_access_server_id,
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
        FROM key_access_server_keys kask
        INNER JOIN key_access_servers kas ON kask.key_access_server_id = kas.id
        GROUP BY kask.key_access_server_id
    ) kask_keys ON kas.id = kask_keys.key_access_server_id
ORDER BY kas.created_at DESC
LIMIT $2
OFFSET $1
`, sqliteArg.Offset, sqliteArg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []listKeyAccessServersRow
	for rows.Next() {
		var res dbsqlite.ListKeyAccessServersRow
		if err := rows.Scan(
			&res.ID,
			&res.Uri,
			&res.PublicKey,
			&res.KasName,
			&res.SourceType,
			&res.Metadata,
			&res.Keys,
			&res.Total,
		); err != nil {
			return nil, err
		}
		converted, err := convertStruct[listKeyAccessServersRow](res)
		if err != nil {
			return nil, err
		}
		items = append(items, converted)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s sqliteQueries) listKeyMappings(ctx context.Context, arg listKeyMappingsParams) ([]listKeyMappingsRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListKeyMappingsParams](arg)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
WITH filtered_keys AS (
    SELECT
        kask.created_at,
        kask.rowid AS rowid,
        kask.id AS id,
        kask.key_id AS kid,
        kas.id AS kas_id,
        kas.uri AS kas_uri
    FROM key_access_server_keys kask
    INNER JOIN key_access_servers kas ON kask.key_access_server_id = kas.id
    WHERE (
        ($3 IS NOT NULL AND kask.id = $3)
        OR (
            $4 IS NOT NULL
            AND kask.key_id = $4
            AND (
                ($5 IS NOT NULL AND kas.id = $5)
                OR ($6 IS NOT NULL AND kas.name = $6)
                OR ($7 IS NOT NULL AND kas.uri = $7)
            )
        )
        OR (
            $3 IS NULL
            AND $4 IS NULL
        )
    )
),
keys_with_mappings AS (
    SELECT id
    FROM filtered_keys fk
    WHERE EXISTS (
        SELECT 1 FROM attribute_namespace_public_key_map anpm WHERE anpm.key_access_server_key_id = fk.id
    ) OR EXISTS (
        SELECT 1 FROM attribute_definition_public_key_map adpm WHERE adpm.key_access_server_key_id = fk.id
    ) OR EXISTS (
        SELECT 1 FROM attribute_value_public_key_map avpm WHERE avpm.key_access_server_key_id = fk.id
    )
),
keys_with_mappings_count AS (
    SELECT COUNT(*) AS total FROM keys_with_mappings
),
namespace_mappings AS (
    SELECT
        fk.id as key_id,
        JSON_AGG(
            CASE
                WHEN anpm.namespace_id IS NOT NULL THEN JSON_BUILD_OBJECT(
                    'id', anpm.namespace_id,
                    'fqn', fqns.fqn
                )
            END
        ) AS namespace_mappings
    FROM filtered_keys fk
    INNER JOIN attribute_namespace_public_key_map anpm ON fk.id = anpm.key_access_server_key_id
    INNER JOIN attribute_fqns fqns ON anpm.namespace_id = fqns.namespace_id AND fqns.attribute_id IS NULL AND fqns.value_id IS NULL
    GROUP BY fk.id
),
definition_mappings AS (
    SELECT
        fk.id as key_id,
        JSON_AGG(
            CASE
                WHEN adpm.definition_id IS NOT NULL THEN JSON_BUILD_OBJECT(
                    'id', adpm.definition_id,
                    'fqn', fqns.fqn
                )
            END
        ) AS definition_mappings
    FROM filtered_keys fk
    INNER JOIN attribute_definition_public_key_map adpm ON fk.id = adpm.key_access_server_key_id
    INNER JOIN attribute_fqns fqns ON adpm.definition_id = fqns.attribute_id AND fqns.value_id IS NULL
    GROUP BY fk.id
),
value_mappings AS (
    SELECT
        fk.id as key_id,
        JSON_AGG(
            CASE
                WHEN avpm.value_id IS NOT NULL THEN JSON_BUILD_OBJECT(
                    'id', avpm.value_id,
                    'fqn', fqns.fqn
                )
            END
        ) AS value_mappings
    FROM filtered_keys fk
    INNER JOIN attribute_value_public_key_map avpm ON fk.id = avpm.key_access_server_key_id
    INNER JOIN attribute_fqns fqns ON avpm.value_id = fqns.value_id
    GROUP BY fk.id
)
SELECT
    fk.kid,
    fk.kas_uri,
    CAST(COALESCE(nm.namespace_mappings, JSON('[]')) AS BLOB) AS namespace_mappings,
    CAST(COALESCE(dm.definition_mappings, JSON('[]')) AS BLOB) AS attribute_mappings,
    CAST(COALESCE(vm.value_mappings, JSON('[]')) AS BLOB) AS value_mappings,
    kwmc.total
FROM filtered_keys fk
INNER JOIN keys_with_mappings kwm ON fk.id = kwm.id
CROSS JOIN keys_with_mappings_count kwmc
LEFT JOIN namespace_mappings nm ON fk.id = nm.key_id
LEFT JOIN definition_mappings dm ON fk.id = dm.key_id
LEFT JOIN value_mappings vm ON fk.id = vm.key_id
ORDER BY fk.created_at DESC
       , fk.rowid DESC
LIMIT $2
OFFSET $1
`, sqliteArg.Offset, sqliteArg.Limit, sqliteArg.ID, sqliteArg.Kid, sqliteArg.KasID, sqliteArg.KasName, sqliteArg.KasUri)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []listKeyMappingsRow{}
	for rows.Next() {
		var kid sql.NullString
		var kasURI string
		var namespaceMappings []byte
		var attributeMappings []byte
		var valueMappings []byte
		var total int64
		if err := rows.Scan(&kid, &kasURI, &namespaceMappings, &attributeMappings, &valueMappings, &total); err != nil {
			return nil, err
		}
		if !kid.Valid {
			return nil, errors.New("unexpected NULL key id in key mappings")
		}
		items = append(items, listKeyMappingsRow{
			Kid:               kid.String,
			KasUri:            kasURI,
			NamespaceMappings: namespaceMappings,
			AttributeMappings: attributeMappings,
			ValueMappings:     valueMappings,
			Total:             total,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s sqliteQueries) listKeys(ctx context.Context, arg listKeysParams) ([]listKeysRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListKeysParams](arg)
	if err != nil {
		return nil, err
	}
	res, err := s.q.ListKeys(ctx, sqliteArg)
	if err != nil {
		return nil, err
	}
	return convertSlice[listKeysRow](res)
}

func (s sqliteQueries) listNamespaces(ctx context.Context, arg listNamespacesParams) ([]listNamespacesRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListNamespacesParams](arg)
	if err != nil {
		return nil, err
	}
	res, err := s.q.ListNamespaces(ctx, sqliteArg)
	if err != nil {
		return nil, err
	}
	return convertSlice[listNamespacesRow](res)
}

func (s sqliteQueries) listObligationTriggers(ctx context.Context, arg listObligationTriggersParams) ([]listObligationTriggersRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListObligationTriggersParams](arg)
	if err != nil {
		return nil, err
	}
	res, err := s.q.ListObligationTriggers(ctx, sqliteArg)
	if err != nil {
		return nil, err
	}
	return convertSlice[listObligationTriggersRow](res)
}

func (s sqliteQueries) listObligations(ctx context.Context, arg listObligationsParams) ([]listObligationsRow, error) {
	var namespaceID interface{}
	if arg.NamespaceID.Valid {
		namespaceID = UUIDToString(arg.NamespaceID)
	}
	var namespaceFqn interface{}
	if arg.NamespaceFqn.Valid {
		namespaceFqn = arg.NamespaceFqn.String
	}

	rows, err := s.db.QueryContext(ctx, `
WITH counted AS (
    SELECT COUNT(od.id) AS total
    FROM obligation_definitions od
    LEFT JOIN attribute_namespaces n ON od.namespace_id = n.id
    LEFT JOIN attribute_fqns fqns ON fqns.namespace_id = n.id AND fqns.attribute_id IS NULL AND fqns.value_id IS NULL
    WHERE
        ($1 IS NULL OR od.namespace_id = $1) AND
        ($2 IS NULL OR fqns.fqn = $2)
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
    CAST(JSON_BUILD_OBJECT(
        'id', n.id,
        'name', n.name,
        'fqn', fqns.fqn
    ) AS BLOB) as namespace,
    CAST(JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(od.metadata, '$.labels'), 'created_at', od.created_at,'updated_at', od.updated_at)) AS BLOB) as metadata,
    CAST(JSON_AGG(
        CASE
            WHEN ov.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                'id', ov.id,
                'value', ov.value,
                'triggers', COALESCE(ota.triggers, '[]')
            )
        END
    ) AS BLOB) as obligation_values,
    counted.total as total
FROM obligation_definitions od
JOIN attribute_namespaces n on od.namespace_id = n.id
LEFT JOIN attribute_fqns fqns ON fqns.namespace_id = n.id AND fqns.attribute_id IS NULL AND fqns.value_id IS NULL
LEFT JOIN obligation_values_standard ov on od.id = ov.obligation_definition_id
LEFT JOIN obligation_triggers_agg ota on ov.id = ota.obligation_value_id
CROSS JOIN counted
WHERE
    ($1 IS NULL OR od.namespace_id = $1) AND
    ($2 IS NULL OR fqns.fqn = $2)
GROUP BY od.id, n.id, fqns.fqn, counted.total
ORDER BY od.created_at DESC
LIMIT $3
OFFSET $4
`, namespaceID, namespaceFqn, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []listObligationsRow{}
	for rows.Next() {
		var item listObligationsRow
		if err := rows.Scan(&item.ID, &item.Name, &item.Namespace, &item.Metadata, &item.Values, &item.Total); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s sqliteQueries) listProviderConfigs(ctx context.Context, arg listProviderConfigsParams) ([]listProviderConfigsRow, error) {
	rows, err := s.db.QueryContext(ctx, `
WITH counted AS (
    SELECT COUNT(pc.id) AS total
    FROM provider_config pc
)
SELECT
    pc.id,
    pc.provider_name,
    pc.manager,
    CAST(pc.config AS BLOB) AS config,
    CAST(JSON_STRIP_NULLS(
        JSON_BUILD_OBJECT(
            'labels', json_extract(pc.metadata, '$.labels'),
            'created_at', pc.created_at,
            'updated_at', pc.updated_at
        )
    ) AS BLOB) AS metadata,
    counted.total
FROM provider_config AS pc
CROSS JOIN counted
ORDER BY pc.created_at DESC
LIMIT $1
OFFSET $2
`, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []listProviderConfigsRow{}
	for rows.Next() {
		var item listProviderConfigsRow
		if err := rows.Scan(&item.ID, &item.ProviderName, &item.Manager, &item.Config, &item.Metadata, &item.Total); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s sqliteQueries) listRegisteredResourceValues(ctx context.Context, arg listRegisteredResourceValuesParams) ([]listRegisteredResourceValuesRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListRegisteredResourceValuesParams](arg)
	if err != nil {
		return nil, err
	}
	res, err := s.q.ListRegisteredResourceValues(ctx, sqliteArg)
	if err != nil {
		return nil, err
	}
	return convertSlice[listRegisteredResourceValuesRow](res)
}

func (s sqliteQueries) listRegisteredResources(ctx context.Context, arg listRegisteredResourcesParams) ([]listRegisteredResourcesRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListRegisteredResourcesParams](arg)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
WITH counted AS (
    SELECT COUNT(id) AS total
    FROM registered_resources
)
SELECT
    r.id,
    r.name,
    JSON_STRIP_NULLS(JSON_BUILD_OBJECT('labels', json_extract(r.metadata, '$.labels'), 'created_at', r.created_at, 'updated_at', r.updated_at)) as metadata,
    JSON_AGG(
        CASE
            WHEN v.id IS NOT NULL THEN JSON_BUILD_OBJECT(
                'id', v.id,
                'value', v.value,
                'action_attribute_values', COALESCE(action_attrs.action_attribute_values, '[]')
            )
        END
    ) as resource_values,
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
        ) AS action_attribute_values
    FROM registered_resource_action_attribute_values rav
    LEFT JOIN actions a on rav.action_id = a.id
    LEFT JOIN attribute_values av on rav.attribute_value_id = av.id
    LEFT JOIN attribute_fqns fqns on av.id = fqns.value_id
    GROUP BY rav.registered_resource_value_id
) action_attrs ON action_attrs.registered_resource_value_id = v.id
GROUP BY r.id, counted.total
ORDER BY r.created_at DESC
LIMIT $1
OFFSET $2
`, sqliteArg.Limit, sqliteArg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]dbsqlite.ListRegisteredResourcesRow, 0)
	for rows.Next() {
		var i dbsqlite.ListRegisteredResourcesRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Metadata,
			&i.Values,
			&i.Total,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return convertSlice[listRegisteredResourcesRow](items)
}

func (s sqliteQueries) listResourceMappingGroups(ctx context.Context, arg listResourceMappingGroupsParams) ([]listResourceMappingGroupsRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListResourceMappingGroupsParams](arg)
	if err != nil {
		return nil, err
	}
	res, err := s.q.ListResourceMappingGroups(ctx, sqliteArg)
	if err != nil {
		return nil, err
	}
	return convertSlice[listResourceMappingGroupsRow](res)
}

func (s sqliteQueries) listResourceMappings(ctx context.Context, arg listResourceMappingsParams) ([]listResourceMappingsRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListResourceMappingsParams](arg)
	if err != nil {
		return nil, err
	}
	res, err := s.q.ListResourceMappings(ctx, sqliteArg)
	if err != nil {
		return nil, err
	}
	return convertSlice[listResourceMappingsRow](res)
}

func (s sqliteQueries) listResourceMappingsByFullyQualifiedGroup(ctx context.Context, arg listResourceMappingsByFullyQualifiedGroupParams) ([]listResourceMappingsByFullyQualifiedGroupRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListResourceMappingsByFullyQualifiedGroupParams](arg)
	if err != nil {
		return nil, err
	}
	res, err := s.q.ListResourceMappingsByFullyQualifiedGroup(ctx, sqliteArg)
	if err != nil {
		return nil, err
	}
	return convertSlice[listResourceMappingsByFullyQualifiedGroupRow](res)
}

func (s sqliteQueries) listSubjectConditionSets(ctx context.Context, arg listSubjectConditionSetsParams) ([]listSubjectConditionSetsRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListSubjectConditionSetsParams](arg)
	if err != nil {
		return nil, err
	}
	res, err := s.q.ListSubjectConditionSets(ctx, sqliteArg)
	if err != nil {
		return nil, err
	}
	return convertSlice[listSubjectConditionSetsRow](res)
}

func (s sqliteQueries) listSubjectMappings(ctx context.Context, arg listSubjectMappingsParams) ([]listSubjectMappingsRow, error) {
	sqliteArg, err := convertStruct[dbsqlite.ListSubjectMappingsParams](arg)
	if err != nil {
		return nil, err
	}
	res, err := s.q.ListSubjectMappings(ctx, sqliteArg)
	if err != nil {
		return nil, err
	}
	return convertSlice[listSubjectMappingsRow](res)
}

func (s sqliteQueries) matchSubjectMappings(ctx context.Context, selectors []string) ([]matchSubjectMappingsRow, error) {
	payloadBytes, err := json.Marshal(selectors)
	if err != nil {
		return nil, err
	}
	payload := string(payloadBytes)

	rows, err := s.db.QueryContext(ctx, `
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
        FROM json_each(COALESCE(json_extract(scs.condition, '$.subjectSets'), scs.condition)) AS ss
        CROSS JOIN json_each(json_extract(ss.value, '$.conditionGroups')) AS cg
        CROSS JOIN json_each(json_extract(cg.value, '$.conditions')) AS cond
        WHERE json_extract(cond.value, '$.subjectExternalSelectorValue') IN (
            SELECT selectors.value FROM json_each(?) AS selectors
        )
    )
GROUP BY
    sm.id,
    sa.standard_actions,
    sa.custom_actions,
    scs.id, scs.condition,
    av.id, av.value, av.active, fqns.fqn
`, payload)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]dbsqlite.MatchSubjectMappingsRow, 0)
	for rows.Next() {
		var i dbsqlite.MatchSubjectMappingsRow
		if err := rows.Scan(
			&i.ID,
			&i.StandardActions,
			&i.CustomActions,
			&i.SubjectConditionSet,
			&i.AttributeValue,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return convertSlice[matchSubjectMappingsRow](items)
}

func (s sqliteQueries) removeKeyAccessServerFromAttribute(ctx context.Context, arg removeKeyAccessServerFromAttributeParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.RemoveKeyAccessServerFromAttributeParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.q.RemoveKeyAccessServerFromAttribute(ctx, sqliteArg)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) removeKeyAccessServerFromAttributeValue(ctx context.Context, arg removeKeyAccessServerFromAttributeValueParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.RemoveKeyAccessServerFromAttributeValueParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.q.RemoveKeyAccessServerFromAttributeValue(ctx, sqliteArg)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) removeKeyAccessServerFromNamespace(ctx context.Context, arg removeKeyAccessServerFromNamespaceParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.RemoveKeyAccessServerFromNamespaceParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.q.RemoveKeyAccessServerFromNamespace(ctx, sqliteArg)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) removePublicKeyFromAttributeDefinition(ctx context.Context, arg removePublicKeyFromAttributeDefinitionParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.RemovePublicKeyFromAttributeDefinitionParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.q.RemovePublicKeyFromAttributeDefinition(ctx, sqliteArg)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) removePublicKeyFromAttributeValue(ctx context.Context, arg removePublicKeyFromAttributeValueParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.RemovePublicKeyFromAttributeValueParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.q.RemovePublicKeyFromAttributeValue(ctx, sqliteArg)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) removePublicKeyFromNamespace(ctx context.Context, arg removePublicKeyFromNamespaceParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.RemovePublicKeyFromNamespaceParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.q.RemovePublicKeyFromNamespace(ctx, sqliteArg)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) rotatePublicKeyForAttributeDefinition(ctx context.Context, arg rotatePublicKeyForAttributeDefinitionParams) ([]string, error) {
	sqliteArg, err := convertStruct[dbsqlite.RotatePublicKeyForAttributeDefinitionParams](arg)
	if err != nil {
		return nil, err
	}
	res, err := s.q.RotatePublicKeyForAttributeDefinition(ctx, sqliteArg)
	if err != nil {
		return nil, err
	}
	return stringsFromNullStrings(res)
}

func (s sqliteQueries) rotatePublicKeyForAttributeValue(ctx context.Context, arg rotatePublicKeyForAttributeValueParams) ([]string, error) {
	sqliteArg, err := convertStruct[dbsqlite.RotatePublicKeyForAttributeValueParams](arg)
	if err != nil {
		return nil, err
	}
	res, err := s.q.RotatePublicKeyForAttributeValue(ctx, sqliteArg)
	if err != nil {
		return nil, err
	}
	return stringsFromNullStrings(res)
}

func (s sqliteQueries) rotatePublicKeyForNamespace(ctx context.Context, arg rotatePublicKeyForNamespaceParams) ([]string, error) {
	sqliteArg, err := convertStruct[dbsqlite.RotatePublicKeyForNamespaceParams](arg)
	if err != nil {
		return nil, err
	}
	res, err := s.q.RotatePublicKeyForNamespace(ctx, sqliteArg)
	if err != nil {
		return nil, err
	}
	return stringsFromNullStrings(res)
}

func stringsFromNullStrings(res []sql.NullString) ([]string, error) {
	out := make([]string, 0, len(res))
	for _, item := range res {
		if !item.Valid {
			return nil, errors.New("unexpected NULL string in result set")
		}
		out = append(out, item.String)
	}
	return out, nil
}

func (s sqliteQueries) setBaseKey(ctx context.Context, keyAccessServerKeyID pgtype.UUID) (int64, error) {
	sqliteArg := sql.NullString{
		String: UUIDToString(keyAccessServerKeyID),
		Valid:  keyAccessServerKeyID.Valid,
	}
	updateRes, err := s.db.ExecContext(ctx, `UPDATE base_keys SET key_access_server_key_id = $1`, sqliteArg)
	if err != nil {
		return 0, err
	}
	updated, err := updateRes.RowsAffected()
	if err != nil {
		return 0, err
	}
	if updated > 0 {
		return updated, nil
	}
	return s.q.SetBaseKey(ctx, sqliteArg)
}

func (s sqliteQueries) updateAttribute(ctx context.Context, arg updateAttributeParams) (int64, error) {
	if arg.Rule.Valid && !isValidAttributeDefinitionRule(arg.Rule.AttributeDefinitionRule) {
		return 0, pkgdb.ErrEnumValueInvalid
	}
	sqliteArg, err := convertStruct[dbsqlite.UpdateAttributeParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.db.ExecContext(ctx, `
UPDATE attribute_definitions
SET
    name = COALESCE($2, name),
    rule = COALESCE($3, rule),
    values_order = COALESCE($4, values_order),
    metadata = COALESCE($5, metadata),
    active = COALESCE($6, active),
    allow_traversal = COALESCE($7, allow_traversal),
    updated_at = STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now', '+0.001 seconds')
WHERE id = $1
`, sqliteArg.ID, sqliteArg.Name, sqliteArg.Rule, sqliteArg.ValuesOrder, sqliteArg.Metadata, sqliteArg.Active, sqliteArg.AllowTraversal)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s sqliteQueries) updateAttributeValue(ctx context.Context, arg updateAttributeValueParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.UpdateAttributeValueParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.db.ExecContext(ctx, `
UPDATE attribute_values
SET
    value = COALESCE($2, value),
    active = COALESCE($3, active),
    metadata = COALESCE($4, metadata),
    updated_at = STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now', '+0.001 seconds')
WHERE id = $1
`, sqliteArg.ID, sqliteArg.Value, sqliteArg.Active, sqliteArg.Metadata)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s sqliteQueries) updateCustomAction(ctx context.Context, arg updateCustomActionParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.UpdateCustomActionParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.q.UpdateCustomAction(ctx, sqliteArg)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) updateKey(ctx context.Context, arg updateKeyParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.UpdateKeyParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.q.UpdateKey(ctx, sqliteArg)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) updateKeyAccessServer(ctx context.Context, arg updateKeyAccessServerParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.UpdateKeyAccessServerParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.q.UpdateKeyAccessServer(ctx, sqliteArg)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) updateNamespace(ctx context.Context, arg updateNamespaceParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.UpdateNamespaceParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.db.ExecContext(ctx, `
UPDATE attribute_namespaces
SET
    name = COALESCE($2, name),
    active = COALESCE($3, active),
    metadata = COALESCE($4, metadata),
    updated_at = STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now', '+0.001 seconds')
WHERE id = $1
`, sqliteArg.ID, sqliteArg.Name, sqliteArg.Active, sqliteArg.Metadata)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func (s sqliteQueries) updateObligation(ctx context.Context, arg updateObligationParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.UpdateObligationParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.q.UpdateObligation(ctx, sqliteArg)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) updateObligationValue(ctx context.Context, arg updateObligationValueParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.UpdateObligationValueParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.q.UpdateObligationValue(ctx, sqliteArg)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) updateProviderConfig(ctx context.Context, arg updateProviderConfigParams) (int64, error) {
	if arg.Config != nil && !json.Valid(arg.Config) {
		return 0, pkgdb.ErrEnumValueInvalid
	}
	sqliteArg, err := convertStruct[dbsqlite.UpdateProviderConfigParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.q.UpdateProviderConfig(ctx, sqliteArg)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) updateRegisteredResource(ctx context.Context, arg updateRegisteredResourceParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.UpdateRegisteredResourceParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.q.UpdateRegisteredResource(ctx, sqliteArg)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) updateRegisteredResourceValue(ctx context.Context, arg updateRegisteredResourceValueParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.UpdateRegisteredResourceValueParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.q.UpdateRegisteredResourceValue(ctx, sqliteArg)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) updateResourceMapping(ctx context.Context, arg updateResourceMappingParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.UpdateResourceMappingParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.db.ExecContext(ctx, `
UPDATE resource_mappings
SET
    attribute_value_id = COALESCE($1, attribute_value_id),
    terms = COALESCE($2, terms),
    metadata = COALESCE($3, metadata),
    group_id = COALESCE($4, group_id),
    updated_at = STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now', '+0.001 seconds')
WHERE id = $5
`, sqliteArg.AttributeValueID, sqliteArg.Terms, sqliteArg.Metadata, sqliteArg.GroupID, sqliteArg.ID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s sqliteQueries) updateResourceMappingGroup(ctx context.Context, arg updateResourceMappingGroupParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.UpdateResourceMappingGroupParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.db.ExecContext(ctx, `
UPDATE resource_mapping_groups
SET
    namespace_id = COALESCE($1, namespace_id),
    name = COALESCE($2, name),
    metadata = COALESCE($3, metadata),
    updated_at = STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now', '+0.001 seconds')
WHERE id = $4
`, sqliteArg.NamespaceID, sqliteArg.Name, sqliteArg.Metadata, sqliteArg.ID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (s sqliteQueries) updateSubjectConditionSet(ctx context.Context, arg updateSubjectConditionSetParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.UpdateSubjectConditionSetParams](arg)
	if err != nil {
		return 0, err
	}
	res, err := s.q.UpdateSubjectConditionSet(ctx, sqliteArg)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (s sqliteQueries) updateSubjectMapping(ctx context.Context, arg updateSubjectMappingParams) (int64, error) {
	sqliteArg, err := convertStruct[dbsqlite.UpdateSubjectMappingParams](arg)
	if err != nil {
		return 0, err
	}
	var rowsAffected int64
	err = s.withSQLiteTx(ctx, func(d dbsqlite.DBTX) error {
		res, execErr := d.ExecContext(ctx, `
UPDATE subject_mappings
SET
    metadata = COALESCE($1, metadata),
    subject_condition_set_id = COALESCE($2, subject_condition_set_id),
    updated_at = STRFTIME('%Y-%m-%dT%H:%M:%fZ', 'now', '+0.001 seconds')
WHERE id = $3
`, sqliteArg.Metadata, sqliteArg.SubjectConditionSetID, sqliteArg.ID)
		if execErr != nil {
			return fmt.Errorf("update subject mapping: %w", execErr)
		}
		rows, rowsErr := res.RowsAffected()
		if rowsErr != nil {
			return fmt.Errorf("update subject mapping rows: %w", rowsErr)
		}
		rowsAffected = rows

		if arg.ActionIds == nil {
			return nil
		}

		actionIDsJSON, marshalErr := json.Marshal(arg.ActionIds)
		if marshalErr != nil {
			return marshalErr
		}

		if _, execErr = d.ExecContext(ctx, `
DELETE FROM subject_mapping_actions
WHERE subject_mapping_id = $1
  AND action_id NOT IN (SELECT value FROM json_each($2))
`, sqliteArg.ID, actionIDsJSON); execErr != nil {
			return fmt.Errorf("delete subject mapping actions: %w", execErr)
		}

		_, execErr = d.ExecContext(ctx, `
INSERT INTO subject_mapping_actions (subject_mapping_id, action_id)
SELECT $1, value FROM json_each($2)
WHERE NOT EXISTS (
    SELECT 1
    FROM subject_mapping_actions
    WHERE subject_mapping_id = $1 AND action_id = value
)
`, sqliteArg.ID, actionIDsJSON)
		if execErr != nil {
			return fmt.Errorf("insert subject mapping actions: %w", execErr)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func (s sqliteQueries) upsertAttributeDefinitionFqn(ctx context.Context, attributeID string) ([]upsertAttributeDefinitionFqnRow, error) {
	res, err := s.q.UpsertAttributeDefinitionFqn(ctx, attributeID)
	if err != nil {
		return nil, err
	}
	return convertSlice[upsertAttributeDefinitionFqnRow](res)
}

func (s sqliteQueries) upsertAttributeNamespaceFqn(ctx context.Context, namespaceID string) ([]upsertAttributeNamespaceFqnRow, error) {
	res, err := s.q.UpsertAttributeNamespaceFqn(ctx, namespaceID)
	if err != nil {
		return nil, err
	}
	return convertSlice[upsertAttributeNamespaceFqnRow](res)
}

func (s sqliteQueries) upsertAttributeValueFqn(ctx context.Context, valueID string) ([]upsertAttributeValueFqnRow, error) {
	res, err := s.q.UpsertAttributeValueFqn(ctx, valueID)
	if err != nil {
		return nil, err
	}
	return convertSlice[upsertAttributeValueFqnRow](res)
}
