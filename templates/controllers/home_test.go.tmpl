package controllers

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMainHome(t *testing.T) {
	t.Parallel()

	// Initialize the Main controller struct with a Root struct mock
	m := Main{
		Root: newRootMock("../templates"),
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	// Set the logger on the request context so calls to Log() don't panic.
	// This is only required if our controller calls Log().
	// In lots of cases the Errors middleware will handle the logging calls instead.
	// r = r.WithContext(context.WithValue(r.Context(), abcmiddleware.CtxLoggerKey, m.Log))

	// Call the controller with the httptest recorder and httptest request
	err := m.Home(w, r)
	if err != nil {
		t.Error(err)
	}

	// Ensure we get the contents of main/home.html back
	if !strings.Contains(w.Body.String(), "<span>Hello World!</span>") {
		t.Error("home template not expected value")
	}
}
