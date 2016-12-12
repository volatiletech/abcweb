package sessions

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

var testCookieKey, _ = MakeSecretKey()
var rgxCookieExpires = regexp.MustCompile(`.*Expires=([^;]*);.*`)

func TestCookieImplements(t *testing.T) {
	t.Parallel()

	// Do assign to nothing to check if implementation of CookieOverseer is complete
	var _ Overseer = &CookieOverseer{}
}

func TestCookieOverseerNew(t *testing.T) {
	t.Parallel()

	c := NewCookieOverseer(NewCookieOptions(), testCookieKey)
	if c == nil {
		t.Error("c should not be nil")
	}

	if c.gcmBlockMode == nil {
		t.Error("block mode should be instantiated")
	}

	if bytes.Compare(c.secretKey[:], testCookieKey[:]) != 0 {
		t.Error("key was not copied")
	}
}

func TestCookieOverseerGetFromCookie(t *testing.T) {
	t.Parallel()

	c := NewCookieOverseer(NewCookieOptions(), testCookieKey)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ct, err := c.encode("hello world")
	if err != nil {
		t.Error(err)
	}
	r.AddCookie(&http.Cookie{
		Name:  c.options.Name,
		Value: ct,
	})

	value, err := c.Get(w, r)
	if err != nil {
		t.Error(err)
	}

	if value != "hello world" {
		t.Error("value was wrong:", value)
	}
}

func TestCookieOverseerGetFromWritten(t *testing.T) {
	t.Parallel()

	opts := NewCookieOptions()
	c := NewCookieOverseer(opts, testCookieKey)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	ct, err := c.encode("hello world")
	if err != nil {
		t.Error(err)
	}

	r = r.WithContext(context.WithValue(r.Context(), c.options.Name, ct))

	value, err := c.Get(w, r)
	if err != nil {
		t.Error(err)
	}

	if value != "hello world" {
		t.Error("value was wrong:", value)
	}
}

func TestCookieOverseerNoSession(t *testing.T) {
	t.Parallel()

	opts := NewCookieOptions()
	c := NewCookieOverseer(opts, testCookieKey)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	_, err := c.Get(w, r)
	if !IsNoSessionError(err) {
		t.Error("wrong err:", err)
	}
}

func TestCookieOverseerSet(t *testing.T) {
	t.Parallel()

	c := NewCookieOverseer(NewCookieOptions(), testCookieKey)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	var err error
	r, err = c.Set(w, r, "hello world")
	if err != nil {
		t.Error(err)
	}

	if val, err := c.Get(w, r); err != nil {
		t.Error(err)
	} else if val != "hello world" {
		t.Error("value was wrong:", val)
	}

	setCookie := w.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Errorf("expected set cookie to be set")
	}
}

func TestCookieOverseerDel(t *testing.T) {
	t.Parallel()

	opts := NewCookieOptions()
	c := NewCookieOverseer(opts, testCookieKey)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	r, err := c.Del(w, r)

	if err != nil {
		t.Error(err)
	}

	header := w.Header().Get("Set-Cookie")

	year := strconv.Itoa(time.Now().UTC().AddDate(-1, 0, 0).Year())
	foundYear := false
	foundMaxAge := false
	for _, s := range strings.Split(header, "; ") {
		if strings.Contains(s, "Expires=") && strings.Contains(s, year) {
			foundYear = true
		}
		if s == "Max-Age=0" {
			foundMaxAge = true
		}
	}

	if !foundYear {
		t.Error("could not find year", header)
	}
	if !foundMaxAge {
		t.Error("could not find maxage", header)
	}
}

func TestCookieOverseerCrypto(t *testing.T) {
	t.Parallel()

	opts := CookieOptions{
		Name: "id",
	}

	c := NewCookieOverseer(opts, testCookieKey)
	if c == nil {
		t.Error("c should not be nil")
	}

	ct, err := c.encode("hello world")
	if err != nil {
		t.Error(err)
	}
	pt, err := c.decode(ct)
	if err != nil {
		t.Error(err)
	}

	if pt != "hello world" {
		t.Error("plaintext was wrong:", pt)
	}
}

func TestCookieOverseerResetExpiry(t *testing.T) {
	t.Parallel()

	opts := NewCookieOptions()
	opts.MaxAge = time.Hour * 1

	c := NewCookieOverseer(opts, testCookieKey)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	err := c.ResetExpiry(w, r)
	if !IsNoSessionError(err) {
		t.Errorf("expected no session error, got %v", err)
	}

	r, err = c.Set(w, r, "hello")
	if err != nil {
		t.Error(err)
	}

	if w.Header().Get("Set-Cookie") == "" {
		t.Errorf("expected set cookie to be set")
	}

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Errorf("Expected cookies len 1, got %d", len(cookies))
	}

	oldCookie := cookies[0]

	// Sleep for a ms to offset time
	time.Sleep(time.Second * 1)

	err = c.ResetExpiry(w, r)
	if err != nil {
		t.Error(err)
	}

	if w.Header().Get("Set-Cookie") == "" {
		t.Errorf("expected set cookie to be set")
	}

	cookies = w.Result().Cookies()
	if len(cookies) != 1 {
		t.Errorf("Expected cookies len 1, got %d", len(cookies))
	}

	newCookie := cookies[0]

	if reflect.DeepEqual(newCookie, oldCookie) {
		t.Errorf("Expected oldcookie and newcookie to be different, got:\n\n%#v\n%#v", oldCookie, newCookie)
	}
}
