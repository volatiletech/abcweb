package main

import (
	"io"

	"github.com/labstack/echo"
	"github.com/unrolled/render"
)

// RenderWrapper is needed to wrap the renderer because
// we need a different signature for echo.
type RenderWrapper struct {
	rnd *render.Render
}

// Render the HTML template
func (r *RenderWrapper) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// The zero status code is overwritten by echo.
	return r.rnd.HTML(w, 0, name, data)
}

// NewRender returns a new Render object inside a RenderWrapper
func NewRender() *RenderWrapper {
	rndOpts := render.Options{
		Directory:  *templates,
		Extensions: []string{".html"},
	}

	// If the ports are not 80 and 443, assume we're in development mode.
	// This allows us to load template changes without having to restart the app.
	if *port != 80 || *tlsport != 443 {
		rndOpts.IsDevelopment = true
	}

	return &RenderWrapper{render.New(rndOpts)}
}
