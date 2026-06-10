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
		Steps []struct {
			ID     string `yaml:"id"`
			Name   string `yaml:"name"`
			Action struct {
				Tool string            `yaml:"tool"`
				Args map[string]string `yaml:"args"`
			} `yaml:"action"`
		} `yaml:"steps"`
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
	want := []string{"server", "app", "steps"}
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

func TestProvideSliceItemCandidates(t *testing.T) {
	t.Parallel()

	source := "steps:\n  - id: first\n    \n"
	candidates := Provide(source, 3, 5, completionSchema(t))
	names := candidateNames(candidates)
	want := []string{"name", "action"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("candidate names = %#v, want %#v", names, want)
	}
}

func TestProvideNestedSliceItemCandidates(t *testing.T) {
	t.Parallel()

	source := "steps:\n  - id: first\n    action:\n      \n"
	candidates := Provide(source, 4, 7, completionSchema(t))
	names := candidateNames(candidates)
	want := []string{"tool", "args"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("candidate names = %#v, want %#v", names, want)
	}
}

func TestProvideToolValueCandidates(t *testing.T) {
	t.Parallel()

	candidates := ProvideWithTools(
		"steps:\n  - action:\n      tool: \n",
		3,
		13,
		completionSchema(t),
		completionToolSchemas(t),
	)
	names := candidateNames(candidates)
	want := []string{"gui."}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("candidate names = %#v, want %#v", names, want)
	}
}

func TestProvideToolStructCandidatesAfterPackage(t *testing.T) {
	t.Parallel()

	candidates := ProvideWithTools(
		"steps:\n  - action:\n      tool: \"gui.\n",
		3,
		18,
		completionSchema(t),
		completionToolSchemas(t),
	)
	names := candidateNames(candidates)
	want := []string{"AddAccount"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("candidate names = %#v, want %#v", names, want)
	}
}

func TestProvideToolCandidatesAcrossNestedNamespaces(t *testing.T) {
	t.Parallel()

	toolSchemas := completionToolSchemas(t)
	toolSchemas["cloud.ecs.RunTask"] = toolSchemas["gui.AddAccount"]

	candidates := ProvideWithTools(
		"steps:\n  - action:\n      tool: \"cloud.\n",
		3,
		20,
		completionSchema(t),
		toolSchemas,
	)
	names := candidateNames(candidates)
	want := []string{"ecs."}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("candidate names = %#v, want %#v", names, want)
	}

	candidates = ProvideWithTools(
		"steps:\n  - action:\n      tool: \"cloud.ecs.\n",
		3,
		24,
		completionSchema(t),
		toolSchemas,
	)
	names = candidateNames(candidates)
	want = []string{"RunTask"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("candidate names = %#v, want %#v", names, want)
	}
}

func TestProvideToolArgsCandidates(t *testing.T) {
	t.Parallel()

	candidates := ProvideWithTools(
		"steps:\n  - action:\n      tool: \"gui.AddAccount\"\n      args:\n        \n",
		5,
		9,
		completionSchema(t),
		completionToolSchemas(t),
	)
	names := candidateNames(candidates)
	want := []string{"Name", "Code", "Tags", "Metadata", "Contacts"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("candidate names = %#v, want %#v", names, want)
	}
}

func TestProvideCandidatesIncludeNestedCollectionSchema(t *testing.T) {
	t.Parallel()

	candidates := ProvideWithTools(
		"steps:\n  - action:\n      tool: \"gui.AddAccount\"\n      args:\n        \n",
		5,
		9,
		completionSchema(t),
		completionToolSchemas(t),
	)

	var contacts Candidate
	for _, candidate := range candidates {
		if candidate.Name == "Contacts" {
			contacts = candidate
			break
		}
	}
	if contacts.Name == "" {
		t.Fatalf("candidates = %#v, want Contacts", candidates)
	}
	if contacts.Type != schema.FieldTypeSlice {
		t.Fatalf("Contacts.Type = %q, want slice", contacts.Type)
	}
	if contacts.Item == nil || contacts.Item.Type != schema.FieldTypeStruct {
		t.Fatalf("Contacts.Item = %#v, want struct item", contacts.Item)
	}
	names := candidateNames(contacts.Item.Children)
	want := []string{"Name", "Email"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("Contacts.Item.Children = %#v, want %#v", names, want)
	}
}

func TestProvideEnumValueCandidates(t *testing.T) {
	t.Parallel()

	candidates := Provide("app:\n  mode: \n", 2, 9, completionSchema(t))
	names := candidateNames(candidates)
	want := []string{"dev", "stg", "prod"}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("candidate names = %#v, want %#v", names, want)
	}
}

func TestProvideNilSchema(t *testing.T) {
	t.Parallel()

	candidates := Provide("", 1, 1, nil)
	if len(candidates) != 0 {
		t.Fatalf("candidates = %#v, want none", candidates)
	}
}

func completionToolSchemas(t *testing.T) map[string]*schema.Field {
	t.Helper()

	type addAccountContact struct {
		Name  string `yaml:"Name"`
		Email string `yaml:"Email"`
	}
	type addAccount struct {
		Name     string              `yaml:"Name"`
		Code     string              `yaml:"Code"`
		Tags     []string            `yaml:"Tags"`
		Metadata map[string]string   `yaml:"Metadata"`
		Contacts []addAccountContact `yaml:"Contacts"`
	}

	field, err := schema.Parse(addAccount{})
	if err != nil {
		t.Fatalf("schema.Parse() returned error: %v", err)
	}
	return map[string]*schema.Field{"gui.AddAccount": field}
}

func candidateNames(candidates []Candidate) []string {
	names := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		names = append(names, candidate.Name)
	}
	return names
}
