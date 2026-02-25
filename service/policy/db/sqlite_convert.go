package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sqlc-dev/pqtype"
)

var (
	typeNullString    = reflect.TypeOf(sql.NullString{})
	typeNullBool      = reflect.TypeOf(sql.NullBool{})
	typeNullInt32     = reflect.TypeOf(sql.NullInt32{})
	typeNullTime      = reflect.TypeOf(sql.NullTime{})
	typePGText        = reflect.TypeOf(pgtype.Text{})
	typePGUUID        = reflect.TypeOf(pgtype.UUID{})
	typePGBool        = reflect.TypeOf(pgtype.Bool{})
	typePGInt4        = reflect.TypeOf(pgtype.Int4{})
	typePGTimestamp   = reflect.TypeOf(pgtype.Timestamp{})
	typePGTimestamptz = reflect.TypeOf(pgtype.Timestamptz{})
	typeNullAttrRule  = reflect.TypeOf(NullAttributeDefinitionRule{})
	typeBytes         = reflect.TypeOf([]byte(nil))
	typeRawMessage    = reflect.TypeOf(json.RawMessage{})
	typePQNullRaw     = reflect.TypeOf(pqtype.NullRawMessage{})
	typeInterface     = reflect.TypeOf((*interface{})(nil)).Elem()
)

func convertStruct[Dst any](src any) (Dst, error) {
	var dst Dst
	dv := reflect.ValueOf(&dst).Elem()
	sv := reflect.ValueOf(src)
	if sv.Kind() == reflect.Pointer {
		if sv.IsNil() {
			return dst, nil
		}
		sv = sv.Elem()
	}
	if sv.Kind() != reflect.Struct {
		return dst, fmt.Errorf("convertStruct expects struct, got %s", sv.Kind())
	}
	for i := 0; i < dv.NumField(); i++ {
		df := dv.Type().Field(i)
		sf := sv.FieldByName(df.Name)
		if !sf.IsValid() {
			continue
		}
		val, err := convertValue(df.Type, sf)
		if err != nil {
			return dst, err
		}
		if val.IsValid() && dv.Field(i).CanSet() {
			dv.Field(i).Set(val)
		}
	}
	return dst, nil
}

func convertSlice[Dst any, Src any](src []Src) ([]Dst, error) {
	out := make([]Dst, 0, len(src))
	for _, item := range src {
		converted, err := convertStruct[Dst](item)
		if err != nil {
			return nil, err
		}
		out = append(out, converted)
	}
	return out, nil
}

//nolint:nestif // Conversion rules intentionally branch by destination/source type pair.
func convertValue(dstType reflect.Type, src reflect.Value) (reflect.Value, error) {
	if !src.IsValid() {
		return reflect.Zero(dstType), nil
	}
	if src.Kind() == reflect.Pointer {
		if src.IsNil() {
			return reflect.Zero(dstType), nil
		}
		src = src.Elem()
	}
	if dstType == typeInterface {
		switch v := src.Interface().(type) {
		case pgtype.UUID:
			if !v.Valid {
				return reflect.Zero(dstType), nil
			}
			return reflect.ValueOf(UUIDToString(v)), nil
		case pgtype.Text:
			if !v.Valid {
				return reflect.Zero(dstType), nil
			}
			return reflect.ValueOf(v.String), nil
		case pgtype.Bool:
			if !v.Valid {
				return reflect.Zero(dstType), nil
			}
			return reflect.ValueOf(v.Bool), nil
		case pgtype.Int4:
			if !v.Valid {
				return reflect.Zero(dstType), nil
			}
			return reflect.ValueOf(v.Int32), nil
		case sql.NullString:
			if !v.Valid {
				return reflect.Zero(dstType), nil
			}
			return reflect.ValueOf(v.String), nil
		case sql.NullBool:
			if !v.Valid {
				return reflect.Zero(dstType), nil
			}
			return reflect.ValueOf(v.Bool), nil
		case sql.NullInt32:
			if !v.Valid {
				return reflect.Zero(dstType), nil
			}
			return reflect.ValueOf(v.Int32), nil
		case sql.NullTime:
			if !v.Valid {
				return reflect.Zero(dstType), nil
			}
			return reflect.ValueOf(v.Time), nil
		case []string:
			b, err := json.Marshal(v)
			if err != nil {
				return reflect.ValueOf(src.Interface()), err
			}
			return reflect.ValueOf(json.RawMessage(b)), nil
		case []byte:
			if json.Valid(v) {
				return reflect.ValueOf(json.RawMessage(v)), nil
			}
		case string:
			if json.Valid([]byte(v)) {
				return reflect.ValueOf(json.RawMessage(v)), nil
			}
		default:
			return reflect.ValueOf(src.Interface()), nil
		}
	}
	if src.Type().AssignableTo(dstType) {
		return src.Convert(dstType), nil
	}
	if dstType.Kind() == reflect.String {
		switch v := src.Interface().(type) {
		case sql.NullString:
			if !v.Valid {
				return reflect.Zero(dstType), nil
			}
			return reflect.ValueOf(v.String).Convert(dstType), nil
		case []byte:
			return reflect.ValueOf(string(v)).Convert(dstType), nil
		}
	}
	if dstType.Kind() == reflect.Bool {
		switch v := src.Interface().(type) {
		case sql.NullBool:
			if !v.Valid {
				return reflect.ValueOf(false), nil
			}
			return reflect.ValueOf(v.Bool), nil
		case pgtype.Bool:
			if !v.Valid {
				return reflect.ValueOf(false), nil
			}
			return reflect.ValueOf(v.Bool), nil
		}
	}
	if dstType.Kind() == reflect.Slice && dstType.Elem().Kind() == reflect.String {
		switch v := src.Interface().(type) {
		case []string:
			return reflect.ValueOf(v), nil
		case string:
			return stringSliceValue(v)
		case []byte:
			return stringSliceValue(string(v))
		case sql.NullString:
			if !v.Valid {
				return reflect.Zero(dstType), nil
			}
			return stringSliceValue(v.String)
		}
	}
	if dstType.Kind() == reflect.Int32 {
		switch v := src.Interface().(type) {
		case sql.NullInt32:
			if !v.Valid {
				return reflect.ValueOf(int32(0)).Convert(dstType), nil
			}
			return reflect.ValueOf(v.Int32).Convert(dstType), nil
		case pgtype.Int4:
			if !v.Valid {
				return reflect.ValueOf(int32(0)).Convert(dstType), nil
			}
			return reflect.ValueOf(v.Int32).Convert(dstType), nil
		}
	}

	switch dstType {
	case typeNullString:
		v, err := toNullString(src)
		return reflect.ValueOf(v), err
	case typeNullBool:
		v, err := toNullBool(src)
		return reflect.ValueOf(v), err
	case typeNullInt32:
		v, err := toNullInt32(src)
		return reflect.ValueOf(v), err
	case typeNullTime:
		v, err := toNullTime(src)
		return reflect.ValueOf(v), err
	case typePGText:
		v, err := toPGText(src)
		return reflect.ValueOf(v), err
	case typePGUUID:
		v, err := toPGUUID(src)
		return reflect.ValueOf(v), err
	case typePGBool:
		v, err := toPGBool(src)
		return reflect.ValueOf(v), err
	case typePGInt4:
		v, err := toPGInt4(src)
		return reflect.ValueOf(v), err
	case typePGTimestamp:
		v, err := toPGTimestamp(src)
		return reflect.ValueOf(v), err
	case typePGTimestamptz:
		v, err := toPGTimestamptz(src)
		return reflect.ValueOf(v), err
	case typeNullAttrRule:
		v, err := toNullAttributeDefinitionRule(src)
		return reflect.ValueOf(v), err
	case typeBytes:
		v, err := toBytes(src)
		return reflect.ValueOf(v), err
	case typeRawMessage:
		v, err := toRawMessage(src)
		return reflect.ValueOf(v), err
	case typePQNullRaw:
		v, err := toPQNullRaw(src)
		return reflect.ValueOf(v), err
	default:
		// handle simple numeric conversions
		if dstType.Kind() == reflect.Int32 && src.Kind() == reflect.Int32 {
			return src.Convert(dstType), nil
		}
		if dstType.Kind() == reflect.Bool && src.Kind() == reflect.Bool {
			return src.Convert(dstType), nil
		}
		if dstType.Kind() == reflect.String && src.Kind() == reflect.String {
			return src.Convert(dstType), nil
		}
	}
	return reflect.Zero(dstType), fmt.Errorf("unsupported conversion from %s to %s", src.Type(), dstType)
}

func stringSliceValue(src string) (reflect.Value, error) {
	if src == "" {
		return reflect.ValueOf([]string{}), nil
	}
	var out []string
	if err := json.Unmarshal([]byte(src), &out); err != nil {
		return reflect.Zero(reflect.TypeOf(out)), fmt.Errorf("decode string slice json: %w", err)
	}
	return reflect.ValueOf(out), nil
}

func toNullString(src reflect.Value) (sql.NullString, error) {
	switch v := src.Interface().(type) {
	case string:
		return sql.NullString{String: v, Valid: v != ""}, nil
	case []byte:
		s := string(v)
		return sql.NullString{String: s, Valid: s != ""}, nil
	case sql.NullString:
		return v, nil
	case pgtype.Text:
		return sql.NullString{String: v.String, Valid: v.Valid}, nil
	case pgtype.UUID:
		if !v.Valid {
			return sql.NullString{}, nil
		}
		return sql.NullString{String: UUIDToString(v), Valid: true}, nil
	case AttributeDefinitionRule:
		if v == "" {
			return sql.NullString{}, nil
		}
		return sql.NullString{String: string(v), Valid: true}, nil
	case NullAttributeDefinitionRule:
		if !v.Valid {
			return sql.NullString{}, nil
		}
		return sql.NullString{String: string(v.AttributeDefinitionRule), Valid: true}, nil
	case []string:
		if v == nil {
			return sql.NullString{}, nil
		}
		encoded, err := json.Marshal(v)
		if err != nil {
			return sql.NullString{}, fmt.Errorf("encode string slice: %w", err)
		}
		return sql.NullString{String: string(encoded), Valid: true}, nil
	default:
		if src.Kind() == reflect.String {
			s := src.String()
			return sql.NullString{String: s, Valid: s != ""}, nil
		}
	}
	return sql.NullString{}, fmt.Errorf("unsupported null string source: %T", src.Interface())
}

func toNullAttributeDefinitionRule(src reflect.Value) (NullAttributeDefinitionRule, error) {
	switch v := src.Interface().(type) {
	case NullAttributeDefinitionRule:
		return v, nil
	case AttributeDefinitionRule:
		return NullAttributeDefinitionRule{AttributeDefinitionRule: v, Valid: true}, nil
	case string:
		if v == "" {
			return NullAttributeDefinitionRule{}, nil
		}
		return NullAttributeDefinitionRule{AttributeDefinitionRule: AttributeDefinitionRule(v), Valid: true}, nil
	case []byte:
		s := string(v)
		if s == "" {
			return NullAttributeDefinitionRule{}, nil
		}
		return NullAttributeDefinitionRule{AttributeDefinitionRule: AttributeDefinitionRule(s), Valid: true}, nil
	case sql.NullString:
		if !v.Valid {
			return NullAttributeDefinitionRule{}, nil
		}
		return NullAttributeDefinitionRule{AttributeDefinitionRule: AttributeDefinitionRule(v.String), Valid: true}, nil
	default:
		if src.Kind() == reflect.String {
			s := src.String()
			if s == "" {
				return NullAttributeDefinitionRule{}, nil
			}
			return NullAttributeDefinitionRule{AttributeDefinitionRule: AttributeDefinitionRule(s), Valid: true}, nil
		}
	}
	return NullAttributeDefinitionRule{}, fmt.Errorf("unsupported null attribute rule source: %T", src.Interface())
}

func toNullBool(src reflect.Value) (sql.NullBool, error) {
	switch v := src.Interface().(type) {
	case bool:
		return sql.NullBool{Bool: v, Valid: true}, nil
	case sql.NullBool:
		return v, nil
	case pgtype.Bool:
		return sql.NullBool{Bool: v.Bool, Valid: v.Valid}, nil
	default:
		if src.Kind() == reflect.Bool {
			return sql.NullBool{Bool: src.Bool(), Valid: true}, nil
		}
	}
	return sql.NullBool{}, fmt.Errorf("unsupported null bool source: %T", src.Interface())
}

func toNullInt32(src reflect.Value) (sql.NullInt32, error) {
	switch v := src.Interface().(type) {
	case int32:
		return sql.NullInt32{Int32: v, Valid: true}, nil
	case sql.NullInt32:
		return v, nil
	case pgtype.Int4:
		return sql.NullInt32{Int32: v.Int32, Valid: v.Valid}, nil
	default:
		if src.Kind() == reflect.Int32 {
			return sql.NullInt32{Int32: int32(src.Int()), Valid: true}, nil
		}
	}
	return sql.NullInt32{}, fmt.Errorf("unsupported null int32 source: %T", src.Interface())
}

func toNullTime(src reflect.Value) (sql.NullTime, error) {
	switch v := src.Interface().(type) {
	case time.Time:
		return sql.NullTime{Time: v, Valid: true}, nil
	case sql.NullTime:
		return v, nil
	case pgtype.Timestamptz:
		return sql.NullTime{Time: v.Time, Valid: v.Valid}, nil
	case pgtype.Timestamp:
		return sql.NullTime{Time: v.Time, Valid: v.Valid}, nil
	default:
		if src.Kind() == reflect.Struct && src.Type() == reflect.TypeOf(time.Time{}) {
			if value, ok := src.Interface().(time.Time); ok {
				return sql.NullTime{Time: value, Valid: true}, nil
			}
		}
	}
	return sql.NullTime{}, fmt.Errorf("unsupported null time source: %T", src.Interface())
}

func toPGText(src reflect.Value) (pgtype.Text, error) {
	switch v := src.Interface().(type) {
	case string:
		return pgtypeText(v), nil
	case sql.NullString:
		if !v.Valid {
			return pgtype.Text{}, nil
		}
		return pgtypeText(v.String), nil
	case pgtype.Text:
		return v, nil
	default:
		if src.Kind() == reflect.String {
			return pgtypeText(src.String()), nil
		}
	}
	return pgtype.Text{}, fmt.Errorf("unsupported pgtype.Text source: %T", src.Interface())
}

func toPGUUID(src reflect.Value) (pgtype.UUID, error) {
	switch v := src.Interface().(type) {
	case string:
		return pgtypeUUID(v), nil
	case sql.NullString:
		if !v.Valid {
			return pgtype.UUID{}, nil
		}
		return pgtypeUUID(v.String), nil
	case pgtype.UUID:
		return v, nil
	default:
		if src.Kind() == reflect.String {
			return pgtypeUUID(src.String()), nil
		}
	}
	return pgtype.UUID{}, fmt.Errorf("unsupported pgtype.UUID source: %T", src.Interface())
}

func toPGBool(src reflect.Value) (pgtype.Bool, error) {
	switch v := src.Interface().(type) {
	case bool:
		return pgtypeBool(v), nil
	case sql.NullBool:
		if !v.Valid {
			return pgtype.Bool{}, nil
		}
		return pgtypeBool(v.Bool), nil
	case pgtype.Bool:
		return v, nil
	default:
		if src.Kind() == reflect.Bool {
			return pgtypeBool(src.Bool()), nil
		}
	}
	return pgtype.Bool{}, fmt.Errorf("unsupported pgtype.Bool source: %T", src.Interface())
}

func toPGInt4(src reflect.Value) (pgtype.Int4, error) {
	switch v := src.Interface().(type) {
	case int32:
		return pgtypeInt4(v, true), nil
	case sql.NullInt32:
		if !v.Valid {
			return pgtype.Int4{}, nil
		}
		return pgtypeInt4(v.Int32, true), nil
	case pgtype.Int4:
		return v, nil
	default:
		if src.Kind() == reflect.Int32 {
			return pgtypeInt4(int32(src.Int()), true), nil
		}
	}
	return pgtype.Int4{}, fmt.Errorf("unsupported pgtype.Int4 source: %T", src.Interface())
}

func toPGTimestamp(src reflect.Value) (pgtype.Timestamp, error) {
	switch v := src.Interface().(type) {
	case time.Time:
		return pgtype.Timestamp{Time: v, Valid: true}, nil
	case sql.NullTime:
		if !v.Valid {
			return pgtype.Timestamp{}, nil
		}
		return pgtype.Timestamp{Time: v.Time, Valid: true}, nil
	case pgtype.Timestamp:
		return v, nil
	default:
		if src.Kind() == reflect.Struct && src.Type() == reflect.TypeOf(time.Time{}) {
			if value, ok := src.Interface().(time.Time); ok {
				return pgtype.Timestamp{Time: value, Valid: true}, nil
			}
		}
	}
	return pgtype.Timestamp{}, fmt.Errorf("unsupported pgtype.Timestamp source: %T", src.Interface())
}

func toPGTimestamptz(src reflect.Value) (pgtype.Timestamptz, error) {
	switch v := src.Interface().(type) {
	case time.Time:
		return pgtype.Timestamptz{Time: v, Valid: true}, nil
	case sql.NullTime:
		if !v.Valid {
			return pgtype.Timestamptz{}, nil
		}
		return pgtype.Timestamptz{Time: v.Time, Valid: true}, nil
	case pgtype.Timestamptz:
		return v, nil
	default:
		if src.Kind() == reflect.Struct && src.Type() == reflect.TypeOf(time.Time{}) {
			if value, ok := src.Interface().(time.Time); ok {
				return pgtype.Timestamptz{Time: value, Valid: true}, nil
			}
		}
	}
	return pgtype.Timestamptz{}, fmt.Errorf("unsupported pgtype.Timestamptz source: %T", src.Interface())
}

func toBytes(src reflect.Value) ([]byte, error) {
	switch v := src.Interface().(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case []string:
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return b, nil
	case sql.NullString:
		if !v.Valid {
			return nil, nil
		}
		return []byte(v.String), nil
	case json.RawMessage:
		return []byte(v), nil
	case pqtype.NullRawMessage:
		if !v.Valid {
			return nil, nil
		}
		return []byte(v.RawMessage), nil
	default:
		if src.Kind() == reflect.String {
			return []byte(src.String()), nil
		}
	}
	return nil, fmt.Errorf("unsupported []byte source: %T", src.Interface())
}

func toRawMessage(src reflect.Value) (json.RawMessage, error) {
	switch v := src.Interface().(type) {
	case json.RawMessage:
		return v, nil
	case []byte:
		return json.RawMessage(v), nil
	case string:
		return json.RawMessage(v), nil
	case []string:
		b, err := json.Marshal(v)
		return json.RawMessage(b), err
	case sql.NullString:
		if !v.Valid {
			return nil, nil
		}
		return json.RawMessage(v.String), nil
	case pqtype.NullRawMessage:
		if !v.Valid {
			return nil, nil
		}
		return v.RawMessage, nil
	default:
		if src.Kind() == reflect.Slice && src.Type().Elem().Kind() == reflect.String {
			b, err := json.Marshal(src.Interface())
			return json.RawMessage(b), err
		}
	}
	return nil, fmt.Errorf("unsupported raw message source: %T", src.Interface())
}

func toPQNullRaw(src reflect.Value) (pqtype.NullRawMessage, error) {
	switch v := src.Interface().(type) {
	case pqtype.NullRawMessage:
		return v, nil
	case []byte:
		return pqtype.NullRawMessage{RawMessage: v, Valid: len(v) > 0}, nil
	case []string:
		if v == nil {
			return pqtype.NullRawMessage{}, nil
		}
		b, err := json.Marshal(v)
		if err != nil {
			return pqtype.NullRawMessage{}, err
		}
		return pqtype.NullRawMessage{RawMessage: b, Valid: true}, nil
	case json.RawMessage:
		return pqtype.NullRawMessage{RawMessage: []byte(v), Valid: len(v) > 0}, nil
	case string:
		return pqtype.NullRawMessage{RawMessage: []byte(v), Valid: v != ""}, nil
	case sql.NullString:
		if !v.Valid {
			return pqtype.NullRawMessage{}, nil
		}
		return pqtype.NullRawMessage{RawMessage: []byte(v.String), Valid: true}, nil
	default:
		if src.Kind() == reflect.String {
			s := src.String()
			return pqtype.NullRawMessage{RawMessage: []byte(s), Valid: s != ""}, nil
		}
	}
	return pqtype.NullRawMessage{}, fmt.Errorf("unsupported NullRawMessage source: %T", src.Interface())
}
