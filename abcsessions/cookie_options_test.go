package abcsessions

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewCookieOptions(t *testing.T) {
	t.Parallel()

	o := NewCookieOptions()
	if o.Name != "id" {
		t.Errorf("expected name to be %q", "id")
	}
	if o.MaxAge != 0 {
		t.Error("expected max age to be 0")
	}
	if o.Secure != true {
		t.Error("expected secure to be true")
	}
	if o.HTTPOnly != true {
		t.Error("expected httponly to be true")
	}
}

func TestMakeCookie(t *testing.T) {
	t.Parallel()

	o := NewCookieOptions()
	c := o.makeCookie("test")

	if c.Name != o.Name {
		t.Errorf("Expected name %q to match %q", c.Name, o.Name)
	}
	if c.Value != "test" {
		t.Errorf("Expected value %q to match %q", c.Value, "test")
	}
	if c.MaxAge != int(o.MaxAge.Seconds()) {
		t.Errorf("Expected maxage %v to match %v", c.MaxAge, o.MaxAge.Seconds())
	}
	if c.HttpOnly != o.HTTPOnly {
		t.Errorf("expected httponly %t to match %t", c.HttpOnly, o.HTTPOnly)
	}
	if c.Secure != o.Secure {
		t.Errorf("expected secure %t to match %t", c.Secure, o.Secure)
	}
	if c.MaxAge != 0 {
		if c.Expires.Equal(time.Time{}) {
			t.Errorf("when maxage is 0 expected expires to be non-zero")
		}
	}
}

func TestGetCookieValue(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := newSessionsResponseWriter(httptest.NewRecorder())
	o := NewCookieOptions()

	_, err := o.getCookieValue(w, r)
	if err == nil {
		t.Error("Expected to get error back")
	}

	w.SetCookie(&http.Cookie{Name: "id", Value: "idvalue"})
	val, err := o.getCookieValue(w, r)
	if err != nil {
		t.Error(err)
	}
	if val != "idvalue" {
		t.Errorf("Expected %q, got %q", "idvalue", val)
	}

	// Init a new map to clear out all cookies
	w.cookies = make(map[string]*http.Cookie)

	r.AddCookie(o.makeCookie("cookievalue"))
	val, err = o.getCookieValue(w, r)
	if err != nil {
		t.Error(err)
	}
	if val != "cookievalue" {
		t.Errorf("Expected %q, got %q", "cookievalue", val)
	}

	// Ensure "idvalue" (cache storage) takes precedence over "cookievalue" (request cookie)
	w.SetCookie(&http.Cookie{Name: "id", Value: "idvalue"})
	val, err = o.getCookieValue(w, r)
	if err != nil {
		t.Error(err)
	}
	if val != "idvalue" {
		t.Errorf("Expected %q, got %q", "idvalue", val)
	}
}
