package sessions

import (
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestSetAndGet(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := newResponse(httptest.NewRecorder())

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	err := Set(s, w, r, "hi", "hello")
	if err != nil {
		t.Fatal(err)
	}

	ret, err := Get(s, w, r, "hi")
	if err != nil {
		t.Fatal(err)
	}
	if ret != "hello" {
		t.Errorf("Expected %q, got: %q", "hello", ret)
	}

	ret, err = Get(s, w, r, "lol")
	if !IsNoMapKeyError(err) {
		t.Error("Expected no map key err")
	}

	if len(m.sessions) != 1 {
		t.Errorf("Expected len 1, got %d", len(m.sessions))
	}

	// Test reassigning existing key value
	err = Set(s, w, r, "hi", "spiders")
	if err != nil {
		t.Fatal(err)
	}

	ret, err = Get(s, w, r, "hi")
	if err != nil {
		t.Fatal(err)
	}
	if ret != "spiders" {
		t.Errorf("Expected %q, got: %q", "spiders", ret)
	}

	if len(m.sessions) != 1 {
		t.Errorf("Expected len 1, got %d", len(m.sessions))
	}
}

func TestDel(t *testing.T) {
	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := newResponse(httptest.NewRecorder())

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	err := Del(s, w, r, "test")
	if !IsNoSessionError(err) {
		t.Error("Expected no session error")
	}

	err = Set(s, w, r, "hi", "hello")
	if err != nil {
		t.Fatal(err)
	}

	err = Del(s, w, r, "test")
	if err != nil {
		t.Errorf("Expected del to noop when there is no key, got %#v", err)
	}

	if len(m.sessions) != 1 {
		t.Errorf("Expected len 1, got %d", len(m.sessions))
	}

	err = Del(s, w, r, "hi")
	if err != nil {
		t.Error(err)
	}

	if len(m.sessions) != 1 {
		t.Fatalf("Expected len 1, got %d", len(m.sessions))
	}

	for _, v := range m.sessions {
		if v.value != `{"Value":{},"Flash":null}` {
			t.Errorf("Expected value to be empty json map, but was: %#v", v.value)
		}
	}
}

type TestSessJSON struct {
	Test string
}

func TestSetAndGetObj(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := newResponse(httptest.NewRecorder())

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	testptr := &TestSessJSON{}

	err := GetObj(s, w, r, testptr)
	if !IsNoSessionError(err) {
		t.Errorf("expected an error")
	}

	val := &TestSessJSON{
		Test: "hello",
	}

	err = SetObj(s, w, r, val)
	if err != nil {
		t.Error(err)
	}

	err = GetObj(s, w, r, testptr)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(val, testptr) {
		t.Errorf("Expected testptr to match val, got:\n%#v\n%#v", testptr, val)
	}

	// Run the same tests again to ensure it overrides instead of creates
	// a new session

	if len(m.sessions) != 1 {
		t.Errorf("Expected len 1, got %d", len(m.sessions))
	}

	err = SetObj(s, w, r, val)
	if err != nil {
		t.Error(err)
	}

	err = GetObj(s, w, r, testptr)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(val, testptr) {
		t.Errorf("Expected testptr to match val, got:\n%#v\n%#v", testptr, val)
	}

	if len(m.sessions) != 1 {
		t.Errorf("Expected len 1, got %d", len(m.sessions))
	}
}

func TestAddFlash(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := newResponse(httptest.NewRecorder())

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	err := AddFlash(s, w, r, "test", "flashvalue")
	if err != nil {
		t.Error(err)
	}

	var sess memorySession
	for _, v := range m.sessions {
		sess = v
	}
	if sess.value != `{"Value":{"test":"flashvalue"},"Flash":null}` {
		t.Errorf("expected session value to be %q, but got %q", `{"test":"flashvalue"}`, sess.value)
	}
	if len(w.cookies) != 1 {
		t.Error("expected cookies len 1, got:", len(w.cookies))
	}
}

func TestGetFlash(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := newResponse(httptest.NewRecorder())

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	// Ensure fails when appropriate
	_, err := GetFlash(s, w, r, "test")
	if err == nil {
		t.Error("expected err to be an error")
	}

	err = AddFlash(s, w, r, "test", "flashvalue")
	if err != nil {
		t.Error(err)
	}

	val, err := GetFlash(s, w, r, "test")
	if val != "flashvalue" {
		t.Errorf("Expected value to be %q but got %q", "flashvalue", val)
	}
	// Ensure len cookies and sessions are still present
	if len(w.cookies) != 1 {
		t.Error("expected cookies len to be 1, got:", len(w.cookies))
	}
	if len(m.sessions) != 1 {
		t.Error("expected sessions len to be 1, got:", len(m.sessions))
	}
	// Ensure key is deleted from JSON
	var sess memorySession
	for _, v := range m.sessions {
		sess = v
	}
	if sess.value != `{"Value":{},"Flash":null}` {
		t.Errorf("expected session value to be %q, but got %q", `{"Value":{},"Flash":null}`, sess.value)
	}
}
