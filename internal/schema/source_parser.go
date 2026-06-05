package schema

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// ParseDir converts Go struct definitions in a directory into a YAML editor schema field.
func ParseDir(dir string, rootTypeName string) (*Field, error) {
	return ParseGoSourceDir(dir, rootTypeName)
}

// ParseGoSourceDir parses Go source files in a directory and converts the named root struct into a schema field.
func ParseGoSourceDir(dir string, rootTypeName string) (*Field, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("schema source directory is empty")
	}

	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("read schema source directory %q: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("schema source path %q is not a directory", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read schema source directory %q: %w", dir, err)
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) == ".go" && !strings.HasSuffix(name, "_test.go") {
			paths = append(paths, filepath.Join(dir, name))
		}
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("schema source directory %q has no Go source files", dir)
	}

	return ParseGoSourceFiles(paths, rootTypeName)
}

// ParseGoSourceFiles parses Go source files and converts the named root struct into a schema field.
func ParseGoSourceFiles(paths []string, rootTypeName string) (*Field, error) {
	rootTypeName = strings.TrimSpace(rootTypeName)
	if rootTypeName == "" {
		return nil, fmt.Errorf("schema root type name is empty")
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("schema source files are empty")
	}

	fileSet := token.NewFileSet()
	structs := make(map[string]*ast.StructType)
	for _, path := range paths {
		file, err := parser.ParseFile(fileSet, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("parse schema source %q: %w", path, err)
		}
		if err := collectStructTypes(file, structs); err != nil {
			return nil, err
		}
	}

	return parseCollectedStructs(structs, rootTypeName)
}

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

	structs := make(map[string]*ast.StructType)
	if err := collectStructTypes(file, structs); err != nil {
		return nil, err
	}
	return parseCollectedStructs(structs, rootTypeName)
}

func parseCollectedStructs(structs map[string]*ast.StructType, rootTypeName string) (*Field, error) {
	sourceParser := &sourceParser{
		structs: structs,
		stack:   make(map[string]bool),
	}
	root, ok := sourceParser.structs[rootTypeName]
	if !ok {
		return nil, fmt.Errorf("schema root type %q was not found", rootTypeName)
	}

	return sourceParser.parseNamedStruct(rootTypeName, root, rootTypeName)
}

type sourceParser struct {
	structs map[string]*ast.StructType
	stack   map[string]bool
}

func collectStructTypes(file *ast.File, structs map[string]*ast.StructType) error {
	if file == nil {
		return nil
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
			if typeSpec.TypeParams != nil && typeSpec.TypeParams.NumFields() > 0 {
				return fmt.Errorf("generic type %q is not supported", typeSpec.Name.Name)
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return fmt.Errorf("type alias or non-struct type %q is not supported", typeSpec.Name.Name)
			}
			structs[typeSpec.Name.Name] = structType
		}
	}
	return nil
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
	case *ast.SelectorExpr:
		return nil, fmt.Errorf("parse field %q: external package type references are not supported", name)
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
		return p.parseNamedStruct(ident.Name, structType, name)
	}
	return nil, fmt.Errorf("parse field %q: unsupported type %s", name, ident.Name)
}

func (p *sourceParser) parseNamedStruct(typeName string, structType *ast.StructType, fieldName string) (*Field, error) {
	if p.stack[typeName] {
		return nil, fmt.Errorf("circular struct reference at %q", typeName)
	}
	p.stack[typeName] = true
	defer delete(p.stack, typeName)

	return p.parseStruct(structType, fieldName)
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

	tag := sourceStructTag(structField)
	var fields []*Field
	for _, fieldName := range structField.Names {
		if fieldName == nil || !fieldName.IsExported() {
			continue
		}
		name, skip := yamlNameFromTag(tag)
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

func yamlNameFromTag(tag reflect.StructTag) (string, bool) {
	raw, ok := tag.Lookup("yaml")
	if !ok {
		return "", true
	}
	name := strings.TrimSpace(strings.Split(raw, ",")[0])
	if name == "" || name == "-" {
		return "", true
	}
	return name, false
}

func applySourceTags(field *Field, tag reflect.StructTag) {
	if field == nil {
		return
	}
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
