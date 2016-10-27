package app

import (
	"net/http"

	"github.com/nullbio/abcweb/controllers"
)

func (s State) InitRoutes() {
	// The state for each route handler
	root := controllers.Root{
		Log:    s.Log,
		Render: s.Render,
	}

	// Serve static assets
	s.Router.FileServer("/assets", http.Dir(s.Config.AssetsIn))

	home := controllers.Home{Root: root}
	s.Router.Get("/", home.Index)
}
