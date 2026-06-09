package app

import (
	"embed"

	"github.com/mryskyj/yaml-editor/internal/schema"
)

//go:embed sampleschema/*.go
var sampleSchemaSource embed.FS

func registerSampleSchema(registry *schema.Registry) error {
	if registry == nil {
		return nil
	}
	root, err := schema.ParseFS(sampleSchemaSource, "sampleschema", "")
	if err != nil {
		return err
	}
	registry.SetRoot(root)
	return nil
}
