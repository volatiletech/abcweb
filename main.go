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
	var state app.State

	state.Root = &cobra.Command{
		Use:   "{{.AppName}} [flags]",
		Short: "{{.AppName}} web app server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Start https and/or http listeners (app/server.go)
			return state.StartServer()
		},
	}

	// Register the command-line configuration flags (app/config.go)
	state.RegisterFlags()

	// Build app Config using env vars, cfg file and cmd line flags (app/config.go)
	if err := state.LoadConfig(); err != nil {
		fmt.Println("failed to load app config", err)
		os.Exit(1)
	}

	// Initialize the zap logger (app/logger.go)
	state.InitLogger()

	// Create a new router
	state.Router = chi.NewRouter()

	// Configure the sessions overseer (app/sessions.go)
	state.InitSessions()

	// Configure the renderer (app/render.go)
	state.InitRenderer()

	// Enable middleware for the router (app/middleware.go)
	state.InitMiddleware()

	// Initialize the routes with the renderer (app/routes.go)
	state.InitRoutes()

	// Execute the root command Run method
	if err := state.Root.Execute(); err != nil {
		state.Log.Fatal("root command execution failed", zap.Error(err))
	}
}
