package abcsessions

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"
)

var testCookieKey = []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

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

	if bytes.Compare(c.secretKey, testCookieKey) != 0 {
		t.Error("key was not copied")
	}
}

func TestCookieOverseerGetFromCookie(t *testing.T) {
	t.Parallel()

	c := NewCookieOverseer(NewCookieOptions(), testCookieKey)
	w := newSessionsResponseWriter(httptest.NewRecorder())
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

func TestCookieOverseerNoSession(t *testing.T) {
	t.Parallel()

	opts := NewCookieOptions()
	c := NewCookieOverseer(opts, testCookieKey)
	w := newSessionsResponseWriter(httptest.NewRecorder())
	r := httptest.NewRequest("GET", "/", nil)

	_, err := c.Get(w, r)
	if !IsNoSessionError(err) {
		t.Error("wrong err:", err)
	}
}

func TestCookieOverseerSet(t *testing.T) {
	t.Parallel()

	c := NewCookieOverseer(NewCookieOptions(), testCookieKey)
	w := newSessionsResponseWriter(httptest.NewRecorder())
	r := httptest.NewRequest("GET", "/", nil)

	var err error
	err = c.Set(w, r, "hello world")
	if err != nil {
		t.Error(err)
	}

	if val, err := c.Get(w, r); err != nil {
		t.Error(err)
	} else if val != "hello world" {
		t.Error("value was wrong:", val)
	}

	if len(w.cookies) != 1 {
		t.Errorf("expected set cookie to be set")
	}

	// make sure it re-uses the same session cookie by utilizing cookies storage
	err = c.Set(w, r, "test")
	if err != nil {
		t.Error(err)
	}

	if len(w.cookies) != 1 {
		t.Errorf("expected set cookie to be set")
	}
}

func TestCookieOverseerDel(t *testing.T) {
	t.Parallel()

	opts := NewCookieOptions()
	c := NewCookieOverseer(opts, testCookieKey)
	w := newSessionsResponseWriter(httptest.NewRecorder())
	r := httptest.NewRequest("GET", "/", nil)

	err := c.Del(w, r)

	if err != nil {
		t.Error(err)
	}

	cookie := w.cookies[opts.Name]
	if !cookie.Expires.UTC().Before(time.Now().UTC().AddDate(0, 0, -1)) {
		t.Error("Expected cookie expires to be set to a year ago, but was not:", cookie.Expires.String())
	}
	if cookie.MaxAge != -1 {
		t.Error("Expected -1, got:", cookie.MaxAge)
	}
	if cookie.Value != "" {
		t.Error("Expected no value, got:", cookie.Value)
	}
	if cookie.Name != opts.Name {
		t.Errorf("Expected cookie Name to be %q, got %q", opts.Name, cookie.Name)
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
	w := newSessionsResponseWriter(httptest.NewRecorder())
	r := httptest.NewRequest("GET", "/", nil)

	err := c.ResetExpiry(w, r)
	if !IsNoSessionError(err) {
		t.Errorf("expected no session error, got %v", err)
	}

	err = c.Set(w, r, "hello")
	if err != nil {
		t.Error(err)
	}

	if len(w.cookies) != 1 {
		t.Errorf("Expected cookies len 1, got %d", len(w.cookies))
	}

	oldCookie := w.cookies[opts.Name]

	// Sleep for a ms to offset time
	time.Sleep(time.Nanosecond * 1)

	err = c.ResetExpiry(w, r)
	if err != nil {
		t.Error(err)
	}

	if len(w.cookies) != 1 {
		t.Errorf("Expected cookies len 1, got %d", len(w.cookies))
	}

	newCookie := w.cookies[opts.Name]

	if !newCookie.Expires.After(oldCookie.Expires) || newCookie.Expires == oldCookie.Expires {
		t.Errorf("Expected oldcookie and newcookie expires to be different, got:\n\n%#v\n%#v", oldCookie, newCookie)
	}
	if newCookie.Value != oldCookie.Value {
		t.Errorf("did not expect cookie values to change, got %q and %q", newCookie.Value, oldCookie.Value)
	}
	if newCookie.Name != oldCookie.Name {
		t.Errorf("did not expect cookie names to change, got %q and %q", newCookie.Name, oldCookie.Name)
	}
	if newCookie.MaxAge != oldCookie.MaxAge {
		t.Errorf("expected maxages to match, got %v and %v", newCookie.MaxAge, oldCookie.MaxAge)
	}
	if newCookie.Secure != oldCookie.Secure {
		t.Errorf("expected secures to match, got %v and %v", newCookie.Secure, oldCookie.Secure)
	}
	if newCookie.HttpOnly != oldCookie.HttpOnly {
		t.Errorf("expected httponlys to match, got %v and %v", newCookie.HttpOnly, oldCookie.HttpOnly)
	}
	if newCookie.Domain != oldCookie.Domain {
		t.Errorf("expected domains to match, got %v and %v", newCookie.Domain, oldCookie.Domain)
	}
	if newCookie.Path != oldCookie.Path {
		t.Errorf("expected paths to match, got %v and %v", newCookie.Path, oldCookie.Path)
	}
}
