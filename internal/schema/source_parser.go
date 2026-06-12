package schema

import (
	"fmt"
	"go/ast"
	"go/build"
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

type sourceTypes struct {
	structs     map[string]*ast.StructType
	scalars     map[string]FieldType
	collections map[string]ast.Expr
	imports     map[string]string
	external    map[string]*sourceTypes
}

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

	sourceTypes, err := parseSourceStructs(dir)
	if err != nil {
		return nil, err
	}

	rootType = strings.TrimSpace(rootType)
	if rootType == "" {
		rootType, err = detectSourceRootType(sourceTypes)
		if err != nil {
			return nil, err
		}
	}

	root, ok := sourceTypes.structs[rootType]
	if !ok {
		return nil, fmt.Errorf("schema source parse failed: root struct %q not found", rootType)
	}

	return parseSourceStruct(rootType, root, sourceTypes, map[string]bool{})
}

// ParseFS converts Go struct definitions in a filesystem directory into a YAML editor schema field.
func ParseFS(sourceFS fs.FS, dir string, rootType string) (*Field, error) {
	if sourceFS == nil {
		return nil, fmt.Errorf("schema source parse failed: source filesystem is required")
	}
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("schema source parse failed: schema dir is required")
	}

	sourceTypes, err := parseSourceStructsFS(sourceFS, dir)
	if err != nil {
		return nil, err
	}

	rootType = strings.TrimSpace(rootType)
	if rootType == "" {
		rootType, err = detectSourceRootType(sourceTypes)
		if err != nil {
			return nil, err
		}
	}

	root, ok := sourceTypes.structs[rootType]
	if !ok {
		return nil, fmt.Errorf("schema source parse failed: root struct %q not found", rootType)
	}

	return parseSourceStruct(rootType, root, sourceTypes, map[string]bool{})
}

// ParseToolSchemasFS converts all YAML-tagged structs in a filesystem directory into tool schemas.
func ParseToolSchemasFS(sourceFS fs.FS, dir string) (map[string]*Field, error) {
	if sourceFS == nil {
		return nil, fmt.Errorf("schema source parse failed: source filesystem is required")
	}
	if strings.TrimSpace(dir) == "" {
		return nil, fmt.Errorf("schema source parse failed: schema dir is required")
	}

	toolSchemas := make(map[string]*Field)
	if err := fs.WalkDir(sourceFS, dir, func(currentDir string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("schema source parse failed: %w", walkErr)
		}
		if !entry.IsDir() {
			return nil
		}
		hasGoSource, err := fsDirHasGoSource(sourceFS, currentDir)
		if err != nil {
			return err
		}
		if !hasGoSource {
			return nil
		}

		packageName, sourceTypes, err := parseSourcePackageStructsFS(sourceFS, currentDir)
		if err != nil {
			return err
		}
		namespace := toolSchemaNamespace(dir, currentDir, packageName)

		for typeName, structType := range sourceTypes.structs {
			if !sourceStructHasYAMLField(structType) {
				continue
			}
			field, err := parseSourceStruct(typeName, structType, sourceTypes, map[string]bool{})
			if err != nil {
				return err
			}
			toolSchemas[namespace+"."+typeName] = field
		}
		for typeName, expr := range sourceTypes.collections {
			field, err := parseSourceExpr(typeName, expr, sourceTypes, map[string]bool{})
			if err != nil {
				return err
			}
			if !fieldHasYAMLContent(field) {
				continue
			}
			toolSchemas[namespace+"."+typeName] = field
		}
		return nil
	}); err != nil {
		return nil, err
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

func fsDirHasGoSource(sourceFS fs.FS, dir string) (bool, error) {
	entries, err := fs.ReadDir(sourceFS, dir)
	if err != nil {
		return false, fmt.Errorf("schema source parse failed: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			return true, nil
		}
	}
	return false, nil
}

func toolSchemaNamespace(rootDir string, currentDir string, packageName string) string {
	relative := strings.Trim(strings.TrimPrefix(currentDir, rootDir), "/")
	if relative == "" {
		return packageName
	}
	return strings.ReplaceAll(relative, "/", ".")
}

func parseSourceStructs(dir string) (*sourceTypes, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("schema source parse failed: %w", err)
	}

	fset := token.NewFileSet()
	sourceTypes := newSourceTypes()

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

		if err := collectSourceTypeDecls(file, filePath, sourceTypes); err != nil {
			return nil, err
		}
	}

	return sourceTypes, nil
}

func parseSourceStructsFS(sourceFS fs.FS, dir string) (*sourceTypes, error) {
	_, sourceTypes, err := parseSourcePackageStructsFS(sourceFS, dir)
	return sourceTypes, err
}

func parseSourcePackageStructsFS(sourceFS fs.FS, dir string) (string, *sourceTypes, error) {
	entries, err := fs.ReadDir(sourceFS, dir)
	if err != nil {
		return "", nil, fmt.Errorf("schema source parse failed: %w", err)
	}

	fset := token.NewFileSet()
	sourceTypes := newSourceTypes()
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
		filePackage, err := collectSourceTypes(fset, filePath, source, sourceTypes)
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
	return packageName, sourceTypes, nil
}

func newSourceTypes() *sourceTypes {
	return &sourceTypes{
		structs:     make(map[string]*ast.StructType),
		scalars:     make(map[string]FieldType),
		collections: make(map[string]ast.Expr),
		imports:     make(map[string]string),
		external:    make(map[string]*sourceTypes),
	}
}

func collectSourceTypes(fset *token.FileSet, filePath string, source []byte, sourceTypes *sourceTypes) (string, error) {
	file, err := parser.ParseFile(fset, filePath, source, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("schema source parse failed: parse %s: %w", filePath, err)
	}

	if err := collectSourceTypeDecls(file, filePath, sourceTypes); err != nil {
		return "", err
	}

	return file.Name.Name, nil
}

func collectSourceTypeDecls(file *ast.File, filePath string, sourceTypes *sourceTypes) error {
	_ = filePath
	collectSourceImports(file, sourceTypes)
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
				return fmt.Errorf("schema source parse failed: generic type %q is not supported", typeSpec.Name.Name)
			}
			if typeSpec.Assign.IsValid() {
				return fmt.Errorf("schema source parse failed: type alias %q is not supported", typeSpec.Name.Name)
			}

			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				sourceTypes.structs[typeSpec.Name.Name] = structType
				continue
			}
			if scalarType, ok := sourceScalarExprType(typeSpec.Type); ok {
				sourceTypes.scalars[typeSpec.Name.Name] = scalarType
				continue
			}
			if sourceCollectionExpr(typeSpec.Type) {
				sourceTypes.collections[typeSpec.Name.Name] = typeSpec.Type
			}
		}
	}
	return nil
}

func collectSourceImports(file *ast.File, sourceTypes *sourceTypes) {
	if file == nil || sourceTypes == nil {
		return
	}
	for _, sourceImport := range file.Imports {
		if sourceImport.Path == nil {
			continue
		}
		importPath, err := strconv.Unquote(sourceImport.Path.Value)
		if err != nil || importPath == "" {
			continue
		}
		alias := path.Base(importPath)
		if sourceImport.Name != nil {
			if sourceImport.Name.Name == "." || sourceImport.Name.Name == "_" {
				continue
			}
			alias = sourceImport.Name.Name
		}
		sourceTypes.imports[alias] = importPath
	}
}

func detectSourceRootType(sourceTypes *sourceTypes) (string, error) {
	referenced := make(map[string]bool)
	for _, structType := range sourceTypes.structs {
		for _, sourceField := range structType.Fields.List {
			if _, ok, err := sourceYAMLName(sourceField); err != nil {
				return "", err
			} else if ok {
				collectSourceReferences(sourceField.Type, sourceTypes, referenced)
			}
		}
	}

	candidates := make([]string, 0)
	for name, structType := range sourceTypes.structs {
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

func fieldHasYAMLContent(field *Field) bool {
	if field == nil {
		return false
	}
	if len(field.Children) > 0 {
		return true
	}
	if field.Item != nil {
		return fieldHasYAMLContent(field.Item)
	}
	if field.MapValue != nil {
		return fieldHasYAMLContent(field.MapValue)
	}
	return field.Type.IsScalar()
}

func sourceCollectionExpr(expr ast.Expr) bool {
	switch expr.(type) {
	case *ast.ArrayType, *ast.MapType:
		return true
	default:
		return false
	}
}

func collectSourceReferences(expr ast.Expr, sourceTypes *sourceTypes, referenced map[string]bool) {
	switch t := expr.(type) {
	case *ast.Ident:
		if _, ok := sourceTypes.structs[t.Name]; ok {
			referenced[t.Name] = true
		}
	case *ast.StarExpr:
		collectSourceReferences(t.X, sourceTypes, referenced)
	case *ast.ArrayType:
		collectSourceReferences(t.Elt, sourceTypes, referenced)
	case *ast.MapType:
		collectSourceReferences(t.Value, sourceTypes, referenced)
	}
}

func parseSourceStruct(name string, structType *ast.StructType, sourceTypes *sourceTypes, stack map[string]bool) (*Field, error) {
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

		child, err := parseSourceExpr(fieldName, sourceField.Type, sourceTypes, stack)
		if err != nil {
			return nil, fmt.Errorf("parse struct %s field %s: %w", name, fieldName, err)
		}
		applySourceTags(child, sourceField)
		children = append(children, child)
	}

	field.Children = children
	return field, nil
}

func parseSourceExpr(name string, expr ast.Expr, sourceTypes *sourceTypes, stack map[string]bool) (*Field, error) {
	switch t := expr.(type) {
	case *ast.Ident:
		if fieldType, ok := sourceScalarType(t.Name); ok {
			return &Field{Name: name, Type: fieldType}, nil
		}
		if fieldType, ok := sourceTypes.scalars[t.Name]; ok {
			return &Field{Name: name, Type: fieldType}, nil
		}
		structType, ok := sourceTypes.structs[t.Name]
		if ok {
			field, err := parseSourceStruct(t.Name, structType, sourceTypes, stack)
			if err != nil {
				return nil, err
			}
			field.Name = name
			return field, nil
		}

		collectionExpr, ok := sourceTypes.collections[t.Name]
		if !ok {
			return nil, fmt.Errorf("unsupported type %q", t.Name)
		}
		if stack[t.Name] {
			return nil, fmt.Errorf("schema source parse failed: circular type reference at %q", t.Name)
		}
		stack[t.Name] = true
		defer delete(stack, t.Name)
		field, err := parseSourceExpr(t.Name, collectionExpr, sourceTypes, stack)
		if err != nil {
			return nil, err
		}
		field.Name = name
		return field, nil
	case *ast.StarExpr:
		return parseSourceExpr(name, t.X, sourceTypes, stack)
	case *ast.StructType:
		return parseSourceStruct(name, t, sourceTypes, stack)
	case *ast.ArrayType:
		item, err := parseSourceExpr("", t.Elt, sourceTypes, stack)
		if err != nil {
			return nil, err
		}
		fieldType := FieldTypeArray
		if t.Len == nil {
			fieldType = FieldTypeSlice
		}
		return &Field{Name: name, Type: fieldType, Item: item}, nil
	case *ast.MapType:
		key, ok := sourceMapKeyType(t.Key, sourceTypes)
		if !ok {
			return nil, fmt.Errorf("unsupported map key type")
		}
		value, err := parseSourceExpr("", t.Value, sourceTypes, stack)
		if err != nil {
			return nil, err
		}
		return &Field{Name: name, Type: FieldTypeMap, MapKeyType: key, MapValue: value}, nil
	case *ast.SelectorExpr:
		packageName, ok := t.X.(*ast.Ident)
		if !ok {
			return nil, fmt.Errorf("unsupported external package selector")
		}
		externalTypes, err := resolveExternalSourceTypes(packageName.Name, sourceTypes)
		if err != nil {
			return nil, err
		}
		field, err := parseSourceExpr(name, ast.NewIdent(t.Sel.Name), externalTypes, map[string]bool{})
		if err != nil {
			return nil, err
		}
		field.Name = name
		return field, nil
	default:
		return nil, fmt.Errorf("unsupported type expression %T", expr)
	}
}

func resolveExternalSourceTypes(alias string, types *sourceTypes) (*sourceTypes, error) {
	if types == nil {
		return nil, fmt.Errorf("schema source parse failed: external package %q cannot be resolved", alias)
	}

	importPath, ok := types.imports[alias]
	if !ok {
		return nil, fmt.Errorf("schema source parse failed: external package alias %q is not imported", alias)
	}
	if types.external == nil {
		types.external = make(map[string]*sourceTypes)
	}
	if cached, ok := types.external[importPath]; ok {
		return cached, nil
	}

	dir, err := resolveImportDir(importPath)
	if err != nil {
		return nil, err
	}
	externalTypes, err := parseSourceStructs(dir)
	if err != nil {
		return nil, fmt.Errorf("schema source parse failed: parse imported package %q: %w", importPath, err)
	}
	types.external[importPath] = externalTypes
	return externalTypes, nil
}

func resolveImportDir(importPath string) (string, error) {
	for _, gopath := range gopathList() {
		dir := filepath.Join(gopath, "src", filepath.FromSlash(importPath))
		info, err := os.Stat(dir)
		if err == nil && info.IsDir() {
			return dir, nil
		}
	}

	pkg, err := build.Default.Import(importPath, "", build.FindOnly)
	if err == nil && pkg.Dir != "" {
		if pkg.Goroot {
			return "", fmt.Errorf("schema source parse failed: standard library package type %q is not supported", importPath)
		}
		return pkg.Dir, nil
	}

	return "", fmt.Errorf("schema source parse failed: import package %q could not be resolved from GOPATH", importPath)
}

func gopathList() []string {
	gopath := os.Getenv("GOPATH")
	if strings.TrimSpace(gopath) == "" {
		gopath = build.Default.GOPATH
	}

	paths := make([]string, 0)
	for _, entry := range filepath.SplitList(gopath) {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		paths = append(paths, entry)
	}
	return paths
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

	name := yamlTagName(tag)
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
	field.Required = !yamlTagHasOption(tag.Get("yaml"), "omitempty")
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

func sourceScalarExprType(expr ast.Expr) (FieldType, bool) {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return FieldTypeUnknown, false
	}
	return sourceScalarType(ident.Name)
}

func sourceMapKeyType(expr ast.Expr, sourceTypes *sourceTypes) (FieldType, bool) {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return FieldTypeUnknown, false
	}
	fieldType, ok := sourceScalarType(ident.Name)
	if !ok {
		fieldType, ok = sourceTypes.scalars[ident.Name]
	}
	if !ok || !fieldType.IsScalar() {
		return FieldTypeUnknown, false
	}
	return fieldType, true
}
