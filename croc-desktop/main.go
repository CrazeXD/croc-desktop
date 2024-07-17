package main

import (
	"archive/zip"
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()
	croc := NewCroc()
	install := NewInstall()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "croc-desktop",
		Width:  450,
		Height: 650,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			croc.Startup(ctx)
			install.Startup(ctx)
		},
		Bind: []interface{}{
			app,
			install,
			croc,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

type Install struct {
	InstallPath string
	ctx         context.Context
}

func NewInstall() *Install {
	return &Install{}
}

func (i *Install) Startup(ctx context.Context) {
	i.ctx = ctx
}

func (i *Install) Install() {
	i.InstallPath = i.SelectFolder()
	if i.InstallPath == "" {
		println("Fatal error: InstallPath is empty.")
		return
	}

	if runtime.GOOS == "windows" {
		println("Installing to " + i.InstallPath)
		var url string
		if runtime.GOARCH == "arm64" {
			url = "https://github.com/schollz/croc/releases/download/v10.0.10/croc_v10.0.10_Windows-ARM.zip"
		} else {
			url = "https://github.com/schollz/croc/releases/download/v10.0.10/croc_v10.0.10_Windows-64bit.zip"
		}
		println(runtime.GOARCH, url)
		// Download the zip file
		resp, err := http.Get(url)
		if err != nil {
			println("Error downloading croc:", err.Error())
			return
		}
		defer resp.Body.Close()

		// Create the destination file
		zipPath := filepath.Join(i.InstallPath, filepath.Base(url))
		out, err := os.Create(zipPath)
		if err != nil {
			println("Error creating file:", err.Error())
			return
		}
		defer out.Close()

		// Copy the downloaded content to the destination file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			println("Error saving downloaded file:", err.Error())
			return
		}

		// Unzip the downloaded file
		err = unzip(zipPath, i.InstallPath)
		if err != nil {
			println("Error unzipping file:", err.Error())
			return
		}

		// Remove the zip file after extraction
		err = os.Remove(zipPath)
		if err != nil {
			println("Error removing zip file:", err.Error())
		}

		// Add the installation path to the system PATH
		err = addToPath(i.InstallPath)
		if err != nil {
			println("Error adding to PATH:", err.Error())
		}

		println("Installation completed successfully.")
	} else {
		println("Unsupported OS")
	}
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func addToPath(path string) error {
	cmd := exec.Command("powershell", "-Command", "[Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', [EnvironmentVariableTarget]::User) + ';' + '"+path+"', [EnvironmentVariableTarget]::User)")
	return cmd.Run()
}

func (i *Install) CheckInstall() bool {
	cmd := exec.Command("croc", "--version")
	err := cmd.Run()
	return err == nil
}

// Create a folder dialog and return the selected folder
func (i *Install) SelectFolder() string {
	if i.ctx == nil {
		println("Error: context is not set")
		return ""
	}
	dir, err := wailsRuntime.OpenDirectoryDialog(i.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select Installation Directory",
	})
	if err != nil {
		println("Error:", err.Error())
		return ""
	}
	return dir
}

// App struct
type App struct {
	ctx     context.Context
	install *Install
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		install: &Install{},
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.install.Startup(ctx)
}

// Quit function
func (a *App) Quit() {
	wailsRuntime.Quit(a.ctx)
}

type Croc struct {
	ctx context.Context
}

func NewCroc() *Croc {
	return &Croc{}
}

func (c *Croc) Startup(ctx context.Context) {
	c.ctx = ctx
}

func (c *Croc) SendFile(file string) {
	if c.ctx == nil {
		println("Error: Croc context is not set in SendFile")
		return
	}

	wailsRuntime.LogInfo(c.ctx, fmt.Sprintf("Attempting to send file: %s", file))

	cmd := exec.Command("croc", "send", file)

	// Capture stdout and stderr separately
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	wailsRuntime.LogInfo(c.ctx, fmt.Sprintf("Command executed. Stdout: %s", stdout.String()))
	wailsRuntime.LogInfo(c.ctx, fmt.Sprintf("Command executed. Stderr: %s", stderr.String()))

	if err != nil {
		errMsg := fmt.Sprintf("Error running croc command: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
		wailsRuntime.LogError(c.ctx, errMsg)
		fmt.Println(errMsg)
	} else {
		successMsg := fmt.Sprintf("File send command completed. Stdout: %s", stdout.String())
		wailsRuntime.LogInfo(c.ctx, successMsg)
		fmt.Println(successMsg)
	}

	// Check if the command is still running
	if cmd.ProcessState == nil {
		wailsRuntime.LogInfo(c.ctx, "Command is still running or did not start")
	} else {
		wailsRuntime.LogInfo(c.ctx, fmt.Sprintf("Command exited with status: %v", cmd.ProcessState.ExitCode()))
	}
}

func (c *Croc) ReceiveCode(code string) {
	cmd := exec.Command("croc", code)
	out, err := cmd.CombinedOutput()
	if err != nil {
		wailsRuntime.LogError(c.ctx, fmt.Sprintf("could not run command: %v", err))
	}
	wailsRuntime.LogInfo(c.ctx, string(out))
}

func (c *Croc) OpenFile() {
	if c.ctx == nil {
		println("Error: Croc context is not set")
		return
	}
	file, err := wailsRuntime.OpenFileDialog(c.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Select File",
	})
	if err != nil {
		wailsRuntime.LogError(c.ctx, fmt.Sprintf("Error opening file dialog: %v", err))
		return
	}
	c.SendFile(file)
}
