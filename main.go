package main

import (
	"context"
	"embed"
	"flag"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

func main() {
	trayStart := flag.Bool("tray", false, "Start minimized in tray")
	flag.Parse()

	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "CensawayApp",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 9, G: 9, B: 11, A: 1},
		StartHidden:      *trayStart,
		OnStartup: func(ctx context.Context) {
			app.Icon = icon
			app.startup(ctx)
			app.SetupTray(ctx)
		},
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			if app.isQuitting {
				return false
			}
			runtime.WindowHide(ctx)
			return true
		},
		Frameless:       true,
		CSSDragProperty: "--wails-draggable",
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal("Error:", err.Error())
	}
}
