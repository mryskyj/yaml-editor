package completion

import (
	"reflect"
	"testing"

	"github.com/mryskyj/yaml-editor/internal/schema"
)

func completionSchema(t *testing.T) *schema.Field {
	t.Helper()

	type config struct {
		Server struct {
			Host string `yaml:"host" required:"true" desc:"listen host" default:"localhost"`
			Port int    `yaml:"port" required:"true" default:"8080"`
		} `yaml:"server"`
		App struct {
			Mode string `yaml:"mode" enum:"dev,stg,prod" desc:"run mode"`
		} `yaml:"app"`
	}

	root, err := schema.Parse(config{})
	if err != nil {
		t.Fatalf("schema.Parse() returned error: %v", err)
	}
	return root
}

func TestProvideRootCandidates(t *testing.T) {
	t.Parallel()

	candidates := Provide("", 1, 1, completionSchema(t))
	names := candidateNames(candidates)
	want := []string{"server", "app"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("candidate names = %#v, want %#v", names, want)
	}
}

func TestProvideNestedCandidates(t *testing.T) {
	t.Parallel()

	source := "server:\n  \n"
	candidates := Provide(source, 2, 3, completionSchema(t))
	names := candidateNames(candidates)
	want := []string{"host", "port"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("candidate names = %#v, want %#v", names, want)
	}
}

func TestProvideExcludesExistingKeys(t *testing.T) {
	t.Parallel()

	source := "server:\n  host: localhost\n  \n"
	candidates := Provide(source, 3, 3, completionSchema(t))
	names := candidateNames(candidates)
	want := []string{"port"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("candidate names = %#v, want %#v", names, want)
	}
}

func TestProvideIncludesMetadata(t *testing.T) {
	t.Parallel()

	candidates := Provide("app:\n  \n", 2, 3, completionSchema(t))
	if len(candidates) != 1 {
		t.Fatalf("candidates count = %d, want 1", len(candidates))
	}

	candidate := candidates[0]
	if candidate.Name != "mode" {
		t.Fatalf("Name = %q, want mode", candidate.Name)
	}
	if candidate.Type != schema.FieldTypeString {
		t.Fatalf("Type = %q, want string", candidate.Type)
	}
	if candidate.Description != "run mode" {
		t.Fatalf("Description = %q", candidate.Description)
	}
	wantEnum := []string{"dev", "stg", "prod"}
	if !reflect.DeepEqual(candidate.Enum, wantEnum) {
		t.Fatalf("Enum = %#v, want %#v", candidate.Enum, wantEnum)
	}
}

func TestProvideNilSchema(t *testing.T) {
	t.Parallel()

	candidates := Provide("", 1, 1, nil)
	if len(candidates) != 0 {
		t.Fatalf("candidates = %#v, want none", candidates)
	}
}

func candidateNames(candidates []Candidate) []string {
	names := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		names = append(names, candidate.Name)
	}
	return names
}
