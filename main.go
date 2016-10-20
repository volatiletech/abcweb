package main

import (
	"fmt"
	"os"

	"github.com/nullbio/abcweb/app"
	"github.com/nullbio/abcweb/controllers"
	"github.com/pressly/chi"
	"github.com/spf13/cobra"
	"github.com/uber-go/zap"
)

func main() {
	var a app.AppState

	a.Root = &cobra.Command{
		Use:   "{{.AppName}} [flags]",
		Short: "{{.AppName}} web app server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.StartServer()
		},
	}

	// Register the command-line configuration flags (config.go)
	a.RegisterFlags()

	// Load the app configuration (config.go)
	if err := a.LoadConfig(); err != nil {
		fmt.Println("failed to load app config", err)
		os.Exit(1)
	}

	// Initialize the zap logger
	a.InitLogger()

	// Create a new router
	a.Router = chi.NewRouter()

	// Enable middleware for the router
	a.InitMiddleware()

	// Configure the renderer (render.go)
	a.InitRenderer()

	// Hook up the controller object (controllers/controller.go)
	controller := controllers.Controller{
		Log:    a.Log,
		Render: a.Render,
	}

	// Initialize the routes with the renderer (routes.go)
	a.InitRoutes(controller)

	if err := a.Root.Execute(); err != nil {
		a.Log.Fatal("root command execution failed", zap.Error(err))
	}
}
