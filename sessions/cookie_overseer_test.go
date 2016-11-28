package sessions

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

var testCookieKey, _ = MakeSecretKey()

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

func TestCookieOverseerPut(t *testing.T) {
	t.Parallel()

	c := NewCookieOverseer(NewCookieOptions(), testCookieKey)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	var err error
	r, err = c.Put(w, r, "hello world")
	if err != nil {
		t.Error(err)
	}

	if val, err := c.Get(w, r); err != nil {
		t.Error(err)
	} else if val != "hello world" {
		t.Error("value was wrong:", val)
	}
}

func TestCookieOverseerDel(t *testing.T) {
	t.Parallel()

	opts := NewCookieOptions()
	c := NewCookieOverseer(opts, testCookieKey)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	if err := c.Del(w, r); err != nil {
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

	c := NewCookieOverseer(CookieOptions{}, testCookieKey)
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
