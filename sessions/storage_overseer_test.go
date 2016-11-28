package sessions

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

var rgxDelCookie = regexp.MustCompile(`_SESSION_ID=; Expires=[^;]*; Max-Age=0; HttpOnly; Secure`)

func TestStorageOverseerNew(t *testing.T) {
	t.Parallel()

	opts := CookieOptions{
		ClientExpiry: 2,
		Secure:       true,
		HTTPOnly:     true,
	}

	mem, err := NewDefaultMemoryStorer()
	if err != nil {
		t.Fatal(err)
	}

	s := NewStorageOverseer(opts, mem)
	if err != nil {
		t.Error(err)
	}

	if s.options.ClientExpiry != 2 {
		t.Error("expected client expiry to be 2")
	}

	if s.options.Secure != true {
		t.Error("expected secure to be true")
	}

	if s.options.HTTPOnly != true {
		t.Error("expected httpOnly to be true")
	}
}

func TestStorageOverseerGet(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := httptest.NewRecorder()

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	val, err := s.Get(w, r)
	if err != ErrNoSession {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}

	cookieOne := &http.Cookie{
		Name:  Key,
		Value: "sessionid",
	}
	r.AddCookie(cookieOne)

	val, err = s.Get(w, r)
	if err != ErrNoSession {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}
	m.sessions["sessionid"] = memorySession{
		value: "whatever",
	}
	val, err = s.Get(w, r)
	if err != nil {
		t.Fatal(err)
	}

	if val != "whatever" {
		t.Errorf("Expected %q, got %q", "whatever", val)
	}
}

func TestStorageOverseerPut(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := httptest.NewRecorder()

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	_, err := s.Get(w, r)
	if err != ErrNoSession {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}

	r, err = s.Put(w, r, "whatever")
	if err != nil {
		t.Error(err)
	}

	setCookie := w.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Errorf("expected set cookie to be set")
	}
	if len(m.sessions) != 1 {
		t.Errorf("Expected sessions len 1, got %d", len(m.sessions))
	}

	id1, err := s.Get(w, r)
	if err != nil {
		t.Error(err)
	}
	if id1 != "whatever" {
		t.Errorf("Expected %q, got %q", "whatever", id1)
	}

	// make sure it re-uses the same session cookie by utilizing context storage
	r, err = s.Put(w, r, "hello")
	if err != nil {
		t.Error(err)
	}

	if len(m.sessions) != 1 {
		t.Errorf("Expected sessions len 1, got %d", len(m.sessions))
	}

	id2, err := s.Get(w, r)
	if err != nil {
		t.Error(err)
	}
	if id2 != "hello" {
		t.Errorf("Expected %q, got %q", "hello", id2)
	}
}

func TestStorageOverseerDel(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := httptest.NewRecorder()

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	if ln := len(r.Cookies()); ln != 0 {
		t.Errorf("Expected cookie len 0, got %d", ln)
	}

	m.mut.RLock()
	if len(m.sessions) != 0 {
		t.Errorf("Expected sessions len 0, got %d", len(m.sessions))
	}
	m.mut.RUnlock()

	cookieOne := &http.Cookie{
		Name:  Key,
		Value: "sessionid",
	}
	r.AddCookie(cookieOne)
	m.sessions["sessionid"] = memorySession{
		value: "whatever",
	}

	err := s.Del(w, r)
	if err != nil {
		t.Error(err)
	}

	setCookie := w.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Errorf("expected del cookie to be set")
	}
	if !rgxDelCookie.MatchString(setCookie) {
		t.Errorf("Expected to match regexp, got: %s", setCookie)
	}
	m.mut.RLock()
	if len(m.sessions) != 0 {
		t.Errorf("Expected sessions len 0, got %d", len(m.sessions))
	}
	m.mut.RUnlock()
}

func TestStorageOverseerMakeCookie(t *testing.T) {
	t.Parallel()

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)
	c := s.makeCookie("hello")

	if c.Name != Key {
		t.Errorf("expected name to be session key, got: %v", c.Name)
	}
	if c.Value == "" {
		t.Errorf("expected value to be a uuid")
	}
	if c.MaxAge != int(s.options.ClientExpiry.Seconds()) {
		t.Errorf("mismatch between %d and %d", c.MaxAge, int(s.options.ClientExpiry.Seconds()))
	}
	if c.HttpOnly != true {
		t.Error("expected httponly true")
	}
	if c.Secure != true {
		t.Error("expected httponly true")
	}
}

func TestStorageOverseerGetCookieID(t *testing.T) {
	t.Parallel()

}
