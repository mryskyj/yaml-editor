package app

import (
	"embed"

	"github.com/mryskyj/yaml-editor/internal/schema"
)

//go:embed rootschema/*.go sampleschema
var sampleSchemaSource embed.FS

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
