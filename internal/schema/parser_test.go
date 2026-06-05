package schema

import (
	"reflect"
	"testing"
)

type parserConfig struct {
	Server parserServer `yaml:"server" required:"true" desc:"server settings"`
	App    parserApp    `yaml:"app"`
	Skip   string       `yaml:"-"`
	JSON   parserJSON   `json:"json"`
	XML    parserXML    `xml:"xml"`
	hidden string
}

type parserServer struct {
	Host string `yaml:"host" default:"127.0.0.1" desc:"listen host"`
	Port int    `yaml:"port" required:"true" default:"8080"`
}

type parserApp struct {
	Mode   string              `yaml:"mode" enum:"dev, stg, prod"`
	Debug  bool                `yaml:"debug"`
	Ratios []float64           `yaml:"ratios"`
	Ports  [2]int              `yaml:"ports"`
	Labels map[string]string   `yaml:"labels"`
	Nested map[string][]string `yaml:"nested"`
}

type parserJSON struct {
	Name string `json:"name"`
}

type parserXML struct {
	Name string `xml:"name"`
}

func TestParseStructTagsAndChildren(t *testing.T) {
	t.Parallel()

	root, err := Parse(parserConfig{})
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if root.Name != "parserConfig" {
		t.Fatalf("root.Name = %q, want %q", root.Name, "parserConfig")
	}
	if root.Type != FieldTypeStruct {
		t.Fatalf("root.Type = %q, want %q", root.Type, FieldTypeStruct)
	}
	if _, ok := root.FindChild("Skip"); ok {
		t.Fatal("Parse() included yaml:\"-\" field")
	}
	if _, ok := root.FindChild("JSON"); ok {
		t.Fatal("Parse() included json-only field")
	}
	if _, ok := root.FindChild("XML"); ok {
		t.Fatal("Parse() included xml-only field")
	}
	if _, ok := root.FindChild("hidden"); ok {
		t.Fatal("Parse() included unexported field")
	}

	server := mustChild(t, root, "server")
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
	if host.Default != "127.0.0.1" {
		t.Fatalf("host.Default = %q", host.Default)
	}

	port := mustChild(t, server, "port")
	if port.Type != FieldTypeInt {
		t.Fatalf("port.Type = %q, want %q", port.Type, FieldTypeInt)
	}
	if !port.Required {
		t.Fatal("port.Required = false, want true")
	}
}

func TestParseEnumSliceArrayAndMap(t *testing.T) {
	t.Parallel()

	root, err := Parse(reflect.TypeOf(parserConfig{}))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	app := mustChild(t, root, "app")

	mode := mustChild(t, app, "mode")
	wantEnum := []string{"dev", "stg", "prod"}
	if !reflect.DeepEqual(mode.Enum, wantEnum) {
		t.Fatalf("mode.Enum = %#v, want %#v", mode.Enum, wantEnum)
	}

	ratios := mustChild(t, app, "ratios")
	if ratios.Type != FieldTypeSlice {
		t.Fatalf("ratios.Type = %q, want %q", ratios.Type, FieldTypeSlice)
	}
	if ratios.Item == nil || ratios.Item.Type != FieldTypeFloat {
		t.Fatalf("ratios.Item = %#v, want float item", ratios.Item)
	}

	ports := mustChild(t, app, "ports")
	if ports.Type != FieldTypeArray {
		t.Fatalf("ports.Type = %q, want %q", ports.Type, FieldTypeArray)
	}
	if ports.Item == nil || ports.Item.Type != FieldTypeInt {
		t.Fatalf("ports.Item = %#v, want int item", ports.Item)
	}

	labels := mustChild(t, app, "labels")
	if labels.Type != FieldTypeMap {
		t.Fatalf("labels.Type = %q, want %q", labels.Type, FieldTypeMap)
	}
	if labels.MapKeyType != FieldTypeString {
		t.Fatalf("labels.MapKeyType = %q, want %q", labels.MapKeyType, FieldTypeString)
	}
	if labels.MapValue == nil || labels.MapValue.Type != FieldTypeString {
		t.Fatalf("labels.MapValue = %#v, want string value", labels.MapValue)
	}

	nested := mustChild(t, app, "nested")
	if nested.MapValue == nil || nested.MapValue.Type != FieldTypeSlice {
		t.Fatalf("nested.MapValue = %#v, want slice value", nested.MapValue)
	}
	if nested.MapValue.Item == nil || nested.MapValue.Item.Type != FieldTypeString {
		t.Fatalf("nested.MapValue.Item = %#v, want string item", nested.MapValue.Item)
	}
}

func TestParsePointerRoot(t *testing.T) {
	t.Parallel()

	root, err := Parse(&parserConfig{})
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}
	if root.Type != FieldTypeStruct {
		t.Fatalf("root.Type = %q, want %q", root.Type, FieldTypeStruct)
	}
}

func TestParseRejectsUnsupportedRoot(t *testing.T) {
	t.Parallel()

	if _, err := Parse("not struct"); err == nil {
		t.Fatal("Parse() error = nil, want error")
	}
}

func TestParseRejectsNilReflectType(t *testing.T) {
	t.Parallel()

	var typ reflect.Type
	if _, err := Parse(typ); err == nil {
		t.Fatal("Parse() error = nil, want error")
	}
}

func TestParseRejectsUnsupportedField(t *testing.T) {
	t.Parallel()

	type invalid struct {
		Callback func() `yaml:"callback"`
	}

	if _, err := Parse(invalid{}); err == nil {
		t.Fatal("Parse() error = nil, want error")
	}
}

func TestParseRejectsUnsupportedMapKey(t *testing.T) {
	t.Parallel()

	type invalid struct {
		Values map[struct{ Name string }]string `yaml:"values"`
	}

	if _, err := Parse(invalid{}); err == nil {
		t.Fatal("Parse() error = nil, want error")
	}
}

func mustChild(t *testing.T, root *Field, name string) *Field {
	t.Helper()

	child, ok := root.FindChild(name)
	if !ok {
		t.Fatalf("FindChild(%q) did not find child", name)
	}
	return child
}
