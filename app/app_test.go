package app

import (
	"os"
	"path/filepath"
	"strings"
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
	if root.Name != "ToolSchemas" {
		t.Fatalf("root.Name = %q, want ToolSchemas", root.Name)
	}
	if _, ok := root.FindChild("schema_version"); ok {
		t.Fatal("Schema() includes root schema field schema_version")
	}
	if _, ok := root.FindChild("common"); ok {
		t.Fatal("Schema() includes root schema field common")
	}
	if _, ok := root.FindChild("scenario"); ok {
		t.Fatal("Schema() includes root schema field scenario")
	}
	if _, ok := root.FindChild("gui.AddAccount"); !ok {
		t.Fatal("Schema() missing gui.AddAccount tool schema")
	}
	if _, ok := root.FindChild("gui.AddAccounts"); !ok {
		t.Fatal("Schema() missing gui.AddAccounts tool schema")
	}
	if _, ok := root.FindChild("cloud.ecs.RunTask"); !ok {
		t.Fatal("Schema() missing cloud.ecs.RunTask tool schema")
	}
	if _, ok := root.FindChild("sampleschema.Config"); !ok {
		t.Fatal("Schema() missing sampleschema.Config tool schema")
	}
}

func TestAppRootSchemaReturnsDocumentSchema(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	root, err := app.RootSchema()
	if err != nil {
		t.Fatalf("RootSchema() returned error: %v", err)
	}
	if root.Name != "File" {
		t.Fatalf("root.Name = %q, want File", root.Name)
	}
	if _, ok := root.FindChild("scenario"); !ok {
		t.Fatal("RootSchema() missing scenario field")
	}

	scenario := mustSchemaChild(t, root, "scenario")
	description := mustSchemaChild(t, scenario, "description")
	if description.Required {
		t.Fatal("scenario.description Required = true, want false")
	}
	docs := mustSchemaChild(t, scenario, "docs")
	if docs.Required {
		t.Fatal("scenario.docs Required = true, want false")
	}
	steps := mustSchemaChild(t, scenario, "steps")
	if !steps.Required {
		t.Fatal("scenario.steps Required = false, want true")
	}
	dayRef := mustSchemaChild(t, steps.Item, "day_ref")
	if dayRef.Required {
		t.Fatal("step.day_ref Required = true, want false")
	}
	scheduleRef := mustSchemaChild(t, steps.Item, "schedule_ref")
	if scheduleRef.Required {
		t.Fatal("step.schedule_ref Required = true, want false")
	}
	action := mustSchemaChild(t, steps.Item, "action")
	args := mustSchemaChild(t, action, "args")
	if args.Required {
		t.Fatal("action.args Required = true, want false")
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

func TestAppCompleteYAMLToolValue(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	candidates, err := app.CompleteYAML("scenario:\n  steps:\n    - action:\n        tool: \n", 4, 15)
	if err != nil {
		t.Fatalf("CompleteYAML() returned error: %v", err)
	}
	if !hasCandidate(candidates, "sampleschema.") {
		t.Fatalf("CompleteYAML() candidates = %#v, want sampleschema.", candidates)
	}
	if !hasCandidate(candidates, "cloud.") {
		t.Fatalf("CompleteYAML() candidates = %#v, want cloud.", candidates)
	}
	if !hasCandidate(candidates, "gui.") {
		t.Fatalf("CompleteYAML() candidates = %#v, want gui.", candidates)
	}

	candidates, err = app.CompleteYAML("scenario:\n  steps:\n    - action:\n        tool: \"sam\n", 4, 19)
	if err != nil {
		t.Fatalf("CompleteYAML() returned error: %v", err)
	}
	if !hasCandidate(candidates, "sampleschema.") {
		t.Fatalf("CompleteYAML() candidates = %#v, want sampleschema.", candidates)
	}

	candidates, err = app.CompleteYAML("scenario:\n  steps:\n    - action:\n        tool: \"sampleschema.\n", 4, 29)
	if err != nil {
		t.Fatalf("CompleteYAML() returned error: %v", err)
	}
	if !hasCandidate(candidates, "Config") {
		t.Fatalf("CompleteYAML() candidates = %#v, want Config", candidates)
	}

	candidates, err = app.CompleteYAML("scenario:\n  steps:\n    - action:\n        tool: \"cloud.\n", 4, 22)
	if err != nil {
		t.Fatalf("CompleteYAML() returned error: %v", err)
	}
	if !hasCandidate(candidates, "ecs.") {
		t.Fatalf("CompleteYAML() candidates = %#v, want ecs.", candidates)
	}

	candidates, err = app.CompleteYAML("scenario:\n  steps:\n    - action:\n        tool: \"cloud.ecs.\n", 4, 26)
	if err != nil {
		t.Fatalf("CompleteYAML() returned error: %v", err)
	}
	if !hasCandidate(candidates, "RunTask") {
		t.Fatalf("CompleteYAML() candidates = %#v, want RunTask", candidates)
	}
}

func TestAppCompleteYAMLToolArgs(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	candidates, err := app.CompleteYAML("scenario:\n  steps:\n    - action:\n        tool: \"sampleschema.Config\"\n        args:\n          \n", 6, 11)
	if err != nil {
		t.Fatalf("CompleteYAML() returned error: %v", err)
	}
	if !hasCandidate(candidates, "server") {
		t.Fatalf("CompleteYAML() candidates = %#v, want server", candidates)
	}
	if !hasCandidate(candidates, "app") {
		t.Fatalf("CompleteYAML() candidates = %#v, want app", candidates)
	}
}

func TestAppCompleteYAMLToolArgsIncludesSliceItemSchema(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	candidates, err := app.CompleteYAML("scenario:\n  steps:\n    - action:\n        tool: \"gui.AddAccount\"\n        args:\n          \n", 6, 11)
	if err != nil {
		t.Fatalf("CompleteYAML() returned error: %v", err)
	}

	var contacts completion.Candidate
	for _, candidate := range candidates {
		if candidate.Name == "Contacts" {
			contacts = candidate
			break
		}
	}
	if contacts.Name == "" {
		t.Fatalf("CompleteYAML() candidates = %#v, want Contacts", candidates)
	}
	if contacts.Type != schema.FieldTypeSlice {
		t.Fatalf("Contacts.Type = %q, want slice", contacts.Type)
	}
	if contacts.Item == nil || contacts.Item.Type != schema.FieldTypeStruct {
		t.Fatalf("Contacts.Item = %#v, want struct item", contacts.Item)
	}
	if _, ok := findCandidate(contacts.Item.Children, "Name"); !ok {
		t.Fatalf("Contacts.Item.Children = %#v, want Name", contacts.Item.Children)
	}
	if _, ok := findCandidate(contacts.Item.Children, "Email"); !ok {
		t.Fatalf("Contacts.Item.Children = %#v, want Email", contacts.Item.Children)
	}
}

func TestAppCompleteYAMLToolArgsRootSliceSchema(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	toolCandidates, err := app.CompleteYAML("scenario:\n  steps:\n    - action:\n        tool: \"gui.\n", 4, 20)
	if err != nil {
		t.Fatalf("CompleteYAML() returned error: %v", err)
	}
	if !hasCandidate(toolCandidates, "AddAccounts") {
		t.Fatalf("CompleteYAML() candidates = %#v, want AddAccounts", toolCandidates)
	}

	candidates, err := app.CompleteYAML("scenario:\n  steps:\n    - action:\n        tool: \"gui.AddAccounts\"\n        args:\n          \n", 6, 11)
	if err != nil {
		t.Fatalf("CompleteYAML() returned error: %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("CompleteYAML() candidates = %#v, want one root slice candidate", candidates)
	}
	candidate := candidates[0]
	if !candidate.Root {
		t.Fatalf("Root = false, want true")
	}
	if candidate.Type != schema.FieldTypeSlice {
		t.Fatalf("Type = %q, want slice", candidate.Type)
	}
	if candidate.Item == nil || candidate.Item.Type != schema.FieldTypeStruct {
		t.Fatalf("Item = %#v, want struct item", candidate.Item)
	}
	if _, ok := findCandidate(candidate.Item.Children, "Name"); !ok {
		t.Fatalf("Item.Children = %#v, want Name", candidate.Item.Children)
	}
}

func TestAppValidateYAMLToolArgs(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	diagnostics, err := app.ValidateYAML(`
schema_version: v1
common:
  schema_version: v1
  dates: {}
  number_of_days: 1
  schedules: {}
scenario:
  id: 1
  name: test
  description: test
  docs: []
  steps:
    - id: "101-02"
      name: test
      day_ref: day1
      schedule_ref: run8
      action:
        tool: "sampleschema.Config"
        args:
          server:
            host: 127.0.0.1
            port: 8080
          app:
            mode: dev
`)
	if err != nil {
		t.Fatalf("ValidateYAML() returned error: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("ValidateYAML() diagnostics = %#v, want none", diagnostics)
	}
}

func TestAppValidateYAMLToolArgsRootSlice(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	diagnostics, err := app.ValidateYAML(`
schema_version: v1
common:
  schema_version: v1
  dates: {}
  number_of_days: 1
  schedules: {}
scenario:
  id: 1
  name: test
  description: test
  docs: []
  steps:
    - id: "101-02"
      name: test
      day_ref: day1
      schedule_ref: run8
      action:
        tool: "gui.AddAccounts"
        args:
          - Name: aiu
            Code: "11111"
            Kind: standard
`)
	if err != nil {
		t.Fatalf("ValidateYAML() returned error: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("ValidateYAML() diagnostics = %#v, want none", diagnostics)
	}
}

func TestAppSchemaCommonDatesUseDayDateHolidayStructure(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	root, err := app.rootSchema()
	if err != nil {
		t.Fatalf("rootSchema() returned error: %v", err)
	}

	common, ok := root.FindChild("common")
	if !ok {
		t.Fatal("Schema() missing common field")
	}
	commonSchemaVersion := mustSchemaChild(t, common, "schema_version")
	if commonSchemaVersion.Required {
		t.Fatal("common.schema_version Required = true, want false")
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
	numberOfDays := mustSchemaChild(t, common, "number_of_days")
	if numberOfDays.Type != schema.FieldTypeInt {
		t.Fatalf("common.number_of_days Type = %q, want int", numberOfDays.Type)
	}
	if !numberOfDays.Required {
		t.Fatal("common.number_of_days Required = false, want true")
	}
}

func TestAppSchemaCommonSchedulesUseRunScalarStructure(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	root, err := app.rootSchema()
	if err != nil {
		t.Fatalf("rootSchema() returned error: %v", err)
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
  number_of_days: 1
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

func TestAppValidateYAMLCommonInclude(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	dir := t.TempDir()
	commonPath := filepath.Join(dir, "common", "common.yaml")
	writeTextFile(t, commonPath, includeCommonYAML)
	mainPath := filepath.Join(dir, "scenario", "config.yaml")

	diagnostics, err := app.ValidateYAMLForPath(`
schema_version: "1.0.0"
common: !include "../common/common.yaml"
scenario:
  id: 1
  name: test
  steps: []
`, mainPath)
	if err != nil {
		t.Fatalf("ValidateYAMLForPath() returned error: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("ValidateYAMLForPath() diagnostics = %#v, want none", diagnostics)
	}
}

func TestAppValidateYAMLCommonIncludeAcceptsWrappedCommon(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	dir := t.TempDir()
	commonPath := filepath.Join(dir, "common.yaml")
	writeTextFile(t, commonPath, `common:
  dates:
    day1:
      date: "2026-03-01"
      holiday: false
  number_of_days: 1
  schedules:
    run1: &run1 1 #BOD
`)

	diagnostics, err := app.ValidateYAMLForPath(`
schema_version: "1.0.0"
common: !include "common.yaml"
scenario:
  id: 1
  name: test
  steps: []
`, filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatalf("ValidateYAMLForPath() returned error: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("ValidateYAMLForPath() diagnostics = %#v, want none", diagnostics)
	}
}

func TestAppValidateYAMLCommonIncludeDiagnostics(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	diagnostics, err := app.ValidateYAML(`
schema_version: "1.0.0"
common: !include "common.yaml"
scenario:
  id: 1
  name: test
  steps: []
`)
	if err != nil {
		t.Fatalf("ValidateYAML() returned error: %v", err)
	}
	if len(diagnostics) == 0 || !strings.Contains(diagnostics[0].Message, "unsaved file") {
		t.Fatalf("ValidateYAML() diagnostics = %#v, want unsaved include diagnostic", diagnostics)
	}

	diagnostics, err = app.ValidateYAMLForPath(`
schema_version: "1.0.0"
common: !include "../../common.yaml"
scenario:
  id: 1
  name: test
  steps: []
`, filepath.Join(t.TempDir(), "scenario", "config.yaml"))
	if err != nil {
		t.Fatalf("ValidateYAMLForPath() returned error: %v", err)
	}
	if len(diagnostics) != 1 || !strings.Contains(diagnostics[0].Message, "cannot be read") {
		t.Fatalf("ValidateYAMLForPath() diagnostics = %#v, want one include read diagnostic", diagnostics)
	}

	diagnostics, err = app.ValidateYAMLForPath(`
schema_version: "1.0.0"
common: !include "/tmp/common.yaml"
scenario:
  id: 1
  name: test
  steps: []
`, filepath.Join(t.TempDir(), "config.yaml"))
	if err != nil {
		t.Fatalf("ValidateYAMLForPath() returned error: %v", err)
	}
	if len(diagnostics) == 0 || !strings.Contains(diagnostics[0].Message, "relative") {
		t.Fatalf("ValidateYAMLForPath() diagnostics = %#v, want relative path diagnostic", diagnostics)
	}

	dir := t.TempDir()
	writeTextFile(t, filepath.Join(dir, "common.yaml"), `dates: !include "dates.yaml"
number_of_days: 1
schedules: {}
`)
	diagnostics, err = app.ValidateYAMLForPath(`
schema_version: "1.0.0"
common: !include "common.yaml"
scenario:
  id: 1
  name: test
  steps: []
`, filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatalf("ValidateYAMLForPath() returned error: %v", err)
	}
	if len(diagnostics) == 0 || !strings.Contains(diagnostics[0].Message, "nested !include") {
		t.Fatalf("ValidateYAMLForPath() diagnostics = %#v, want nested include diagnostic", diagnostics)
	}
}

func TestAppCompleteYAMLReferencesFromCommonInclude(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	dir := t.TempDir()
	commonPath := filepath.Join(dir, "common", "common.yaml")
	writeTextFile(t, commonPath, includeCommonYAML)
	mainPath := filepath.Join(dir, "scenario", "config.yaml")
	source := `
schema_version: "1.0.0"
common: !include "../common/common.yaml"
scenario:
  id: 1
  name: test
  steps:
    - id: "101-02"
      name: step
      day_ref:
      schedule_ref:
      action:
        tool: "gui.AddAccount"
`

	dayCandidates, err := app.CompleteYAMLForPath(source, 10, 16, mainPath)
	if err != nil {
		t.Fatalf("CompleteYAMLForPath() returned error: %v", err)
	}
	day1, ok := findCandidate(dayCandidates, "day1")
	if !ok {
		t.Fatalf("CompleteYAMLForPath() candidates = %#v, want day1", dayCandidates)
	}
	if !strings.Contains(day1.Description, "2026-03-01") || !strings.Contains(day1.Description, "holiday: false") {
		t.Fatalf("day1.Description = %q, want date and holiday", day1.Description)
	}

	scheduleCandidates, err := app.CompleteYAMLForPath(source, 11, 21, mainPath)
	if err != nil {
		t.Fatalf("CompleteYAMLForPath() returned error: %v", err)
	}
	run2, ok := findCandidate(scheduleCandidates, "run2")
	if !ok {
		t.Fatalf("CompleteYAMLForPath() candidates = %#v, want run2", scheduleCandidates)
	}
	if !strings.Contains(run2.Description, "value: 2") || !strings.Contains(run2.Description, "deploy") {
		t.Fatalf("run2.Description = %q, want value and comment", run2.Description)
	}
}

func TestAppCompleteYAMLCommonIncludeValue(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	candidates, err := app.CompleteYAML("common: \n", 1, 9)
	if err != nil {
		t.Fatalf("CompleteYAML() returned error: %v", err)
	}
	candidate, ok := findCandidate(candidates, `!include ""`)
	if !ok {
		t.Fatalf("CompleteYAML() candidates = %#v, want !include", candidates)
	}
	if candidate.InsertText != `!include "$0"` {
		t.Fatalf("InsertText = %q, want snippet include", candidate.InsertText)
	}
}

func TestNewWithSchemaSourceAutoDetectsAlternateSample(t *testing.T) {
	t.Parallel()

	app, err := NewWithSchemaSource(filepath.Join("..", "schemas", "alternate-sample"), "")
	if err != nil {
		t.Fatalf("NewWithSchemaSource() returned error: %v", err)
	}

	root, err := app.rootSchema()
	if err != nil {
		t.Fatalf("rootSchema() returned error: %v", err)
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

func TestAppSchemaWithExternalSourceShowsExternalToolSchemas(t *testing.T) {
	t.Parallel()

	app, err := NewWithSchemaSource(filepath.Join("..", "schemas", "alternate-sample"), "")
	if err != nil {
		t.Fatalf("NewWithSchemaSource() returned error: %v", err)
	}

	root, err := app.Schema()
	if err != nil {
		t.Fatalf("Schema() returned error: %v", err)
	}
	if root.Name != "ToolSchemas" {
		t.Fatalf("root.Name = %q, want ToolSchemas", root.Name)
	}
	if _, ok := root.FindChild("sample.Project"); !ok {
		t.Fatalf("Schema() = %#v, want sample.Project", root)
	}
	if _, ok := root.FindChild("common"); ok {
		t.Fatal("Schema() includes root schema field common")
	}
	if _, ok := root.FindChild("gui.AddAccount"); ok {
		t.Fatal("Schema() includes built-in sample tool schema")
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

func TestNewWithSchemaSourceUsesExternalToolSchemas(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeAppSourceFile(t, dir, "config.go", `package externaltool

type Config struct {
	Steps []Step `+"`yaml:\"steps\"`"+`
}

type Step struct {
	Action Action `+"`yaml:\"action\"`"+`
}

type Action struct {
	Tool string            `+"`yaml:\"tool\"`"+`
	Args map[string]string `+"`yaml:\"args\"`"+`
}

type Database struct {
	Driver string `+"`yaml:\"driver\"`"+`
	Host   string `+"`yaml:\"host\"`"+`
}
`)

	app, err := NewWithSchemaSource(dir, "Config")
	if err != nil {
		t.Fatalf("NewWithSchemaSource() returned error: %v", err)
	}

	toolCandidates, err := app.CompleteYAML("steps:\n  - action:\n      tool: \n", 3, 13)
	if err != nil {
		t.Fatalf("CompleteYAML() returned error: %v", err)
	}
	if !hasCandidate(toolCandidates, "externaltool.") {
		t.Fatalf("CompleteYAML() candidates = %#v, want externaltool.", toolCandidates)
	}

	toolCandidates, err = app.CompleteYAML("steps:\n  - action:\n      tool: \"externaltool.\n", 3, 27)
	if err != nil {
		t.Fatalf("CompleteYAML() returned error: %v", err)
	}
	if !hasCandidate(toolCandidates, "Database") {
		t.Fatalf("CompleteYAML() candidates = %#v, want Database", toolCandidates)
	}

	argsCandidates, err := app.CompleteYAML("steps:\n  - action:\n      tool: \"externaltool.Database\"\n      args:\n        \n", 5, 9)
	if err != nil {
		t.Fatalf("CompleteYAML() returned error: %v", err)
	}
	if !hasCandidate(argsCandidates, "driver") {
		t.Fatalf("CompleteYAML() candidates = %#v, want driver", argsCandidates)
	}
}

func TestAppFileOperations(t *testing.T) {
	t.Parallel()

	app := testApp(t)
	document := app.NewDocument()
	if document.Path != "" {
		t.Fatalf("NewDocument().Path = %q, want empty", document.Path)
	}
	if document.Content != requiredRootTemplate {
		t.Fatalf("NewDocument().Content =\n%s\nwant:\n%s", document.Content, requiredRootTemplate)
	}
	if containsAny(document.Content, "description:", "docs:", "day_ref:", "schedule_ref:", "user:", "password:", "path:", "args:") {
		t.Fatalf("NewDocument().Content includes optional key:\n%s", document.Content)
	}
	if got := app.ScheduleTemplate(); got != defaultScheduleTemplate {
		t.Fatalf("ScheduleTemplate() =\n%s\nwant:\n%s", got, defaultScheduleTemplate)
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

const requiredRootTemplate = `schema_version: "1.0.0"
common:
  dates:
    day1:
      date: ""
      holiday: false
  number_of_days: 1
  schedules:
    run1: &run1 1 #BOD
    run2: &run2 2 #あいうえお
    run3: &run3 3 #かきくけこ
scenario:
  id: 0
  name: ""
  steps:
    - id: ""
      name: ""
      action:
        tool: ""
`

const defaultScheduleTemplate = `run1: &run1 1 #BOD
run2: &run2 2 #あいうえお
run3: &run3 3 #かきくけこ`

func containsAny(value string, patterns ...string) bool {
	for _, pattern := range patterns {
		if strings.Contains(value, pattern) {
			return true
		}
	}
	return false
}

func testApp(t *testing.T) *App {
	t.Helper()

	recentPath := filepath.Join(t.TempDir(), "recent.json")
	return NewWithServices(
		filex.NewService(filex.NewRecentStore(recentPath, 10)),
		schema.NewRegistry(),
	)
}

func mustSchemaChild(t *testing.T, root *schema.Field, name string) *schema.Field {
	t.Helper()

	if root == nil {
		t.Fatalf("FindChild(%q) on nil schema field", name)
	}
	child, ok := root.FindChild(name)
	if !ok {
		t.Fatalf("FindChild(%q) did not find child", name)
	}
	return child
}

func writeAppSourceFile(t *testing.T, dir string, name string, content string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}
}

func writeTextFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func hasCandidate(candidates []completion.Candidate, name string) bool {
	_, ok := findCandidate(candidates, name)
	return ok
}

const includeCommonYAML = `dates:
  day1:
    date: "2026-03-01"
    holiday: false
  day2:
    date: "2026-03-02"
    holiday: true
number_of_days: 2
schedules:
  run1: &run1 1 #BOD
  run2: &run2 2 #deploy
`

func findCandidate(candidates []completion.Candidate, name string) (completion.Candidate, bool) {
	for _, candidate := range candidates {
		if candidate.Name == name {
			return candidate, true
		}
	}
	return completion.Candidate{}, false
}
