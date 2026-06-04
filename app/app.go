package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mryskyj/yaml-editor/internal/completion"
	filex "github.com/mryskyj/yaml-editor/internal/file"
	"github.com/mryskyj/yaml-editor/internal/schema"
	"github.com/mryskyj/yaml-editor/internal/validator"
)

// App exposes application operations to the UI layer.
type App struct {
	files    *filex.Service
	registry *schema.Registry
}

// New creates an application service instance.
func New() *App {
	return NewWithSchemaOptions(StartupSchemaOptions{})
}

// NewWithSchemaOptions creates an application service with startup schema options.
func NewWithSchemaOptions(options StartupSchemaOptions) *App {
	recentPath := filepath.Join(userConfigDir(), "yaml-struct-editor", "recent.json")
	return NewWithServicesAndSchemaOptions(
		filex.NewService(filex.NewRecentStore(recentPath, 10)),
		schema.NewRegistry(),
		options,
	)
}

// NewWithServices creates an application service with injected dependencies.
func NewWithServices(files *filex.Service, registry *schema.Registry) *App {
	return NewWithServicesAndSchemaOptions(files, registry, StartupSchemaOptions{})
}

// NewWithServicesAndSchemaOptions creates an application service with injected dependencies and startup schema options.
func NewWithServicesAndSchemaOptions(files *filex.Service, registry *schema.Registry, options StartupSchemaOptions) *App {
	app := &App{
		files:    files,
		registry: registry,
	}
	if app.registry != nil {
		_ = registerStartupSchema(app.registry, options)
	}
	return app
}

// NewDocument returns an empty unsaved YAML document.
func (a *App) NewDocument() filex.Document {
	if a == nil || a.files == nil {
		return filex.Document{}
	}
	return a.files.NewDocument()
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

// ValidateYAML validates YAML content against the registered schema.
func (a *App) ValidateYAML(content string) ([]validator.Diagnostic, error) {
	root, err := a.rootSchema()
	if err != nil {
		return nil, err
	}
	return validator.Validate(content, root), nil
}

// CompleteYAML returns completion candidates for a cursor position.
func (a *App) CompleteYAML(content string, line int, column int) ([]completion.Candidate, error) {
	root, err := a.rootSchema()
	if err != nil {
		return nil, err
	}
	return completion.Provide(content, line, column, root), nil
}

// Schema returns the registered root schema.
func (a *App) Schema() (*schema.Field, error) {
	return a.rootSchema()
}

func (a *App) rootSchema() (*schema.Field, error) {
	if a == nil || a.registry == nil {
		return nil, fmt.Errorf("schema registry is not configured")
	}
	return a.registry.Root()
}

func userConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return os.TempDir()
	}
	return dir
}
