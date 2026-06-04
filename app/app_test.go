package app

import (
	"path/filepath"
	"testing"

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
	if _, ok := root.FindChild("server"); !ok {
		t.Fatal("Schema() missing server field")
	}
}

func TestAppValidateYAML(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	diagnostics, err := app.ValidateYAML("server:\n  host: localhost\napp:\n  mode: dev\n")
	if err != nil {
		t.Fatalf("ValidateYAML() returned error: %v", err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("ValidateYAML() diagnostics empty, want missing port diagnostic")
	}
}

func TestAppCompleteYAML(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	candidates, err := app.CompleteYAML("server:\n  \n", 2, 3)
	if err != nil {
		t.Fatalf("CompleteYAML() returned error: %v", err)
	}
	if len(candidates) != 2 {
		t.Fatalf("CompleteYAML() candidates count = %d, want 2", len(candidates))
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
