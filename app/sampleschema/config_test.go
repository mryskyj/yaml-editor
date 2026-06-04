package sampleschema

import (
	"testing"

	"github.com/mryskyj/yaml-editor/internal/schema"
)

func TestConfigSchemaIncludesTopLevelSections(t *testing.T) {
	t.Parallel()

	root, err := schema.Parse(Config{})
	if err != nil {
		t.Fatalf("schema.Parse() returned error: %v", err)
	}

	for _, name := range []string{
		"server",
		"app",
		"aws",
		"cloudformation",
		"ecs",
		"ssm",
		"observability",
		"deployment",
		"security",
	} {
		if _, ok := root.FindChild(name); !ok {
			t.Fatalf("Config schema missing %q field", name)
		}
	}
}
