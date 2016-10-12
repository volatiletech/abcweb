package controllers

import (
	"net/http"

	"github.com/labstack/echo"
)

// Index page
func Index(c echo.Context) error {
	return c.Render(http.StatusOK, "index", nil)
}
