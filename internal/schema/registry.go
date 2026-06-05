package schema

import "fmt"

// Registry holds the parsed root schema used by the application.
type Registry struct {
	root *Field
}

// NewRegistry creates an empty schema registry.
func NewRegistry() *Registry {
	return &Registry{}
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

// RegisterGoSourceFile parses and stores the root schema from a Go source file.
func (r *Registry) RegisterGoSourceFile(path string, rootTypeName string) error {
	if r == nil {
		return fmt.Errorf("schema registry is nil")
	}

	root, err := ParseGoSourceFile(path, rootTypeName)
	if err != nil {
		return err
	}

	r.root = root
	return nil
}

// RegisterGoSourceDir parses and stores the root schema from Go source files in a directory.
func (r *Registry) RegisterGoSourceDir(dir string, rootTypeName string) error {
	if r == nil {
		return fmt.Errorf("schema registry is nil")
	}

	root, err := ParseGoSourceDir(dir, rootTypeName)
	if err != nil {
		return err
	}

	r.root = root
	return nil
}

// RegisterFromDir parses and stores the root schema from Go source files in a directory.
func (r *Registry) RegisterFromDir(dir string, rootTypeName string) error {
	return r.RegisterGoSourceDir(dir, rootTypeName)
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
