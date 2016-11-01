package sessions

import (
	"net/http"
	"testing"
)

func TestMemorySessionsNew(t *testing.T) {
	t.Parallel()
}

func TestMemorySessionsGet(t *testing.T) {
	t.Parallel()

	r, err := http.NewRequest("GET", "http://localhost", nil)
	if err != nil {
		t.Fatal(err)
	}

	m, err := NewMemorySessions(false, false, 0, 0)
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

func TestMemorySessionsPut(t *testing.T) {
	t.Parallel()
}

func TestMemorySessionsDel(t *testing.T) {
	t.Parallel()
}
