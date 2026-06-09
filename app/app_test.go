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

func TestAppCompleteYAMLScenarioStepListItem(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	candidates, err := app.CompleteYAML("scenario:\n  steps:\n    - id: \"101-02\"\n      \n", 4, 7)
	if err != nil {
		t.Fatalf("CompleteYAML() returned error: %v", err)
	}
	if !hasCandidate(candidates, "name") {
		t.Fatalf("CompleteYAML() candidates = %#v, want name", candidates)
	}
	if !hasCandidate(candidates, "day_ref") {
		t.Fatalf("CompleteYAML() candidates = %#v, want day_ref", candidates)
	}
	if hasCandidate(candidates, "id") {
		t.Fatalf("CompleteYAML() candidates = %#v, want id excluded as existing key", candidates)
	}
}

func TestAppSchemaCommonDatesUseDayDateHolidayStructure(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	root, err := app.Schema()
	if err != nil {
		t.Fatalf("Schema() returned error: %v", err)
	}

	common, ok := root.FindChild("common")
	if !ok {
		t.Fatal("Schema() missing common field")
	}
	dates, ok := common.FindChild("dates")
	if !ok {
		t.Fatal("Schema() missing common.dates field")
	}
	if dates.Type != schema.FieldTypeMap {
		t.Fatalf("common.dates Type = %q, want map", dates.Type)
	}
	if dates.MapValue == nil {
		t.Fatal("common.dates MapValue is nil")
	}
	if _, ok := dates.MapValue.FindChild("date"); !ok {
		t.Fatal("common.dates day entry missing date field")
	}
	holiday, ok := dates.MapValue.FindChild("holiday")
	if !ok {
		t.Fatal("common.dates day entry missing holiday field")
	}
	if holiday.Type != schema.FieldTypeBool {
		t.Fatalf("common.dates holiday Type = %q, want bool", holiday.Type)
	}
}

func TestAppSchemaCommonSchedulesUseRunScalarStructure(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	root, err := app.Schema()
	if err != nil {
		t.Fatalf("Schema() returned error: %v", err)
	}

	common, ok := root.FindChild("common")
	if !ok {
		t.Fatal("Schema() missing common field")
	}
	schedules, ok := common.FindChild("schedules")
	if !ok {
		t.Fatal("Schema() missing common.schedules field")
	}
	if schedules.Type != schema.FieldTypeMap {
		t.Fatalf("common.schedules Type = %q, want map", schedules.Type)
	}
	if schedules.MapValue == nil || schedules.MapValue.Type != schema.FieldTypeInt {
		t.Fatalf("common.schedules MapValue = %#v, want int", schedules.MapValue)
	}
}

func TestAppValidateCommonSchedulesAllowsAnchoredRuns(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	diagnostics, err := app.ValidateYAML(`
schema_version: v1
common:
  schema_version: v1
  dates: {}
  schedules:
    run1: &run1 1 #BOD
    run2: &run2 2 #あいうえお
    run3: &run3 3 #かきくけこ
scenario:
  id: 1
  name: test
  description: test
  docs: []
  steps: []
`)
	if err != nil {
		t.Fatalf("ValidateYAML() returned error: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("ValidateYAML() diagnostics = %#v, want none", diagnostics)
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
