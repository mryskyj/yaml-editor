package schema

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// ParseGoSourceFile parses a Go source file and converts the named root struct into a schema field.
func ParseGoSourceFile(path string, rootTypeName string) (*Field, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("schema source path is empty")
	}

	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read schema source %q: %w", path, err)
	}
	return ParseGoSource(source, rootTypeName)
}

// ParseGoSource parses Go source and converts the named root struct into a schema field.
func ParseGoSource(source []byte, rootTypeName string) (*Field, error) {
	rootTypeName = strings.TrimSpace(rootTypeName)
	if rootTypeName == "" {
		return nil, fmt.Errorf("schema root type name is empty")
	}

	file, err := parser.ParseFile(token.NewFileSet(), "schema.go", source, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse schema source: %w", err)
	}

	sourceParser := &sourceParser{structs: collectStructTypes(file)}
	root, ok := sourceParser.structs[rootTypeName]
	if !ok {
		return nil, fmt.Errorf("schema root type %q was not found", rootTypeName)
	}

	return sourceParser.parseStruct(root, rootTypeName)
}

type sourceParser struct {
	structs map[string]*ast.StructType
}

func collectStructTypes(file *ast.File) map[string]*ast.StructType {
	structs := make(map[string]*ast.StructType)
	if file == nil {
		return structs
	}

	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.TYPE {
			continue
		}
		for _, spec := range general.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if ok {
				structs[typeSpec.Name.Name] = structType
			}
		}
	}
	return structs
}

func (p *sourceParser) parseType(expr ast.Expr, name string) (*Field, error) {
	switch typed := expr.(type) {
	case *ast.Ident:
		return p.parseIdent(typed, name)
	case *ast.StarExpr:
		return p.parseType(typed.X, name)
	case *ast.StructType:
		return p.parseStruct(typed, name)
	case *ast.ArrayType:
		item, err := p.parseType(typed.Elt, "")
		if err != nil {
			return nil, fmt.Errorf("parse array %q: %w", name, err)
		}
		field := &Field{Name: name, Item: item}
		if typed.Len == nil {
			field.Type = FieldTypeSlice
		} else {
			field.Type = FieldTypeArray
		}
		return field, nil
	case *ast.MapType:
		keyType, err := p.parseMapKey(typed.Key)
		if err != nil {
			return nil, fmt.Errorf("parse map %q: %w", name, err)
		}
		value, err := p.parseType(typed.Value, "")
		if err != nil {
			return nil, fmt.Errorf("parse map %q: %w", name, err)
		}
		return &Field{
			Name:       name,
			Type:       FieldTypeMap,
			MapKeyType: keyType,
			MapValue:   value,
		}, nil
	default:
		return nil, fmt.Errorf("parse field %q: unsupported type %T", name, expr)
	}
}

func (p *sourceParser) parseIdent(ident *ast.Ident, name string) (*Field, error) {
	if ident == nil {
		return nil, fmt.Errorf("parse field %q: empty identifier", name)
	}
	if fieldType, ok := fieldTypeFromGoName(ident.Name); ok {
		return &Field{Name: name, Type: fieldType}, nil
	}
	if structType, ok := p.structs[ident.Name]; ok {
		return p.parseStruct(structType, name)
	}
	return nil, fmt.Errorf("parse field %q: unsupported type %s", name, ident.Name)
}

func (p *sourceParser) parseStruct(structType *ast.StructType, name string) (*Field, error) {
	if structType == nil || structType.Fields == nil {
		return &Field{Name: name, Type: FieldTypeStruct}, nil
	}

	field := &Field{Name: name, Type: FieldTypeStruct}
	for _, structField := range structType.Fields.List {
		children, err := p.parseStructField(structField)
		if err != nil {
			return nil, err
		}
		field.Children = append(field.Children, children...)
	}
	return field, nil
}

func (p *sourceParser) parseStructField(structField *ast.Field) ([]*Field, error) {
	if structField == nil || len(structField.Names) == 0 {
		return nil, nil
	}

	var fields []*Field
	tag := sourceStructTag(structField)
	for _, fieldName := range structField.Names {
		if fieldName == nil || !fieldName.IsExported() {
			continue
		}
		name, skip := yamlNameFromTag(fieldName.Name, tag)
		if skip {
			continue
		}

		field, err := p.parseType(structField.Type, name)
		if err != nil {
			return nil, fmt.Errorf("parse struct field %s: %w", fieldName.Name, err)
		}
		applySourceTags(field, tag)
		fields = append(fields, field)
	}
	return fields, nil
}

func (p *sourceParser) parseMapKey(expr ast.Expr) (FieldType, error) {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return FieldTypeUnknown, fmt.Errorf("unsupported key type %T", expr)
	}

	fieldType, ok := fieldTypeFromGoName(ident.Name)
	if !ok || !fieldType.IsScalar() {
		return FieldTypeUnknown, fmt.Errorf("unsupported key type %s", ident.Name)
	}
	return fieldType, nil
}

func sourceStructTag(field *ast.Field) reflect.StructTag {
	if field == nil || field.Tag == nil {
		return ""
	}

	value, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return ""
	}
	return reflect.StructTag(value)
}

func yamlNameFromTag(fallback string, tag reflect.StructTag) (string, bool) {
	raw := tag.Get("yaml")
	name := strings.TrimSpace(strings.Split(raw, ",")[0])
	if name == "-" {
		return "", true
	}
	if name != "" {
		return name, false
	}
	return fallback, false
}

func applySourceTags(field *Field, tag reflect.StructTag) {
	field.Required = tag.Get("required") == "true"
	field.Description = tag.Get("desc")
	field.Default = tag.Get("default")
	field.Enum = splitCSVTag(tag.Get("enum"))
}

func fieldTypeFromGoName(name string) (FieldType, bool) {
	switch name {
	case "string":
		return FieldTypeString, true
	case "bool":
		return FieldTypeBool, true
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr":
		return FieldTypeInt, true
	case "float32", "float64":
		return FieldTypeFloat, true
	default:
		return FieldTypeUnknown, false
	}
}
