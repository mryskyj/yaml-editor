package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	appservice "github.com/mryskyj/yaml-editor/app"
	"github.com/mryskyj/yaml-editor/frontend"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type startupOptions struct {
	schemaDir  string
	schemaType string
}

func main() {
	options, err := parseStartupOptions(os.Args[1:])
	sanitizeProcessArgsForWails()
	startupDiagnostics := make([]string, 0)
	if err != nil {
		startupDiagnostics = append(startupDiagnostics, fmt.Sprintf("startup option error: %v", err))
	}

	service := appservice.New()
	if err == nil {
		configuredService, configureErr := appservice.NewWithSchemaSource(options.schemaDir, options.schemaType)
		if configureErr != nil {
			startupDiagnostics = append(startupDiagnostics, fmt.Sprintf("schema option error: %v", configureErr))
		} else {
			service = configuredService
		}
	}
	appservice.WithStartupDiagnostics(service, startupDiagnostics)

	wailsApp := application.New(application.Options{
		Name:        "YAML Struct Editor",
		Description: "YAML editor powered by Go struct schemas",
		Services: []application.Service{
			application.NewService(service),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(frontend.Assets()),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	window := wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "YAML Struct Editor",
		Width:            1200,
		Height:           800,
		MinWidth:         900,
		MinHeight:        600,
		BackgroundColour: application.NewRGB(255, 255, 255),
		URL:              "/",
	})
	window.Center()
	window.Show()

	if err := wailsApp.Run(); err != nil {
		log.Fatal(err)
	}
}

func parseStartupOptions(args []string) (startupOptions, error) {
	var options startupOptions
	flags := flag.NewFlagSet("yaml-struct-editor", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&options.schemaDir, "schema-dir", "", "directory containing Go source schema files")
	flags.StringVar(&options.schemaType, "schema-type", "", "root schema struct name; auto-detected when omitted")
	if err := flags.Parse(args); err != nil {
		return options, err
	}
	if remaining := flags.Args(); len(remaining) > 0 {
		return options, fmt.Errorf("unexpected argument: %s", strings.Join(remaining, " "))
	}
	return options, nil
}

func sanitizeProcessArgsForWails() {
	if len(os.Args) <= 1 {
		return
	}
	os.Args = os.Args[:1]
}
