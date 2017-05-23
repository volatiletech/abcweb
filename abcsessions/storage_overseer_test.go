package abcsessions

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"
)

var (
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
	w := newSessionsResponseWriter(httptest.NewRecorder())

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
	w := newSessionsResponseWriter(httptest.NewRecorder())

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	_, err := s.Get(w, r)
	if !IsNoSessionError(err) {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}

	err = s.Set(w, r, "whatever")
	if err != nil {
		t.Error(err)
	}

	if len(w.cookies) != 1 {
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

	// make sure it re-uses the same session cookie by utilizing cookies storage
	err = s.Set(w, r, "hello")
	if err != nil {
		t.Error(err)
	}

	if len(m.sessions) != 1 {
		t.Errorf("Expected sessions len 1, got %d", len(m.sessions))
	}

	if len(w.cookies) != 1 {
		t.Errorf("expected set cookie to be set")
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
	w := newSessionsResponseWriter(httptest.NewRecorder())

	opts := NewCookieOptions()
	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(opts, m)

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

	err := s.Del(w, r)
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
	w := newSessionsResponseWriter(httptest.NewRecorder())

	opts := NewCookieOptions()
	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(opts, m)

	_, err := s.SessionID(w, r)
	if !IsNoSessionError(err) {
		t.Error("Expected to get a error back")
	}

	err = s.Set(w, r, "test")
	if err != nil {
		t.Error(err)
	}

	id, err := s.SessionID(w, r)
	if err != nil {
		t.Error(err)
	}

	err = s.Regenerate(w, r)
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

	cookie := w.cookies[opts.Name]
	if cookie.Expires.String() != (time.Time{}).String() {
		t.Error("Expected cookie Expires to be zero value, but got:", cookie.Expires.String())
	}
	if cookie.MaxAge != 0 {
		t.Error("Expected 0, got:", cookie.MaxAge)
	}
	if cookie.Value == "" {
		t.Error("Expected value to be random id, got:", cookie.Value)
	}
	if cookie.Name != opts.Name {
		t.Errorf("Expected cookie Name to be %q, got %q", opts.Name, cookie.Name)
	}
}

func TestStorageOverseerSessionID(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := newSessionsResponseWriter(httptest.NewRecorder())

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	_, err := s.SessionID(w, r)
	if !IsNoSessionError(err) {
		t.Error("Expected to get a error back")
	}

	err = s.Set(w, r, "test")
	if err != nil {
		t.Error(err)
	}

	id, err := s.SessionID(w, r)
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

	opts := NewCookieOptions()
	opts.MaxAge = time.Hour * 1

	mem, err := NewDefaultMemoryStorer()
	if err != nil {
		t.Error(err)
	}

	c := NewStorageOverseer(opts, mem)
	w := newSessionsResponseWriter(httptest.NewRecorder())
	r := httptest.NewRequest("GET", "/", nil)

	err = c.ResetExpiry(w, r)
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
