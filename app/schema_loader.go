package app

import (
	"fmt"
	"strings"

	"github.com/mryskyj/yaml-editor/app/sampleschema"
	"github.com/mryskyj/yaml-editor/internal/schema"
)

const (
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

	dir := startupSchemaDir(options)
	if dir == "" {
		return registry.Register(sampleschema.Config{})
	}
	return registry.RegisterGoSourceDir(dir, startupSchemaType(options))
}

func startupSchemaDir(options StartupSchemaOptions) string {
	if dir := strings.TrimSpace(options.Dir); dir != "" {
		return dir
	}
	return ""
}

func startupSchemaType(options StartupSchemaOptions) string {
	if name := strings.TrimSpace(options.Type); name != "" {
		return name
	}
	return defaultSchemaType
}
