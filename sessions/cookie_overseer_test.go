package sessions

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
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
	http.SetCookie(w, opts.makeCookie(ct))

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

	c.Put(w, r, "hello world")
	w.Header().Get("Set-Cookie")

	if val, err := c.Get(w, r); err != nil {
		t.Error(err)
	} else if val != "hello world" {
		t.Error("value was wrong:", val)
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
