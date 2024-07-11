package main

import (
	"embed"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "croc-desktop",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
			&Install{},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

type Install struct {
	InstallPath string
}

func (i *Install) Install() {
	if i.InstallPath == "" {
		println("Fatal error: InstallPath is empty.")
	}
	if runtime.GOOS == "windows" {
		println("Installing to " + i.InstallPath)
		var url string
		if runtime.GOARCH == "arm64" {
			url = "https://github.com/schollz/croc/releases/download/v10.0.10/croc_v10.0.10_Windows-ARM.zip"
		} else if runtime.GOARCH == "386" {
			url = "https://github.com/schollz/croc/releases/download/v10.0.10/croc_v10.0.10_Windows-64bit.zip"
			// download and unzip
		}
		resp, err := http.Get(url)
		if err != nil {
			println("Error downloading croc:", err.Error())
		}
		defer resp.Body.Close()
		// move to install path
		out, err := os.Create(i.InstallPath + strings.Split(url, "/")[len(strings.Split(url, "/"))-1])
		if err != nil {
			println("Error creating file:", err.Error())
		}
		defer out.Close()
		_, _ = io.Copy(out, resp.Body)
	} else {
		println("Unsupported OS")
	}
}
