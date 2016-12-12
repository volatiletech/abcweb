package sessions

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

var (
	rgxDelCookie = regexp.MustCompile(`id=; Expires=[^;]*; Max-Age=0; HttpOnly; Secure`)
	rgxSetCookie = regexp.MustCompile(`id=[A-Za-z0-9\-]+; HttpOnly; Secure`)
)

func TestStorageImplements(t *testing.T) {
	t.Parallel()

	// Do assign to nothing to check if implementation of StorageOverseer is complete
	var _ Overseer = &StorageOverseer{}
}

func TestStorageOverseerNew(t *testing.T) {
	t.Parallel()

	opts := CookieOptions{
		MaxAge:   2,
		Secure:   true,
		HTTPOnly: true,
		Name:     "id",
	}

	mem, err := NewDefaultMemoryStorer()
	if err != nil {
		t.Fatal(err)
	}

	s := NewStorageOverseer(opts, mem)
	if err != nil {
		t.Error(err)
	}

	if s.options.MaxAge != 2 {
		t.Error("expected client expiry to be 2")
	}

	if s.options.Secure != true {
		t.Error("expected secure to be true")
	}

	if s.options.HTTPOnly != true {
		t.Error("expected httpOnly to be true")
	}
}

func TestStorageOverseerGetSessionID(t *testing.T) {
	t.Parallel()
}

func TestStorageOverseerGet(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := httptest.NewRecorder()

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	val, err := s.Get(w, r)
	if !IsNoSessionError(err) {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}

	cookieOne := &http.Cookie{
		Name:  s.options.Name,
		Value: "sessionid",
	}
	r.AddCookie(cookieOne)

	val, err = s.Get(w, r)
	if !IsNoSessionError(err) {
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

func TestStorageOverseerSet(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := httptest.NewRecorder()

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	_, err := s.Get(w, r)
	if !IsNoSessionError(err) {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}

	r, err = s.Set(w, r, "whatever")
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
	r, err = s.Set(w, r, "hello")
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
		Name:  s.options.Name,
		Value: "sessionid",
	}
	r.AddCookie(cookieOne)
	m.sessions["sessionid"] = memorySession{
		value: "whatever",
	}

	r, err := s.Del(w, r)
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
	c := s.options.makeCookie("hello")

	if c.Name != s.options.Name {
		t.Errorf("expected name to be session key, got: %v", c.Name)
	}
	if c.Value == "" {
		t.Errorf("expected value to be a uuid")
	}
	if c.MaxAge != int(s.options.MaxAge.Seconds()) {
		t.Errorf("mismatch between %d and %d", c.MaxAge, int(s.options.MaxAge.Seconds()))
	}
	if c.HttpOnly != true {
		t.Error("expected httponly true")
	}
	if c.Secure != true {
		t.Error("expected httponly true")
	}
}

func TestStorageOverseerRegenerate(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := httptest.NewRecorder()

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	_, err := s.SessionID(r)
	if !IsNoSessionError(err) {
		t.Error("Expected to get a error back")
	}

	r, err = s.Set(w, r, "test")
	if err != nil {
		t.Error(err)
	}

	id, err := s.SessionID(r)
	if err != nil {
		t.Error(err)
	}

	r, err = s.Regenerate(w, r)
	if err != nil {
		t.Error(err)
	}

	if len(m.sessions) != 1 {
		t.Errorf("Expected sessions len 1, got %d", len(m.sessions))
	}
	for k, v := range m.sessions {
		if k == id {
			t.Errorf("Expected id %q to NOT equal %q", id, k)
		}
		if v.value != "test" {
			t.Errorf("Expected val %q to be %q", "test", v.value)
		}
	}

	setCookie := w.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Errorf("expected set cookie to be set")
	}
	if len(m.sessions) != 1 {
		t.Errorf("Expected sessions len 1, got %d", len(m.sessions))
	}
	if !rgxSetCookie.MatchString(setCookie) {
		t.Errorf("Expected to match regexp, got: %s", setCookie)
	}

}

func TestStorageOverseerSessionID(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := httptest.NewRecorder()

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	_, err := s.SessionID(r)
	if !IsNoSessionError(err) {
		t.Error("Expected to get a error back")
	}

	r, err = s.Set(w, r, "test")
	if err != nil {
		t.Error(err)
	}

	id, err := s.SessionID(r)
	if err != nil {
		t.Error(err)
	}

	if len(m.sessions) != 1 {
		t.Errorf("Expected sessions len 1, got %d", len(m.sessions))
	}
	for k, v := range m.sessions {
		if k != id {
			t.Errorf("Expected id %q to be %q", id, k)
		}
		if v.value != "test" {
			t.Errorf("Expected val %q to be %q", "test", v.value)
		}
	}
}

func TestStorageOverseerResetExpiry(t *testing.T) {
	t.Parallel()

	t.Error("not implemented")
}
