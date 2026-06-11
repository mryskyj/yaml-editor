package app

import (
	"embed"

	"github.com/mryskyj/yaml-editor/internal/schema"
)

//go:embed rootschema/*.go rootschema/defaults.yaml sampleschema
var sampleSchemaSource embed.FS

//go:embed rootschema/defaults.yaml
var rootDefaultsSource []byte

func registerSampleSchema(registry *schema.Registry) error {
	if registry == nil {
		return nil
	}
	root, err := schema.ParseFS(sampleSchemaSource, "rootschema", "File")
	if err != nil {
		return err
	}
	if err := registry.SetRoot(root); err != nil {
		return err
	}
	return registry.RegisterToolSchemasFS(sampleSchemaSource, "sampleschema")
}
