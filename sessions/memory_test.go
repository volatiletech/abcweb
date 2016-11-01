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

	r := &http.Request{}

	m, err := NewMemorySessions(false, false, 0, 0)
	if err != nil {
		t.Error(err)
	}

	val, err := m.Get(r)
	if err != ErrNoSession {
		t.Error("Expected ErrSessionNotExist, got: %s", err)
	}

	cookieOne := &http.Cookie{
		Name:  "test",
		Value: "test",
	}
	r.AddCookie(cookieOne)

	if ln := len(r.Cookies()); ln != 1 {
		t.Error("Expected cookie len 1, got %d", ln)
	}

	val, err = m.Get(r)
	if err != ErrNoSession {
		t.Error("Expected ErrSessionNotExist, got: %s", err)
	}

	cookieTwo := &http.Cookie{
		Name:  SessionKey,
		Value: "test2",
	}
	r.AddCookie(cookieTwo)

	if ln := len(r.Cookies()); ln != 2 {
		t.Error("Expected cookie len 2, got %d", ln)
	}

	val, err = m.Get(r)
	if err != nil {
		t.Fatal(err)
	}

	if val != "test2" {
		t.Error("Expected %q, got %q", "test2", val)
	}
}

func TestMemorySessionsPut(t *testing.T) {
	t.Parallel()
}

func TestMemorySessionsDel(t *testing.T) {
	t.Parallel()
}
