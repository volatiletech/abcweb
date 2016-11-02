package sessions

import (
	"net/http"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

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

	r, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	m, err := NewDefaultMemoryStorer(false)
	if err != nil {
		t.Fatal(err)
	}

	val, err := m.Get(r)
	if err != ErrNoSession {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}

	cookieOne := &http.Cookie{
		Name:  "test",
		Value: "test",
	}
	r.AddCookie(cookieOne)

	if ln := len(r.Cookies()); ln != 1 {
		t.Errorf("Expected cookie len 1, got %d", ln)
	}

	val, err = m.Get(r)
	if err != ErrNoSession {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}

	cookieTwo := &http.Cookie{
		Name:  SessionKey,
		Value: "test2",
	}
	r.AddCookie(cookieTwo)

	if ln := len(r.Cookies()); ln != 2 {
		t.Errorf("Expected cookie len 2, got %d", ln)
	}

	val, err = m.Get(r)
	// should be ErrNoSession because cookie does not exist in session store yet
	if err != ErrNoSession {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}

	m.sessions["test2"] = memorySession{value: "whatever"}
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

	r, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	m, err := NewMemoryStorer(true, true, time.Minute, time.Hour, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	_, err = m.Get(r)
	if err != ErrNoSession {
		t.Errorf("Expected ErrNoSession, got: %v", err)
	}

	if ln := len(r.Cookies()); ln != 0 {
		t.Errorf("Expected cookie len 0, got %d", ln)
	}

	m.Put(r, "whatever")

	if ln := len(r.Cookies()); ln != 1 {
		t.Errorf("Expected cookie len 1, got %d", ln)
	}

	spew.Dump(r.Cookies())

	val, err := m.Get(r)
	if err != nil {
		t.Error(err)
	}

	if val != "whatever" {
		t.Errorf("Expected %q, got %q", "whatever", val)
	}

	c, err := r.Cookie(SessionKey)
	if err != nil {
		t.Error(err)
	}
	id := c.Value

	m.Put(r, "hello")

	if ln := len(r.Cookies()); ln != 1 {
		t.Errorf("Expected cookie len 1, got %d", ln)
	}

	val, err = m.Get(r)
	if err != nil {
		t.Error(err)
	}

	if val != "hello" {
		t.Errorf("Expected %q, got %q", "hello", val)
	}

	c, err = r.Cookie(SessionKey)
	if err != nil {
		t.Error(err)
	}
	id2 := c.Value

	if id != id2 {
		t.Errorf("Expected to re-use cookie, but got different id: %q and %q", id, id2)
	}
}

func TestMemoryStorerDel(t *testing.T) {
	t.Parallel()
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
