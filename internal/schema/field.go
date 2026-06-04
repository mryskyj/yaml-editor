package schema

// FieldType represents a Go type category used by the YAML editor schema.
type FieldType string

const (
	FieldTypeUnknown FieldType = "unknown"
	FieldTypeStruct  FieldType = "struct"
	FieldTypeSlice   FieldType = "slice"
	FieldTypeArray   FieldType = "array"
	FieldTypeMap     FieldType = "map"
	FieldTypeString  FieldType = "string"
	FieldTypeBool    FieldType = "bool"
	FieldTypeInt     FieldType = "int"
	FieldTypeFloat   FieldType = "float"
)

// Field describes a YAML field derived from a Go struct field.
type Field struct {
	Name        string
	Type        FieldType
	Required    bool
	Description string
	Default     string
	Enum        []string
	Children    []*Field
	Item        *Field
	MapKeyType  FieldType
	MapValue    *Field
}

// FindChild returns the direct child field with the given YAML key name.
func (f *Field) FindChild(name string) (*Field, bool) {
	if f == nil {
		return nil, false
	}

	for _, child := range f.Children {
		if child != nil && child.Name == name {
			return child, true
		}
	}

	return nil, false
}

// IsScalar reports whether the field stores a scalar YAML value.
func (t FieldType) IsScalar() bool {
	switch t {
	case FieldTypeString, FieldTypeBool, FieldTypeInt, FieldTypeFloat:
		return true
	default:
		return false
	}
}
