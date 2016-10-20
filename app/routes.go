package app

import (
	"net/http"

	"github.com/nullbio/abcweb/controllers"
)

func (a AppState) InitRoutes() {
	// The state for each route handler
	controller := controllers.Controller{
		Log:    a.Log,
		Render: a.Render,
	}

	// Serve static assets
	a.Router.FileServer("/assets", http.Dir(a.Config.AssetsIn))

	a.Router.Get("/", controller.Home)
}
