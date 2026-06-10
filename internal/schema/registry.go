package schema

import (
	"fmt"
	"io/fs"
)

// Registry holds the parsed root schema used by the application.
type Registry struct {
	root        *Field
	toolSchemas map[string]*Field
}

// NewRegistry creates an empty schema registry.
func NewRegistry() *Registry {
	return &Registry{
		toolSchemas: make(map[string]*Field),
	}
}

// Register parses and stores the root schema from a Go struct value or type.
func (r *Registry) Register(v any) error {
	if r == nil {
		return fmt.Errorf("schema registry is nil")
	}

	root, err := Parse(v)
	if err != nil {
		return err
	}

	r.root = root
	return nil
}

// RegisterFromDir parses and stores the root schema from Go source files in a directory.
func (r *Registry) RegisterFromDir(dir string, rootType string) error {
	if r == nil {
		return fmt.Errorf("schema registry is nil")
	}

	root, err := ParseDir(dir, rootType)
	if err != nil {
		return err
	}

	r.root = root
	return nil
}

// SetRoot stores an already parsed root schema.
func (r *Registry) SetRoot(root *Field) error {
	if r == nil {
		return fmt.Errorf("schema registry is nil")
	}
	if root == nil {
		return fmt.Errorf("root schema is nil")
	}

	r.root = root
	return nil
}

// RegisterToolSchemasFS parses tool schemas from Go source files in a filesystem directory.
func (r *Registry) RegisterToolSchemasFS(sourceFS fs.FS, dir string) error {
	if r == nil {
		return fmt.Errorf("schema registry is nil")
	}

	toolSchemas, err := ParseToolSchemasFS(sourceFS, dir)
	if err != nil {
		return err
	}
	if r.toolSchemas == nil {
		r.toolSchemas = make(map[string]*Field, len(toolSchemas))
	}
	for name, field := range toolSchemas {
		r.toolSchemas[name] = field
	}
	return nil
}

// RegisterToolSchemasFromDir parses tool schemas from Go source files in a directory.
func (r *Registry) RegisterToolSchemasFromDir(dir string) error {
	if r == nil {
		return fmt.Errorf("schema registry is nil")
	}

	toolSchemas, err := ParseToolSchemasDir(dir)
	if err != nil {
		return err
	}
	if r.toolSchemas == nil {
		r.toolSchemas = make(map[string]*Field, len(toolSchemas))
	}
	for name, field := range toolSchemas {
		r.toolSchemas[name] = field
	}
	return nil
}

// Root returns the registered root schema.
func (r *Registry) Root() (*Field, error) {
	if r == nil {
		return nil, fmt.Errorf("schema registry is nil")
	}
	if r.root == nil {
		return nil, fmt.Errorf("root schema is not registered")
	}

	return r.root, nil
}

// ToolSchemas returns registered tool argument schemas keyed by <namespace>.<struct>.
func (r *Registry) ToolSchemas() map[string]*Field {
	if r == nil || len(r.toolSchemas) == 0 {
		return nil
	}

	toolSchemas := make(map[string]*Field, len(r.toolSchemas))
	for name, field := range r.toolSchemas {
		toolSchemas[name] = field
	}
	return toolSchemas
}
