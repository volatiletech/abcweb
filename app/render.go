package app

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
func (a AppState) InitRenderer() *render.Render {
	return render.New(render.Options{
		Directory:     a.Config.Templates,
		Layout:        "layout",
		Extensions:    []string{".tmpl", ".html"},
		IsDevelopment: a.Config.RenderRecompile,
		Funcs:         []template.FuncMap{appHelpers},
	})
}
