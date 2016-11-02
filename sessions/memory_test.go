package sessions

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"
)

var rgxPutCookie = regexp.MustCompile(`_SESSION_ID=([a-z0-9\-]+); Expires=[^;]*; Max-Age=60; HttpOnly; Secure`)
var rgxDelCookie = regexp.MustCompile(`_SESSION_ID=; Expires=[^;]*; Max-Age=0; HttpOnly; Secure`)

func TestMemoryStorerNew(t *testing.T) {
	t.Parallel()

	m, err := NewMemoryStorer(true, true, 1, 1, time.Hour)
	if err != nil {
		t.Error(err)
	}

	if m.clientExpiry != 1 {
		t.Error("expected client expiry to be 1")
	}

	if m.serverExpiry != 1 {
		t.Error("expected server expiry to be 1")
	}

	if m.secure != true {
		t.Error("expected secure to be true")
	}

	if m.httpOnly != true {
		t.Error("expected httpOnly to be true")
	}
}

func TestMemoryStorerNewDefault(t *testing.T) {
	t.Parallel()

	m, err := NewDefaultMemoryStorer(true)
	if err != nil {
		t.Error(err)
	}

	if m.clientExpiry != 0 {
		t.Error("expected client expiry to be zero")
	}

	if m.serverExpiry != time.Hour*24*7 {
		t.Error("expected server expiry to be a week")
	}

	if m.secure != true {
		t.Error("expected secure to be true")
	}

	if m.httpOnly != true {
		t.Error("expected httpOnly to be true")
	}
}

func TestMemoryStorerGet(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)

	m, err := NewDefaultMemoryStorer(false)
	if err != nil {
		t.Fatal(err)
	}

	val, err := m.Get(r)
	if err != ErrNoSession {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}

	cookieOne := &http.Cookie{
		Name:  SessionKey,
		Value: "sessionid",
	}
	r.AddCookie(cookieOne)

	val, err = m.Get(r)
	if err != ErrNoSession {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}
	m.sessions["sessionid"] = &memorySession{
		value: "whatever",
	}
	val, err = m.Get(r)
	if err != nil {
		t.Fatal(err)
	}

	if val != "whatever" {
		t.Errorf("Expected %q, got %q", "whatever", val)
	}
}

func TestMemoryStorerPut(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := httptest.NewRecorder()

	m, err := NewMemoryStorer(true, true, time.Minute, time.Hour, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	_, err = m.Get(r)
	if err != ErrNoSession {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}

	m.Put(r, w, "whatever")

	setCookie := w.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Errorf("expected set cookie to be set")
	}
	if !rgxPutCookie.MatchString(setCookie) {
		t.Errorf("Expected to match regexp, got: %s", setCookie)
	}
	if len(m.sessions) != 1 {
		t.Errorf("Expected sessions len 1, got %d", len(m.sessions))
	}

	id1 := rgxPutCookie.FindStringSubmatch(setCookie)[1]

	sess, ok := m.sessions[id1]
	if !ok {
		t.Errorf("could not find session with id: %s", id1)
	}
	if sess.value != "whatever" {
		t.Errorf("expected sess value %q, got %q", "whatever", sess.value)
	}

	// make sure it re-uses the same session cookie
	m.Put(r, w, "hello")
	setCookie = w.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Errorf("expected set cookie to be set")
	}
	if !rgxPutCookie.MatchString(setCookie) {
		t.Errorf("Expected to match regexp, got: %s", setCookie)
	}
	if len(m.sessions) != 1 {
		t.Errorf("Expected sessions len 1, got %d", len(m.sessions))
	}

	id2 := rgxPutCookie.FindStringSubmatch(setCookie)[1]

	sess, ok = m.sessions[id2]
	if !ok {
		t.Errorf("could not find session with id: %s", id2)
	}
	if sess.value != "hello" {
		t.Errorf("expected sess value %q, got %q", "hello", sess.value)
	}

	if id1 != id2 {
		t.Error("expected to use same session variable")
	}
}

func TestMemoryStorerDel(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := httptest.NewRecorder()

	m, err := NewMemoryStorer(true, true, time.Minute, time.Hour, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	if ln := len(r.Cookies()); ln != 0 {
		t.Errorf("Expected cookie len 0, got %d", ln)
	}

	if len(m.sessions) != 0 {
		t.Errorf("Expected sessions len 0, got %d", len(m.sessions))
	}

	cookieOne := &http.Cookie{
		Name:  SessionKey,
		Value: "sessionid",
	}
	r.AddCookie(cookieOne)
	m.sessions["sessionid"] = &memorySession{
		value: "whatever",
	}

	m.Del(r, w)

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
	if len(m.sessions) != 0 {
		t.Errorf("Expected sessions len 0, got %d", len(m.sessions))
	}
}

func TestMemoryStorerCleaner(t *testing.T) {
	t.Parallel()

	wait := make(chan struct{})
	sleepFunc = func(x time.Duration) {
		<-wait
	}

	// stop sleep in cleaner loop
	wait <- struct{}{}
}

func TestMakeCookie(t *testing.T) {
	t.Parallel()

}
