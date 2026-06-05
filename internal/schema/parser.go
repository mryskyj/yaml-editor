package schema

import (
	"fmt"
	"reflect"
	"strings"
)

// Parse converts a Go struct value or type into a YAML editor schema field.
func Parse(v any) (*Field, error) {
	if v == nil {
		return nil, fmt.Errorf("schema parse failed: nil value")
	}

	t, ok := typeFrom(v)
	if !ok {
		return nil, fmt.Errorf("schema parse failed: unsupported input %T", v)
	}

	t = dereference(t)
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("schema parse failed: root must be struct, got %s", t.Kind())
	}

	return parseType(t, t.Name())
}

func typeFrom(v any) (reflect.Type, bool) {
	if t, ok := v.(reflect.Type); ok {
		if t == nil {
			return nil, false
		}
		return t, true
	}

	return reflect.TypeOf(v), true
}

func parseType(t reflect.Type, name string) (*Field, error) {
	t = dereference(t)

	field := &Field{Name: name}

	switch t.Kind() {
	case reflect.Struct:
		field.Type = FieldTypeStruct
		children, err := parseStructFields(t)
		if err != nil {
			return nil, err
		}
		field.Children = children
	case reflect.Slice:
		field.Type = FieldTypeSlice
		item, err := parseType(t.Elem(), "")
		if err != nil {
			return nil, fmt.Errorf("parse slice %q: %w", name, err)
		}
		field.Item = item
	case reflect.Array:
		field.Type = FieldTypeArray
		item, err := parseType(t.Elem(), "")
		if err != nil {
			return nil, fmt.Errorf("parse array %q: %w", name, err)
		}
		field.Item = item
	case reflect.Map:
		keyType, ok := fieldTypeFromKind(dereference(t.Key()).Kind())
		if !ok || !keyType.IsScalar() {
			return nil, fmt.Errorf("parse map %q: unsupported key type %s", name, t.Key())
		}

		value, err := parseType(t.Elem(), "")
		if err != nil {
			return nil, fmt.Errorf("parse map %q: %w", name, err)
		}

		field.Type = FieldTypeMap
		field.MapKeyType = keyType
		field.MapValue = value
	default:
		fieldType, ok := fieldTypeFromKind(t.Kind())
		if !ok {
			return nil, fmt.Errorf("parse field %q: unsupported type %s", name, t)
		}
		field.Type = fieldType
	}

	return field, nil
}

func parseStructFields(t reflect.Type) ([]*Field, error) {
	fields := make([]*Field, 0, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		structField := t.Field(i)
		if structField.PkgPath != "" {
			continue
		}

		name, skip := yamlName(structField)
		if skip {
			continue
		}

		field, err := parseType(structField.Type, name)
		if err != nil {
			return nil, fmt.Errorf("parse struct %s field %s: %w", t.Name(), structField.Name, err)
		}

		applyTags(field, structField)
		fields = append(fields, field)
	}

	return fields, nil
}

func yamlName(field reflect.StructField) (string, bool) {
	tag, ok := field.Tag.Lookup("yaml")
	if !ok {
		return "", true
	}

	name := strings.TrimSpace(strings.Split(tag, ",")[0])
	if name == "-" {
		return "", true
	}
	if name != "" {
		return name, false
	}

	return "", true
}

func applyTags(field *Field, structField reflect.StructField) {
	field.Required = structField.Tag.Get("required") == "true"
	field.Description = structField.Tag.Get("desc")
	field.Default = structField.Tag.Get("default")
	field.Enum = splitCSVTag(structField.Tag.Get("enum"))
}

func splitCSVTag(value string) []string {
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}

	return values
}

func dereference(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func fieldTypeFromKind(kind reflect.Kind) (FieldType, bool) {
	switch kind {
	case reflect.String:
		return FieldTypeString, true
	case reflect.Bool:
		return FieldTypeBool, true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return FieldTypeInt, true
	case reflect.Float32, reflect.Float64:
		return FieldTypeFloat, true
	default:
		return FieldTypeUnknown, false
	}
}
