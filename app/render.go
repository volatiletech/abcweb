package app

import (
	"html/template"

	"github.com/unrolled/render"
)

// appHelpers is a map of the template helper functions.
// Assign the template helper funcs in here, example:
// "titleCase": strings.TitleCase
var appHelpers = template.FuncMap{}

// Initialize the renderer using the app configuration
func (s State) InitRenderer() *render.Render {
	return render.New(render.Options{
		Directory:     s.Config.Templates,
		Layout:        "layout",
		Extensions:    []string{".tmpl", ".html"},
		IsDevelopment: s.Config.RenderRecompile,
		Funcs:         []template.FuncMap{appHelpers},
	})
}
