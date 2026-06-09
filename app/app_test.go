package app

import (
	"path/filepath"
	"testing"

	"github.com/mryskyj/yaml-editor/internal/completion"
	filex "github.com/mryskyj/yaml-editor/internal/file"
	"github.com/mryskyj/yaml-editor/internal/schema"
)

func TestNewReturnsApp(t *testing.T) {
	t.Parallel()

	got := New()
	if got == nil {
		t.Fatal("New() returned nil")
	}
}

func TestAppSchema(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	root, err := app.Schema()
	if err != nil {
		t.Fatalf("Schema() returned error: %v", err)
	}
	if root.Name != "File" {
		t.Fatalf("root.Name = %q, want File", root.Name)
	}
	if _, ok := root.FindChild("schema_version"); !ok {
		t.Fatal("Schema() missing schema_version field")
	}
	if _, ok := root.FindChild("common"); !ok {
		t.Fatal("Schema() missing common field")
	}
	if _, ok := root.FindChild("scenario"); !ok {
		t.Fatal("Schema() missing scenario field")
	}
}

func TestAppValidateYAML(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	diagnostics, err := app.ValidateYAML("schema_version: v1\nunknown: true\n")
	if err != nil {
		t.Fatalf("ValidateYAML() returned error: %v", err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("ValidateYAML() diagnostics empty, want unknown key diagnostic")
	}
}

func TestAppCompleteYAML(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	candidates, err := app.CompleteYAML("scenario:\n  \n", 2, 3)
	if err != nil {
		t.Fatalf("CompleteYAML() returned error: %v", err)
	}
	if !hasCandidate(candidates, "steps") {
		t.Fatalf("CompleteYAML() candidates = %#v, want steps", candidates)
	}
}

func TestNewWithSchemaSourceAutoDetectsAlternateSample(t *testing.T) {
	t.Parallel()

	app, err := NewWithSchemaSource(filepath.Join("..", "schemas", "alternate-sample"), "")
	if err != nil {
		t.Fatalf("NewWithSchemaSource() returned error: %v", err)
	}

	root, err := app.Schema()
	if err != nil {
		t.Fatalf("Schema() returned error: %v", err)
	}
	if root.Name != "Workspace" {
		t.Fatalf("root.Name = %q, want Workspace", root.Name)
	}
	if _, ok := root.FindChild("project"); !ok {
		t.Fatal("Schema() missing project field")
	}
	if _, ok := root.FindChild("server"); ok {
		t.Fatal("Schema() includes built-in sample server field")
	}
}

func TestNewWithSchemaSourceUsesExplicitAlternateRoot(t *testing.T) {
	t.Parallel()

	app, err := NewWithSchemaSource(filepath.Join("..", "schemas", "alternate-sample"), "Workspace")
	if err != nil {
		t.Fatalf("NewWithSchemaSource() returned error: %v", err)
	}

	candidates, err := app.CompleteYAML("database:\n  \n", 2, 3)
	if err != nil {
		t.Fatalf("CompleteYAML() returned error: %v", err)
	}
	if !hasCandidate(candidates, "driver") {
		t.Fatalf("CompleteYAML() candidates = %#v, want driver", candidates)
	}
}

func TestAppFileOperations(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	document := app.NewDocument()
	if document.Path != "" || document.Content != "" {
		t.Fatalf("NewDocument() = %#v, want empty", document)
	}

	path := filepath.Join(t.TempDir(), "config.yaml")
	content := "server:\n  host: localhost\n"
	if err := app.SaveFile(path, content); err != nil {
		t.Fatalf("SaveFile() returned error: %v", err)
	}

	opened, err := app.OpenFile(path)
	if err != nil {
		t.Fatalf("OpenFile() returned error: %v", err)
	}
	if opened.Content != content {
		t.Fatalf("OpenFile() content = %q, want %q", opened.Content, content)
	}

	recent, err := app.RecentFiles()
	if err != nil {
		t.Fatalf("RecentFiles() returned error: %v", err)
	}
	if len(recent) != 1 || recent[0] != path {
		t.Fatalf("RecentFiles() = %#v, want [%q]", recent, path)
	}
}

func testApp(t *testing.T) *App {
	t.Helper()

	recentPath := filepath.Join(t.TempDir(), "recent.json")
	return NewWithServices(
		filex.NewService(filex.NewRecentStore(recentPath, 10)),
		schema.NewRegistry(),
	)
}

func hasCandidate(candidates []completion.Candidate, name string) bool {
	for _, candidate := range candidates {
		if candidate.Name == name {
			return true
		}
	}
	return false
}
