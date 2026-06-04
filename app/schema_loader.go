package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mryskyj/yaml-editor/internal/schema"
)

const (
	schemaFileEnv     = "YAML_STRUCT_SCHEMA_FILE"
	schemaTypeEnv     = "YAML_STRUCT_SCHEMA_TYPE"
	defaultSchemaFile = "schemas/sample_schema.go"
	defaultSchemaType = "Config"
)

func registerStartupSchema(registry *schema.Registry) error {
	if registry == nil {
		return fmt.Errorf("schema registry is not configured")
	}

	path, err := startupSchemaFile()
	if err != nil {
		return err
	}
	return registry.RegisterGoSourceFile(path, startupSchemaType())
}

func startupSchemaFile() (string, error) {
	if path := os.Getenv(schemaFileEnv); path != "" {
		return path, nil
	}
	return findDefaultSchemaFile()
}

func startupSchemaType() string {
	if name := os.Getenv(schemaTypeEnv); name != "" {
		return name
	}
	return defaultSchemaType
}

func findDefaultSchemaFile() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}

	for {
		path := filepath.Join(dir, defaultSchemaFile)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("default schema file %q was not found", defaultSchemaFile)
}
