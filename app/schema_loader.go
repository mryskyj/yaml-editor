package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mryskyj/yaml-editor/internal/schema"
)

const (
	defaultSchemaDir  = "schemas"
	defaultSchemaType = "Config"
)

// StartupSchemaOptions configures the schema loaded once during application startup.
type StartupSchemaOptions struct {
	Dir  string
	Type string
}

func registerStartupSchema(registry *schema.Registry, options StartupSchemaOptions) error {
	if registry == nil {
		return fmt.Errorf("schema registry is not configured")
	}

	dir, err := startupSchemaDir(options)
	if err != nil {
		return err
	}
	return registry.RegisterGoSourceDir(dir, startupSchemaType(options))
}

func startupSchemaDir(options StartupSchemaOptions) (string, error) {
	if dir := strings.TrimSpace(options.Dir); dir != "" {
		return dir, nil
	}
	return findDefaultSchemaDir()
}

func startupSchemaType(options StartupSchemaOptions) string {
	if name := strings.TrimSpace(options.Type); name != "" {
		return name
	}
	return defaultSchemaType
}

func findDefaultSchemaDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}

	for {
		path := filepath.Join(dir, defaultSchemaDir)
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("default schema directory %q was not found", defaultSchemaDir)
}
