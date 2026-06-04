package file

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestServiceNewDocument(t *testing.T) {
	t.Parallel()

	service := NewService(nil)
	document := service.NewDocument()
	if document.Path != "" || document.Content != "" {
		t.Fatalf("NewDocument() = %#v, want empty document", document)
	}
}

func TestServiceSaveAndOpenUTF8(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	recent := NewRecentStore(filepath.Join(dir, "recent.json"), 10)
	service := NewService(recent)
	path := filepath.Join(dir, "config.yaml")
	content := "server:\n  host: ローカル\n"

	if err := service.Save(path, content); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	document, err := service.Open(path)
	if err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}
	if document.Path != path {
		t.Fatalf("document.Path = %q, want %q", document.Path, path)
	}
	if document.Content != content {
		t.Fatalf("document.Content = %q, want %q", document.Content, content)
	}

	recentFiles, err := service.RecentFiles()
	if err != nil {
		t.Fatalf("RecentFiles() returned error: %v", err)
	}
	if !reflect.DeepEqual(recentFiles, []string{path}) {
		t.Fatalf("RecentFiles() = %#v, want %#v", recentFiles, []string{path})
	}
}

func TestServiceOpenRejectsInvalidUTF8(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.yaml")
	if err := os.WriteFile(path, []byte{0xff, 0xfe}, 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	service := NewService(nil)
	if _, err := service.Open(path); err == nil {
		t.Fatal("Open() error = nil, want error")
	}
}

func TestServiceSaveRejectsInvalidUTF8(t *testing.T) {
	t.Parallel()

	service := NewService(nil)
	path := filepath.Join(t.TempDir(), "invalid.yaml")
	if err := service.Save(path, string([]byte{0xff, 0xfe})); err == nil {
		t.Fatal("Save() error = nil, want error")
	}
}
