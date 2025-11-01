package main

import (
	"context"
	"embed"
	"logika/internal/app"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	wsManager := app.NewWSManager()
	baseApp := app.NewApp(wsManager)
	bookmarks := app.NewBookmarks(wsManager)

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "logika",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			baseApp.Start(ctx)
		},
		OnShutdown: func(ctx context.Context) {
			baseApp.Shutdown()
		},
		Bind: []interface{}{
			baseApp,
			bookmarks,
		},
		DragAndDrop: &options.DragAndDrop{
			EnableFileDrop: true,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
