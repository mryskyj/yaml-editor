package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/mryskyj/yaml-editor/internal/completion"
	filex "github.com/mryskyj/yaml-editor/internal/file"
	"github.com/mryskyj/yaml-editor/internal/schema"
	"github.com/mryskyj/yaml-editor/internal/validator"
	"gopkg.in/yaml.v3"
)

// App exposes application operations to the UI layer.
type App struct {
	files        *filex.Service
	registry     *schema.Registry
	rootDefaults *yaml.Node
}

// New creates an application service instance.
func New() *App {
	recentPath := filepath.Join(userConfigDir(), "yaml-struct-editor", "recent.json")
	return NewWithServices(filex.NewService(filex.NewRecentStore(recentPath, 10)), schema.NewRegistry())
}

// NewWithSchemaSource creates an application service using an external Go source schema when configured.
func NewWithSchemaSource(schemaDir string, schemaType string) (*App, error) {
	recentPath := filepath.Join(userConfigDir(), "yaml-struct-editor", "recent.json")
	if schemaDir != "" {
		registry, err := externalSchemaRegistry(schemaDir, schemaType)
		if err != nil {
			return nil, err
		}
		return &App{
			files:        filex.NewService(filex.NewRecentStore(recentPath, 10)),
			registry:     registry,
			rootDefaults: nil,
		}, nil
	}

	return NewWithServices(filex.NewService(filex.NewRecentStore(recentPath, 10)), schema.NewRegistry()), nil
}

// NewWithServices creates an application service with injected dependencies.
func NewWithServices(files *filex.Service, registry *schema.Registry) *App {
	app := &App{
		files:        files,
		registry:     registry,
		rootDefaults: loadBuiltinRootDefaults(),
	}
	if app.registry != nil {
		_ = registerSampleSchema(app.registry)
	}
	return app
}

// NewDocument returns an unsaved YAML document initialized from the root schema.
func (a *App) NewDocument() filex.Document {
	if a == nil || a.files == nil {
		return filex.Document{}
	}

	document := a.files.NewDocument()
	root, err := a.rootSchema()
	if err != nil {
		return document
	}
	document.Content = rootSchemaTemplate(root, a.rootDefaults)
	return document
}

// ScheduleTemplate returns the default schedule template from root defaults.
func (a *App) ScheduleTemplate() string {
	root, err := a.rootSchema()
	if err != nil {
		return ""
	}
	return rootScheduleTemplate(root, a.rootDefaults)
}

// OpenFile opens a UTF-8 YAML file.
func (a *App) OpenFile(path string) (filex.Document, error) {
	if a == nil || a.files == nil {
		return filex.Document{}, fmt.Errorf("file service is not configured")
	}
	return a.files.Open(path)
}

// SaveFile saves UTF-8 YAML text.
func (a *App) SaveFile(path string, content string) error {
	if a == nil || a.files == nil {
		return fmt.Errorf("file service is not configured")
	}
	return a.files.Save(path, content)
}

// RecentFiles returns recently opened files.
func (a *App) RecentFiles() ([]string, error) {
	if a == nil || a.files == nil {
		return nil, nil
	}
	return a.files.RecentFiles()
}

// LoadExternalSchema loads Go source schemas from a user-selected directory.
func (a *App) LoadExternalSchema(schemaDir string) error {
	if a == nil {
		return fmt.Errorf("app service is not configured")
	}
	registry, err := externalSchemaRegistry(schemaDir, "")
	if err != nil {
		return err
	}
	a.registry = registry
	a.rootDefaults = nil
	return nil
}

// ValidateYAML validates YAML content against the registered schema.
func (a *App) ValidateYAML(content string) ([]validator.Diagnostic, error) {
	return a.ValidateYAMLForPath(content, "")
}

// ValidateYAMLForPath validates YAML content and resolves relative includes from the document path.
func (a *App) ValidateYAMLForPath(content string, documentPath string) ([]validator.Diagnostic, error) {
	root, err := a.rootSchema()
	if err != nil {
		return nil, err
	}
	return validator.ValidateWithToolsForPath(content, root, a.toolSchemas(), documentPath), nil
}

// CompleteYAML returns completion candidates for a cursor position.
func (a *App) CompleteYAML(content string, line int, column int) ([]completion.Candidate, error) {
	return a.CompleteYAMLForPath(content, line, column, "")
}

// CompleteYAMLForPath returns completion candidates and resolves relative includes from the document path.
func (a *App) CompleteYAMLForPath(content string, line int, column int, documentPath string) ([]completion.Candidate, error) {
	root, err := a.rootSchema()
	if err != nil {
		return nil, err
	}
	return completion.ProvideWithToolsForPath(content, line, column, root, a.toolSchemas(), documentPath), nil
}

// Schema returns schemas shown in the UI schema pane.
func (a *App) Schema() (*schema.Field, error) {
	toolSchemas := a.toolSchemas()
	if len(toolSchemas) == 0 {
		return &schema.Field{Name: "ToolSchemas", Type: schema.FieldTypeStruct}, nil
	}

	names := make([]string, 0, len(toolSchemas))
	for name := range toolSchemas {
		names = append(names, name)
	}
	sort.Strings(names)

	children := make([]*schema.Field, 0, len(names))
	for _, name := range names {
		field := cloneSchemaField(toolSchemas[name])
		if field == nil {
			continue
		}
		field.Name = name
		children = append(children, field)
	}
	return &schema.Field{Name: "ToolSchemas", Type: schema.FieldTypeStruct, Children: children}, nil
}

// RootSchema returns the registered YAML document schema for context-sensitive UI.
func (a *App) RootSchema() (*schema.Field, error) {
	return a.rootSchema()
}

func (a *App) rootSchema() (*schema.Field, error) {
	if a == nil || a.registry == nil {
		return nil, fmt.Errorf("schema registry is not configured")
	}
	return a.registry.Root()
}

func (a *App) toolSchemas() map[string]*schema.Field {
	if a == nil || a.registry == nil {
		return nil
	}
	return a.registry.ToolSchemas()
}

func cloneSchemaField(field *schema.Field) *schema.Field {
	if field == nil {
		return nil
	}

	clone := *field
	if len(field.Enum) > 0 {
		clone.Enum = append([]string(nil), field.Enum...)
	}
	if len(field.Children) > 0 {
		clone.Children = make([]*schema.Field, 0, len(field.Children))
		for _, child := range field.Children {
			clone.Children = append(clone.Children, cloneSchemaField(child))
		}
	}
	clone.Item = cloneSchemaField(field.Item)
	clone.MapValue = cloneSchemaField(field.MapValue)
	return &clone
}

func externalSchemaRegistry(schemaDir string, schemaType string) (*schema.Registry, error) {
	if schemaDir == "" {
		return nil, fmt.Errorf("schema directory is required")
	}
	registry := schema.NewRegistry()
	if err := registry.RegisterFromDir(schemaDir, schemaType); err != nil {
		return nil, err
	}
	if err := registry.RegisterToolSchemasFromDir(schemaDir); err != nil {
		return nil, err
	}
	return registry, nil
}

func userConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return os.TempDir()
	}
	return dir
}
