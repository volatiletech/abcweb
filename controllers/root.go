package controllers

import (
	"github.com/uber-go/zap"
	"github.com/unrolled/render"
)

// Root struct exposes useful variables to every controller route handler.
// If you wanted to pass additional objects to each controller route handler
// you can add them here, and then assign them in main.go where the
// instance of controller is created.
type Root struct {
	Log    zap.Logger
	Render *render.Render
}
