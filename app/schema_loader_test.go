package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mryskyj/yaml-editor/internal/schema"
)

func TestRegisterStartupSchemaFromOptions(t *testing.T) {
	source := `package configs

type CustomConfig struct {
	Database Database ` + "`yaml:\"database\" required:\"true\"`" + `
}

type Database struct {
	Host string ` + "`yaml:\"host\" required:\"true\"`" + `
}
`
	path := filepath.Join(t.TempDir(), "schema.go")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatalf("os.WriteFile() returned error: %v", err)
	}

	registry := schema.NewRegistry()
	if err := registerStartupSchema(registry, StartupSchemaOptions{
		Dir:  filepath.Dir(path),
		Type: "CustomConfig",
	}); err != nil {
		t.Fatalf("registerStartupSchema() returned error: %v", err)
	}

	root, err := registry.Root()
	if err != nil {
		t.Fatalf("Root() returned error: %v", err)
	}
	if _, ok := root.FindChild("database"); !ok {
		t.Fatal("Root() missing database field")
	}
}

func TestStartupSchemaTypeDefault(t *testing.T) {
	if got := startupSchemaType(StartupSchemaOptions{}); got != defaultSchemaType {
		t.Fatalf("startupSchemaType() = %q, want %q", got, defaultSchemaType)
	}
}

func TestStartupSchemaDirUsesOption(t *testing.T) {
	dir := t.TempDir()

	got := startupSchemaDir(StartupSchemaOptions{Dir: dir})
	if got != dir {
		t.Fatalf("startupSchemaDir() = %q, want %q", got, dir)
	}
}

func TestRegisterStartupSchemaUsesSampleSchemaByDefault(t *testing.T) {
	registry := schema.NewRegistry()
	if err := registerStartupSchema(registry, StartupSchemaOptions{}); err != nil {
		t.Fatalf("registerStartupSchema() returned error: %v", err)
	}

	root, err := registry.Root()
	if err != nil {
		t.Fatalf("Root() returned error: %v", err)
	}
	if _, ok := root.FindChild("aws"); !ok {
		t.Fatal("Root() missing aws field from sample schema")
	}
}
