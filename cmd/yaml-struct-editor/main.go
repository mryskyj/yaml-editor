package main

import (
	"flag"
	"log"
	"os"

	appservice "github.com/mryskyj/yaml-editor/app"
	"github.com/mryskyj/yaml-editor/frontend"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func main() {
	schemaOptions := parseSchemaOptions(os.Args[1:])
	service := appservice.NewWithSchemaOptions(schemaOptions)
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

func parseSchemaOptions(args []string) appservice.StartupSchemaOptions {
	flags := flag.NewFlagSet("yaml-struct-editor", flag.ExitOnError)
	schemaDir := flags.String("schema-dir", "", "directory containing Go schema source files")
	schemaType := flags.String("schema-type", "", "root schema struct type name")
	flags.Parse(args)

	options := appservice.StartupSchemaOptions{
		Dir:  *schemaDir,
		Type: *schemaType,
	}
	if options.Dir == "" && flags.NArg() > 0 {
		options.Dir = flags.Arg(0)
	}
	return options
}
