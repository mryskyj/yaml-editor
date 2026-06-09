package sampleschema

import (
	"testing"

	"github.com/mryskyj/yaml-editor/internal/schema"
)

func TestConfigSchemaIncludesTopLevelSections(t *testing.T) {
	t.Parallel()

	root, err := schema.ParseDir(".", "")
	if err != nil {
		t.Fatalf("schema.ParseDir() returned error: %v", err)
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

func TestConfigSchemaUsesOnlyYAMLTaggedFields(t *testing.T) {
	t.Parallel()

	root, err := schema.ParseDir(".", "")
	if err != nil {
		t.Fatalf("schema.ParseDir() returned error: %v", err)
	}

	if _, ok := root.FindChild("JSONImport"); ok {
		t.Fatal("Config schema included json-only field name")
	}
	if _, ok := root.FindChild("json_import"); ok {
		t.Fatal("Config schema included json-only tag name")
	}
	if _, ok := root.FindChild("XMLImport"); ok {
		t.Fatal("Config schema included xml-only field name")
	}
	if _, ok := root.FindChild("xml_import"); ok {
		t.Fatal("Config schema included xml-only tag name")
	}

	server, ok := root.FindChild("server")
	if !ok {
		t.Fatal("Config schema missing yaml-tagged server field")
	}
	if _, ok := root.FindChild("Server"); ok {
		t.Fatal("Config schema used Go field name instead of yaml tag")
	}
	if _, ok := server.FindChild("host"); !ok {
		t.Fatal("server schema missing yaml-tagged host field")
	}
	if _, ok := server.FindChild("Host"); ok {
		t.Fatal("server schema used Go field name instead of yaml tag")
	}
}
