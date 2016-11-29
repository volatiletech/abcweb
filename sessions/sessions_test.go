package sessions

import (
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestPutAndGet(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := httptest.NewRecorder()

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	r, err := Put(s, w, r, "hi", "hello")
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
	r, err = Put(s, w, r, "hi", "spiders")
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
	w := httptest.NewRecorder()

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	err := Del(s, w, r, "test")
	if !IsNoSessionError(err) {
		t.Error("Expected no session error")
	}

	r, err = Put(s, w, r, "hi", "hello")
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
		if v.value != "{}" {
			t.Errorf("Expected value to be empty json map, but was: %#v", v.value)
		}
	}
}

type TestSessJSON struct {
	Test string
}

func TestPutAndGetObj(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := httptest.NewRecorder()

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

	r, err = PutObj(s, w, r, val)
	if err != nil {
		t.Error(err)
	}

	err = GetObj(s, w, r, testptr)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(val, testptr) {
		t.Errorf("Expected testptr to match val, got: %#v", testptr)
	}

	// Run the same tests again to ensure it overrides instead of creates
	// a new session

	if len(m.sessions) != 1 {
		t.Errorf("Expected len 1, got %d", len(m.sessions))
	}

	r, err = PutObj(s, w, r, val)
	if err != nil {
		t.Error(err)
	}

	err = GetObj(s, w, r, testptr)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(val, testptr) {
		t.Errorf("Expected testptr to match val, got: %#v", testptr)
	}

	if len(m.sessions) != 1 {
		t.Errorf("Expected len 1, got %d", len(m.sessions))
	}
}
