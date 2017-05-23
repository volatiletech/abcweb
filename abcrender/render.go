package abcrender

import (
	"io"

	"github.com/unrolled/render"
)

// Renderer implements template rendering methods.
// If you'd like to create a renderer that uses a different rendering engine
// opposed to standard text/templates or html/templates you can do so by
// implementing this interface.
type Renderer interface {
	Data(w io.Writer, status int, v []byte) error
	JSON(w io.Writer, status int, v interface{}) error
	Text(w io.Writer, status int, v string) error
	// HTML renders a HTML template. Example:
	// Assumes you have a template in ./templates called "home.tmpl"
	// $ mkdir -p templates && echo "<h1>Hello {{.}}</h1>" > templates/home.tmpl
	// HTML(w, http.StatusOK, "home", "World")
	HTML(w io.Writer, status int, name string, binding interface{}) error
	// HTMLWithLayout renders a HTML template using a different layout to the
	// one specified in your renderer's configuration. Example:
	// Example: HTMLWithLayout(w, http.StatusOK, "home", "World", "layout")
	HTMLWithLayout(w io.Writer, status int, name string, binding interface{}, layout string) error
}

// Render implements the HTML and HTMLWithLayout functions on the Renderer
// interface and imbeds the unrolled Render type to satisfy the rest of the interface.
// The custom HTML/HTMLWithLayout implementation is required due to the Render
// HTML function having a package-specific type for the layout string (Render.HTMLOptions)
type Render struct {
	*render.Render
}

// HTML renders a HTML template by calling unrolled Render package's HTML function
func (r *Render) HTML(w io.Writer, status int, name string, binding interface{}) error {
	return r.Render.HTML(w, status, name, binding)
}

// HTMLWithLayout renders a HTML template using a specified layout file by calling
// unrolled Render package's HTML function with a HTMLOptions argument
func (r *Render) HTMLWithLayout(w io.Writer, status int, name string, binding interface{}, layout string) error {
	return r.Render.HTML(w, status, name, binding, render.HTMLOptions{Layout: layout})
}
