package app

import (
	"net/http"

	"github.com/nullbio/abcweb/controllers"
)

func (a AppState) InitRoutes(c controllers.Controller) {
	// Serve static assets
	a.Router.FileServer("/assets", http.Dir(a.Config.AssetsIn))

	a.Router.Get("/", c.Home)
}
