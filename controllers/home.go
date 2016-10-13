package controllers

import (
	"net/http"

	"github.com/labstack/echo"
)

// Home page
func Home(c echo.Context) error {
	return c.Render(http.StatusOK, "home", nil)
}
