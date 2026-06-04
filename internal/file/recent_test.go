package file

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestRecentStoreAddListDeduplicatesAndLimits(t *testing.T) {
	t.Parallel()

	store := NewRecentStore(filepath.Join(t.TempDir(), "recent.json"), 3)
	for _, path := range []string{"a.yaml", "b.yaml", "c.yaml", "b.yaml", "d.yaml"} {
		if err := store.Add(path); err != nil {
			t.Fatalf("Add(%q) returned error: %v", path, err)
		}
	}

	files, err := store.List()
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	want := []string{"d.yaml", "b.yaml", "c.yaml"}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("List() = %#v, want %#v", files, want)
	}
}

func TestRecentStoreListMissingFile(t *testing.T) {
	t.Parallel()

	store := NewRecentStore(filepath.Join(t.TempDir(), "missing", "recent.json"), 10)
	files, err := store.List()
	if err != nil {
		t.Fatalf("List() returned error: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("List() = %#v, want none", files)
	}
}

func TestRecentStoreRejectsEmptyPath(t *testing.T) {
	t.Parallel()

	store := NewRecentStore(filepath.Join(t.TempDir(), "recent.json"), 10)
	if err := store.Add(""); err == nil {
		t.Fatal("Add() error = nil, want error")
	}
}
