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
	m.mut.Lock()
	m.sessions["sessionid"] = &memorySession{
		value: "whatever",
	}
	m.mut.Unlock()
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
	m.mut.RLock()
	if len(m.sessions) != 1 {
		t.Errorf("Expected sessions len 1, got %d", len(m.sessions))
	}
	m.mut.RUnlock()

	id1 := rgxPutCookie.FindStringSubmatch(setCookie)[1]

	m.mut.RLock()
	sess, ok := m.sessions[id1]
	if !ok {
		t.Errorf("could not find session with id: %s", id1)
	}
	if sess.value != "whatever" {
		t.Errorf("expected sess value %q, got %q", "whatever", sess.value)
	}
	m.mut.RUnlock()

	// make sure it re-uses the same session cookie
	m.Put(r, w, "hello")
	setCookie = w.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Errorf("expected set cookie to be set")
	}
	if !rgxPutCookie.MatchString(setCookie) {
		t.Errorf("Expected to match regexp, got: %s", setCookie)
	}
	m.mut.RLock()
	if len(m.sessions) != 1 {
		t.Errorf("Expected sessions len 1, got %d", len(m.sessions))
	}
	m.mut.RUnlock()

	id2 := rgxPutCookie.FindStringSubmatch(setCookie)[1]

	m.mut.RLock()
	sess, ok = m.sessions[id2]
	if !ok {
		t.Errorf("could not find session with id: %s", id2)
	}
	if sess.value != "hello" {
		t.Errorf("expected sess value %q, got %q", "hello", sess.value)
	}
	m.mut.RUnlock()

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

	m.mut.RLock()
	if len(m.sessions) != 0 {
		t.Errorf("Expected sessions len 0, got %d", len(m.sessions))
	}
	m.mut.RUnlock()

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
	m.mut.RLock()
	if len(m.sessions) != 0 {
		t.Errorf("Expected sessions len 0, got %d", len(m.sessions))
	}
	m.mut.RUnlock()
}

func TestMemoryStorerCleaner(t *testing.T) {
	wait := make(chan struct{})

	sleepFunc = func(time.Duration) {
		<-wait
	}

	m, err := NewMemoryStorer(true, true, time.Minute, time.Hour, time.Hour)
	if err != nil {
		t.Error(err)
	}

	m.sessions["testid1"] = &memorySession{
		value:   "test1",
		expires: time.Now().Add(time.Hour),
	}
	m.sessions["testid2"] = &memorySession{
		value:   "test2",
		expires: time.Now().AddDate(0, 0, -1),
	}

	m.mut.RLock()
	if len(m.sessions) != 2 {
		t.Error("expected len 2")
	}
	m.mut.RUnlock()

	// stop sleep in cleaner loop
	wait <- struct{}{}
	wait <- struct{}{}

	m.mut.RLock()
	if len(m.sessions) != 1 {
		t.Errorf("expected len 1, got %d", len(m.sessions))
	}

	_, ok := m.sessions["testid2"]
	if ok {
		t.Error("expected testid2 to be deleted, but was not")
	}
	m.mut.RUnlock()
}

func TestMakeCookie(t *testing.T) {
	t.Parallel()

	m, err := NewMemoryStorer(true, true, time.Minute, time.Hour, time.Hour)
	if err != nil {
		t.Error(err)
	}
	c := m.makeCookie()

	if c.Name != SessionKey {
		t.Errorf("expected name to be session key, got: %v", c.Name)
	}
	if c.Value == "" {
		t.Errorf("expected value to be a uuid")
	}
	if c.MaxAge != int(m.clientExpiry.Seconds()) {
		t.Errorf("mismatch between %d and %d", c.MaxAge, int(m.clientExpiry.Seconds()))
	}
	if c.HttpOnly != true {
		t.Error("expected httponly true")
	}
	if c.Secure != true {
		t.Error("expected httponly true")
	}
}
