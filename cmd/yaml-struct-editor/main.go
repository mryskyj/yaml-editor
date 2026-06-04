package main

import (
	"log"

	appservice "github.com/mryskyj/yaml-editor/app"
	"github.com/mryskyj/yaml-editor/frontend"
	"github.com/wailsapp/wails/v3/pkg/application"
)

func main() {
	service := appservice.New()
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
