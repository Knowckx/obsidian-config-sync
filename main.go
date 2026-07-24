package main

import (
	"embed"
	"log"
	"os"

	"obsi-conf-sync/go_src/inner/svc"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	windowsOptions := application.WindowsOptions{}
	if os.Getenv("OBSI_WEBVIEW_DEVTOOLS") == "1" {
		windowsOptions.AdditionalBrowserArgs = []string{"--remote-debugging-port=9222"}
	}

	var mainWindow *application.WebviewWindow

	// Create a new Wails application by providing the necessary options.
	// Variables 'Name' and 'Description' are for application metadata.
	// 'Assets' configures the asset server with the 'FS' variable pointing to the frontend files.
	// 'Bind' is a list of Go struct instances. The frontend has access to the methods of these instances.
	// 'Mac' options tailor the application when running an macOS.
	app := application.New(application.Options{
		Name:        "obsi-conf-sync",
		Description: "A demo of using raw HTML & CSS",
		Services: []application.Service{
			application.NewService(&svc.VaultService{}),
			application.NewService(&svc.DevService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "app.obsi-conf-sync.desktop",
			ExitCode: 0,
			OnSecondInstanceLaunch: func(_ application.SecondInstanceData) {
				if mainWindow == nil {
					return
				}
				mainWindow.Restore()
				mainWindow.Focus()
			},
		},
		Windows: windowsOptions,
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	screen := app.Screen.GetPrimary()
	windowWidth := 1181
	windowHeight := 890
	if screen != nil {
		windowWidth = int(float64(screen.WorkArea.Width) * 0.82)
		windowHeight = int(float64(screen.WorkArea.Height) * 0.82)
	}

	// Create a new window with the necessary options.
	// 'Title' is the title of the window.
	// 'Mac' options tailor the window when running on macOS.
	// 'BackgroundColour' is the background colour of the window.
	// 'URL' is the URL that will be loaded into the webview.
	mainWindow = app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:           "Window 1",
		Width:           windowWidth,
		Height:          windowHeight,
		MinWidth:        900,
		MinHeight:       640,
		InitialPosition: application.WindowCentered,
		Screen:          screen,
		Zoom:            1.0,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(27, 38, 54),
		URL:              "/",
	})

	// Run the application. This blocks until the application has been exited.
	err := app.Run()

	// If an error occurred while running the application, log it and exit.
	if err != nil {
		log.Fatal(err)
	}
}
