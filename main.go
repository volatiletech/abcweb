package main

import (
	"fmt"
	"os"

	"github.com/nullbio/abcweb/app"
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
			// Start https and/or http listeners (app/server.go)
			return a.StartServer()
		},
	}

	// Register the command-line configuration flags (app/config.go)
	a.RegisterFlags()

	// Build app Config using env vars, cfg file and cmd line flags (app/config.go)
	if err := a.LoadConfig(); err != nil {
		fmt.Println("failed to load app config", err)
		os.Exit(1)
	}

	// Initialize the zap logger (app/logger.go)
	a.InitLogger()

	// Create a new router
	a.Router = chi.NewRouter()

	// Enable middleware for the router (app/middleware.go)
	a.InitMiddleware()

	// Configure the renderer (app/render.go)
	a.InitRenderer()

	// Initialize the routes with the renderer (app/routes.go)
	a.InitRoutes()

	if err := a.Root.Execute(); err != nil {
		a.Log.Fatal("root command execution failed", zap.Error(err))
	}
}
