package schema

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/fstest"
)

func TestParseDirResolvesStructsAcrossFiles(t *testing.T) {
	t.Parallel()

	root, err := ParseDir(filepath.Join("..", "..", "schemas", "external-sample"), "Config")
	if err != nil {
		t.Fatalf("ParseDir() returned error: %v", err)
	}

	if root.Name != "Config" {
		t.Fatalf("root.Name = %q, want Config", root.Name)
	}

	server := mustChild(t, root, "server")
	host := mustChild(t, server, "host")
	if host.Type != FieldTypeString {
		t.Fatalf("host.Type = %q, want %q", host.Type, FieldTypeString)
	}
	if !host.Required {
		t.Fatal("host.Required = false, want true")
	}
	if host.Default != "127.0.0.1" {
		t.Fatalf("host.Default = %q, want 127.0.0.1", host.Default)
	}

	app := mustChild(t, root, "app")
	mode := mustChild(t, app, "mode")
	wantEnum := []string{"dev", "stg", "prod"}
	if !reflect.DeepEqual(mode.Enum, wantEnum) {
		t.Fatalf("mode.Enum = %#v, want %#v", mode.Enum, wantEnum)
	}

	weights := mustChild(t, app, "weights")
	if weights.Type != FieldTypeSlice || weights.Item == nil || weights.Item.Type != FieldTypeFloat {
		t.Fatalf("weights = %#v, want float slice", weights)
	}

	labels := mustChild(t, app, "labels")
	if labels.Type != FieldTypeMap || labels.MapKeyType != FieldTypeString || labels.MapValue == nil || labels.MapValue.Type != FieldTypeString {
		t.Fatalf("labels = %#v, want string map", labels)
	}
}

func TestParseDirRejectsMissingRoot(t *testing.T) {
	t.Parallel()

	if _, err := ParseDir(filepath.Join("..", "..", "schemas", "external-sample"), "Missing"); err == nil {
		t.Fatal("ParseDir() error = nil, want error")
	}
}

func TestParseDirDetectsRootWhenTypeIsEmpty(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeSourceFile(t, dir, "schema.go", `package sample

type CustomRoot struct {
	Server Server `+"`yaml:\"server\"`"+`
}

type Server struct {
	Host string `+"`yaml:\"host\"`"+`
}
`)

	root, err := ParseDir(dir, "")
	if err != nil {
		t.Fatalf("ParseDir() returned error: %v", err)
	}
	if root.Name != "CustomRoot" {
		t.Fatalf("root.Name = %q, want CustomRoot", root.Name)
	}
	if _, ok := root.FindChild("server"); !ok {
		t.Fatal("ParseDir() root missing server field")
	}
}

func TestParseDirParsesAlternateSampleWithoutConfigRoot(t *testing.T) {
	t.Parallel()

	root, err := ParseDir(filepath.Join("..", "..", "schemas", "alternate-sample"), "")
	if err != nil {
		t.Fatalf("ParseDir() returned error: %v", err)
	}
	if root.Name != "Workspace" {
		t.Fatalf("root.Name = %q, want Workspace", root.Name)
	}
	if _, ok := root.FindChild("project"); !ok {
		t.Fatal("ParseDir() root missing project field")
	}
	if _, ok := root.FindChild("server"); ok {
		t.Fatal("ParseDir() unexpectedly included external-sample server field")
	}
}

func TestParseDirParsesAlternateSampleWithExplicitRoot(t *testing.T) {
	t.Parallel()

	root, err := ParseDir(filepath.Join("..", "..", "schemas", "alternate-sample"), "Workspace")
	if err != nil {
		t.Fatalf("ParseDir() returned error: %v", err)
	}

	database := mustChild(t, root, "database")
	driver := mustChild(t, database, "driver")
	wantEnum := []string{"postgres", "mysql", "sqlite"}
	if !reflect.DeepEqual(driver.Enum, wantEnum) {
		t.Fatalf("driver.Enum = %#v, want %#v", driver.Enum, wantEnum)
	}
}

func TestParseDirRejectsAmbiguousAutoRoot(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeSourceFile(t, dir, "schema.go", `package sample

type First struct {
	Name string `+"`yaml:\"name\"`"+`
}

type Second struct {
	Enabled bool `+"`yaml:\"enabled\"`"+`
}
`)

	if _, err := ParseDir(dir, ""); err == nil {
		t.Fatal("ParseDir() error = nil, want ambiguous root error")
	}
}

func TestParseDirRejectsMissingDirectory(t *testing.T) {
	t.Parallel()

	if _, err := ParseDir(filepath.Join("..", "..", "schemas", "missing"), "Config"); err == nil {
		t.Fatal("ParseDir() error = nil, want error")
	}
}

func TestParseDirIgnoresTestFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeSourceFile(t, dir, "config.go", `package sample

type Config struct {
	Name string `+"`yaml:\"name\"`"+`
}
`)
	writeSourceFile(t, dir, "ignored_test.go", `package sample

type Ignored struct {
	Bad Missing `+"`yaml:\"bad\"`"+`
}
`)

	root, err := ParseDir(dir, "Config")
	if err != nil {
		t.Fatalf("ParseDir() returned error: %v", err)
	}
	if _, ok := root.FindChild("name"); !ok {
		t.Fatal("ParseDir() did not parse config.go")
	}
}

func TestParseDirRejectsUnsupportedExternalPackageType(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeSourceFile(t, dir, "config.go", `package sample

import "time"

type Config struct {
	Started time.Time `+"`yaml:\"started\"`"+`
}
`)

	if _, err := ParseDir(dir, "Config"); err == nil {
		t.Fatal("ParseDir() error = nil, want error")
	}
}

func TestParseDirRejectsTypeAlias(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeSourceFile(t, dir, "config.go", `package sample

type Name = string

type Config struct {
	Name Name `+"`yaml:\"name\"`"+`
}
`)

	if _, err := ParseDir(dir, "Config"); err == nil {
		t.Fatal("ParseDir() error = nil, want error")
	}
}

func TestParseDirRejectsGenericType(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeSourceFile(t, dir, "config.go", `package sample

type Config[T any] struct {
	Name string `+"`yaml:\"name\"`"+`
}
`)

	if _, err := ParseDir(dir, "Config"); err == nil {
		t.Fatal("ParseDir() error = nil, want error")
	}
}

func TestParseDirRejectsCircularStructReference(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeSourceFile(t, dir, "config.go", `package sample

type Config struct {
	Server Server `+"`yaml:\"server\"`"+`
}

type Server struct {
	Config Config `+"`yaml:\"config\"`"+`
}
`)

	if _, err := ParseDir(dir, "Config"); err == nil {
		t.Fatal("ParseDir() error = nil, want error")
	}
}

func TestParseToolSchemasFS(t *testing.T) {
	t.Parallel()

	sourceFS := fstest.MapFS{
		"tools/gui.go": {
			Data: []byte(`package gui

type AddAccount struct {
	Name string ` + "`yaml:\"Name\"`" + `
	Code string ` + "`yaml:\"Code\"`" + `
}
`),
		},
	}

	toolSchemas, err := ParseToolSchemasFS(sourceFS, "tools")
	if err != nil {
		t.Fatalf("ParseToolSchemasFS() returned error: %v", err)
	}

	addAccount := toolSchemas["gui.AddAccount"]
	if addAccount == nil {
		t.Fatalf("toolSchemas = %#v, want gui.AddAccount", toolSchemas)
	}
	if _, ok := addAccount.FindChild("Name"); !ok {
		t.Fatal("gui.AddAccount missing Name field")
	}
	if _, ok := addAccount.FindChild("Code"); !ok {
		t.Fatal("gui.AddAccount missing Code field")
	}
}

func writeSourceFile(t *testing.T, dir string, name string, content string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("os.WriteFile() returned error: %v", err)
	}
}
