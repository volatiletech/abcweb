package main

import (
	"html/template"

	"github.com/unrolled/render"
)

// appHelpers is a map of the template helper functions
var appHelpers map[string]interface{}

// Initialize the template helper function map
func init() {
	// Assign the template helper funcs in here, example:
	// "titleCase": strings.TitleCase
	appHelpers = map[string]interface{}{}
}

// Initialize the renderer using the app configuration
func initRenderer(c appConfig) *render.Render {
	return render.New(render.Options{
		Directory:     c.templates,
		Layout:        "layout",
		Extensions:    []string{".tmpl", ".html"},
		IsDevelopment: c.renderRecompile,
		Funcs:         []template.FuncMap{appHelpers},
	})
}
