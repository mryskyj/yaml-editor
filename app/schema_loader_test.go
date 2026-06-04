package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mryskyj/yaml-editor/internal/schema"
)

func TestRegisterStartupSchemaFromEnv(t *testing.T) {
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

	t.Setenv(schemaFileEnv, path)
	t.Setenv(schemaTypeEnv, "CustomConfig")

	registry := schema.NewRegistry()
	if err := registerStartupSchema(registry); err != nil {
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
	t.Setenv(schemaTypeEnv, "")

	if got := startupSchemaType(); got != defaultSchemaType {
		t.Fatalf("startupSchemaType() = %q, want %q", got, defaultSchemaType)
	}
}
