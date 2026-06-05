package completion

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mryskyj/yaml-editor/internal/schema"
)

func TestProvideIncludesExternalSourceNestedCandidates(t *testing.T) {
	t.Parallel()

	root, err := schema.ParseDir(filepath.Join("..", "..", "schemas", "external-sample"), "Config")
	if err != nil {
		t.Fatalf("schema.ParseDir() returned error: %v", err)
	}

	candidates := Provide("server:\n  \n", 2, 3, root)
	names := candidateNames(candidates)
	want := []string{"host", "port"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("candidate names = %#v, want %#v", names, want)
	}
}
