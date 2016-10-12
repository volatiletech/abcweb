package main

import (
	"github.com/labstack/echo"
	"github.com/nullbio/go-web-one/controllers"
)

func initRoutes(e *echo.Echo) {
	e.Static("/assets", *assets)

	e.GET("/", controllers.Index)
}
