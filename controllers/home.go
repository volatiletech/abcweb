package controllers

import "net/http"

// Home page
func (c Controller) Home(w http.ResponseWriter, r *http.Request) {
	err := c.Render.HTML(w, http.StatusOK, "home", nil)
	if err != nil {
		c.Log.Error(err.Error())
	}
}
