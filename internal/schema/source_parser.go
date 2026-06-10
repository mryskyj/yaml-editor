package schema

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

// ParseDir converts Go struct definitions in a directory into a YAML editor schema field.
func ParseDir(dir string, rootType string) (*Field, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("schema source parse failed: schema dir is required")
	}

	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("schema source parse failed: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("schema source parse failed: %s is not a directory", dir)
	}

	structs, err := parseSourceStructs(dir)
	if err != nil {
		return nil, err
	}

	rootType = strings.TrimSpace(rootType)
	if rootType == "" {
		rootType, err = detectSourceRootType(structs)
		if err != nil {
			return nil, err
		}
	}

	root, ok := structs[rootType]
	if !ok {
		return nil, fmt.Errorf("schema source parse failed: root struct %q not found", rootType)
	}

	return parseSourceStruct(rootType, root, structs, map[string]bool{})
}

// ParseFS converts Go struct definitions in a filesystem directory into a YAML editor schema field.
func ParseFS(sourceFS fs.FS, dir string, rootType string) (*Field, error) {
	if sourceFS == nil {
		return nil, fmt.Errorf("schema source parse failed: source filesystem is required")
	}
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("schema source parse failed: schema dir is required")
	}

	structs, err := parseSourceStructsFS(sourceFS, dir)
	if err != nil {
		return nil, err
	}

	rootType = strings.TrimSpace(rootType)
	if rootType == "" {
		rootType, err = detectSourceRootType(structs)
		if err != nil {
			return nil, err
		}
	}

	root, ok := structs[rootType]
	if !ok {
		return nil, fmt.Errorf("schema source parse failed: root struct %q not found", rootType)
	}

	return parseSourceStruct(rootType, root, structs, map[string]bool{})
}

// ParseToolSchemasFS converts all YAML-tagged structs in a filesystem directory into tool schemas.
func ParseToolSchemasFS(sourceFS fs.FS, dir string) (map[string]*Field, error) {
	if sourceFS == nil {
		return nil, fmt.Errorf("schema source parse failed: source filesystem is required")
	}
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("schema source parse failed: schema dir is required")
	}

	packageName, structs, err := parseSourcePackageStructsFS(sourceFS, dir)
	if err != nil {
		return nil, err
	}

	toolSchemas := make(map[string]*Field)
	for typeName, structType := range structs {
		if !sourceStructHasYAMLField(structType) {
			continue
		}
		field, err := parseSourceStruct(typeName, structType, structs, map[string]bool{})
		if err != nil {
			return nil, err
		}
		toolSchemas[packageName+"."+typeName] = field
	}
	return toolSchemas, nil
}

// ParseToolSchemasDir converts all YAML-tagged structs in a directory into tool schemas.
func ParseToolSchemasDir(dir string) (map[string]*Field, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("schema source parse failed: schema dir is required")
	}

	cleanDir := filepath.Clean(dir)
	return ParseToolSchemasFS(os.DirFS(filepath.Dir(cleanDir)), filepath.Base(cleanDir))
}

func parseSourceStructs(dir string) (map[string]*ast.StructType, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("schema source parse failed: %w", err)
	}

	fset := token.NewFileSet()
	structs := make(map[string]*ast.StructType)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		filePath := filepath.Join(dir, name)
		file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("schema source parse failed: parse %s: %w", filePath, err)
		}

		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if typeSpec.TypeParams != nil && typeSpec.TypeParams.NumFields() > 0 {
					return nil, fmt.Errorf("schema source parse failed: generic type %q is not supported", typeSpec.Name.Name)
				}

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					return nil, fmt.Errorf("schema source parse failed: type alias or non-struct type %q is not supported", typeSpec.Name.Name)
				}
				structs[typeSpec.Name.Name] = structType
			}
		}
	}

	return structs, nil
}

func parseSourceStructsFS(sourceFS fs.FS, dir string) (map[string]*ast.StructType, error) {
	_, structs, err := parseSourcePackageStructsFS(sourceFS, dir)
	return structs, err
}

func parseSourcePackageStructsFS(sourceFS fs.FS, dir string) (string, map[string]*ast.StructType, error) {
	entries, err := fs.ReadDir(sourceFS, dir)
	if err != nil {
		return "", nil, fmt.Errorf("schema source parse failed: %w", err)
	}

	fset := token.NewFileSet()
	structs := make(map[string]*ast.StructType)
	packageName := ""

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		filePath := path.Join(dir, name)
		source, err := fs.ReadFile(sourceFS, filePath)
		if err != nil {
			return "", nil, fmt.Errorf("schema source parse failed: read %s: %w", filePath, err)
		}
		filePackage, err := collectSourceStructs(fset, filePath, source, structs)
		if err != nil {
			return "", nil, err
		}
		if packageName == "" {
			packageName = filePackage
		} else if filePackage != packageName {
			return "", nil, fmt.Errorf("schema source parse failed: mixed package names %q and %q", packageName, filePackage)
		}
	}

	if packageName == "" {
		return "", nil, fmt.Errorf("schema source parse failed: package could not be detected")
	}
	return packageName, structs, nil
}

func collectSourceStructs(fset *token.FileSet, filePath string, source []byte, structs map[string]*ast.StructType) (string, error) {
	file, err := parser.ParseFile(fset, filePath, source, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("schema source parse failed: parse %s: %w", filePath, err)
	}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if typeSpec.TypeParams != nil && typeSpec.TypeParams.NumFields() > 0 {
				return "", fmt.Errorf("schema source parse failed: generic type %q is not supported", typeSpec.Name.Name)
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return "", fmt.Errorf("schema source parse failed: type alias or non-struct type %q is not supported", typeSpec.Name.Name)
			}
			structs[typeSpec.Name.Name] = structType
		}
	}

	return file.Name.Name, nil
}

func detectSourceRootType(structs map[string]*ast.StructType) (string, error) {
	referenced := make(map[string]bool)
	for _, structType := range structs {
		for _, sourceField := range structType.Fields.List {
			if _, ok, err := sourceYAMLName(sourceField); err != nil {
				return "", err
			} else if ok {
				collectSourceReferences(sourceField.Type, structs, referenced)
			}
		}
	}

	candidates := make([]string, 0)
	for name, structType := range structs {
		if referenced[name] || !sourceStructHasYAMLField(structType) {
			continue
		}
		candidates = append(candidates, name)
	}

	if len(candidates) == 1 {
		return candidates[0], nil
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("schema source parse failed: root struct could not be detected")
	}
	return "", fmt.Errorf("schema source parse failed: root struct is ambiguous: %s", strings.Join(candidates, ", "))
}

func sourceStructHasYAMLField(structType *ast.StructType) bool {
	if structType == nil || structType.Fields == nil {
		return false
	}
	for _, sourceField := range structType.Fields.List {
		if _, ok, err := sourceYAMLName(sourceField); err == nil && ok {
			return true
		}
	}
	return false
}

func collectSourceReferences(expr ast.Expr, structs map[string]*ast.StructType, referenced map[string]bool) {
	switch t := expr.(type) {
	case *ast.Ident:
		if _, ok := structs[t.Name]; ok {
			referenced[t.Name] = true
		}
	case *ast.StarExpr:
		collectSourceReferences(t.X, structs, referenced)
	case *ast.ArrayType:
		collectSourceReferences(t.Elt, structs, referenced)
	case *ast.MapType:
		collectSourceReferences(t.Value, structs, referenced)
	}
}

func parseSourceStruct(name string, structType *ast.StructType, structs map[string]*ast.StructType, stack map[string]bool) (*Field, error) {
	if stack[name] {
		return nil, fmt.Errorf("schema source parse failed: circular struct reference at %q", name)
	}

	stack[name] = true
	defer delete(stack, name)

	field := &Field{Name: name, Type: FieldTypeStruct}
	children := make([]*Field, 0, len(structType.Fields.List))

	for _, sourceField := range structType.Fields.List {
		fieldName, ok, err := sourceYAMLName(sourceField)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		child, err := parseSourceExpr(fieldName, sourceField.Type, structs, stack)
		if err != nil {
			return nil, fmt.Errorf("parse struct %s field %s: %w", name, fieldName, err)
		}
		applySourceTags(child, sourceField)
		children = append(children, child)
	}

	field.Children = children
	return field, nil
}

func parseSourceExpr(name string, expr ast.Expr, structs map[string]*ast.StructType, stack map[string]bool) (*Field, error) {
	switch t := expr.(type) {
	case *ast.Ident:
		if fieldType, ok := sourceScalarType(t.Name); ok {
			return &Field{Name: name, Type: fieldType}, nil
		}
		structType, ok := structs[t.Name]
		if !ok {
			return nil, fmt.Errorf("unsupported type %q", t.Name)
		}
		field, err := parseSourceStruct(t.Name, structType, structs, stack)
		if err != nil {
			return nil, err
		}
		field.Name = name
		return field, nil
	case *ast.StarExpr:
		return parseSourceExpr(name, t.X, structs, stack)
	case *ast.StructType:
		return parseSourceStruct(name, t, structs, stack)
	case *ast.ArrayType:
		item, err := parseSourceExpr("", t.Elt, structs, stack)
		if err != nil {
			return nil, err
		}
		fieldType := FieldTypeArray
		if t.Len == nil {
			fieldType = FieldTypeSlice
		}
		return &Field{Name: name, Type: fieldType, Item: item}, nil
	case *ast.MapType:
		key, ok := sourceMapKeyType(t.Key)
		if !ok {
			return nil, fmt.Errorf("unsupported map key type")
		}
		value, err := parseSourceExpr("", t.Value, structs, stack)
		if err != nil {
			return nil, err
		}
		return &Field{Name: name, Type: FieldTypeMap, MapKeyType: key, MapValue: value}, nil
	case *ast.SelectorExpr:
		return nil, fmt.Errorf("external package type references are not supported")
	default:
		return nil, fmt.Errorf("unsupported type expression %T", expr)
	}
}

func sourceYAMLName(field *ast.Field) (string, bool, error) {
	if field.Tag == nil {
		return "", false, nil
	}

	tagValue, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return "", false, fmt.Errorf("schema source parse failed: invalid struct tag %q: %w", field.Tag.Value, err)
	}

	tag, ok := reflect.StructTag(tagValue).Lookup("yaml")
	if !ok {
		return "", false, nil
	}

	name := strings.TrimSpace(strings.Split(tag, ",")[0])
	if name == "" || name == "-" {
		return "", false, nil
	}
	return name, true, nil
}

func applySourceTags(field *Field, sourceField *ast.Field) {
	if field == nil || sourceField.Tag == nil {
		return
	}

	tagValue, err := strconv.Unquote(sourceField.Tag.Value)
	if err != nil {
		return
	}

	tag := reflect.StructTag(tagValue)
	field.Required = tag.Get("required") == "true"
	field.Description = tag.Get("desc")
	field.Default = tag.Get("default")
	field.Enum = splitCSVTag(tag.Get("enum"))
}

func sourceScalarType(name string) (FieldType, bool) {
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

func sourceMapKeyType(expr ast.Expr) (FieldType, bool) {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return FieldTypeUnknown, false
	}
	fieldType, ok := sourceScalarType(ident.Name)
	if !ok || !fieldType.IsScalar() {
		return FieldTypeUnknown, false
	}
	return fieldType, true
}
