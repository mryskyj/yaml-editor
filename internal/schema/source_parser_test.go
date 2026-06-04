package schema

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const sourceParserFixture = `package configs

type Config struct {
	Server Server ` + "`yaml:\"server\" required:\"true\" desc:\"server settings\"`" + `
	App struct {
		Mode string ` + "`yaml:\"mode\" enum:\"dev, stg, prod\" default:\"dev\" desc:\"runtime mode\"`" + `
		Debug bool ` + "`yaml:\"debug\"`" + `
	} ` + "`yaml:\"app\"`" + `
	Ports []int ` + "`yaml:\"ports\"`" + `
	Labels map[string]string ` + "`yaml:\"labels\"`" + `
	Skip string ` + "`yaml:\"-\"`" + `
	hidden string ` + "`yaml:\"hidden\"`" + `
}

type Server struct {
	Host string ` + "`yaml:\"host\" required:\"true\" default:\"localhost\"`" + `
	Port *int ` + "`yaml:\"port\" required:\"true\"`" + `
}
`

func TestParseGoSource(t *testing.T) {
	t.Parallel()

	root, err := ParseGoSource([]byte(sourceParserFixture), "Config")
	if err != nil {
		t.Fatalf("ParseGoSource() returned error: %v", err)
	}

	if root.Name != "Config" {
		t.Fatalf("root.Name = %q, want Config", root.Name)
	}

	server := mustChild(t, root, "server")
	if server.Type != FieldTypeStruct {
		t.Fatalf("server.Type = %q, want %q", server.Type, FieldTypeStruct)
	}
	if !server.Required {
		t.Fatal("server.Required = false, want true")
	}
	if server.Description != "server settings" {
		t.Fatalf("server.Description = %q", server.Description)
	}

	host := mustChild(t, server, "host")
	if host.Type != FieldTypeString {
		t.Fatalf("host.Type = %q, want %q", host.Type, FieldTypeString)
	}
	if host.Default != "localhost" {
		t.Fatalf("host.Default = %q, want localhost", host.Default)
	}

	port := mustChild(t, server, "port")
	if port.Type != FieldTypeInt {
		t.Fatalf("port.Type = %q, want %q", port.Type, FieldTypeInt)
	}

	app := mustChild(t, root, "app")
	mode := mustChild(t, app, "mode")
	wantEnum := []string{"dev", "stg", "prod"}
	if !reflect.DeepEqual(mode.Enum, wantEnum) {
		t.Fatalf("mode.Enum = %#v, want %#v", mode.Enum, wantEnum)
	}
	if mode.Default != "dev" {
		t.Fatalf("mode.Default = %q, want dev", mode.Default)
	}

	ports := mustChild(t, root, "ports")
	if ports.Type != FieldTypeSlice || ports.Item == nil || ports.Item.Type != FieldTypeInt {
		t.Fatalf("ports = %#v, want int slice", ports)
	}

	labels := mustChild(t, root, "labels")
	if labels.Type != FieldTypeMap || labels.MapKeyType != FieldTypeString || labels.MapValue == nil || labels.MapValue.Type != FieldTypeString {
		t.Fatalf("labels = %#v, want string map", labels)
	}

	if _, ok := root.FindChild("Skip"); ok {
		t.Fatal("ParseGoSource() included yaml:\"-\" field")
	}
	if _, ok := root.FindChild("hidden"); ok {
		t.Fatal("ParseGoSource() included unexported field")
	}
}

func TestParseGoSourceFile(t *testing.T) {
	t.Parallel()

	path := writeSchemaSource(t, sourceParserFixture)
	root, err := ParseGoSourceFile(path, "Config")
	if err != nil {
		t.Fatalf("ParseGoSourceFile() returned error: %v", err)
	}
	if _, ok := root.FindChild("server"); !ok {
		t.Fatal("ParseGoSourceFile() missing server field")
	}
}

func TestParseGoSourceRejectsMissingRootType(t *testing.T) {
	t.Parallel()

	if _, err := ParseGoSource([]byte(sourceParserFixture), "Missing"); err == nil {
		t.Fatal("ParseGoSource() error = nil, want error")
	}
}

func TestParseGoSourceRejectsUnsupportedField(t *testing.T) {
	t.Parallel()

	source := []byte(`package configs

type Config struct {
	Callback func() ` + "`yaml:\"callback\"`" + `
}
`)

	if _, err := ParseGoSource(source, "Config"); err == nil {
		t.Fatal("ParseGoSource() error = nil, want error")
	}
}

func TestRegistryRegisterGoSourceFile(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()
	if err := registry.RegisterGoSourceFile(writeSchemaSource(t, sourceParserFixture), "Config"); err != nil {
		t.Fatalf("RegisterGoSourceFile() returned error: %v", err)
	}

	root, err := registry.Root()
	if err != nil {
		t.Fatalf("Root() returned error: %v", err)
	}
	if _, ok := root.FindChild("app"); !ok {
		t.Fatal("Root() missing app field")
	}
}

func writeSchemaSource(t *testing.T, source string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "schema.go")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatalf("os.WriteFile() returned error: %v", err)
	}
	return path
}
