package abcsessions

import (
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestSetAndGet(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := newSessionsResponseWriter(httptest.NewRecorder())

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
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := newSessionsResponseWriter(httptest.NewRecorder())

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
	w := newSessionsResponseWriter(httptest.NewRecorder())

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
	w := newSessionsResponseWriter(httptest.NewRecorder())

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
	if sess.value != `{"Value":null,"Flash":{"test":"flashvalue"}}` {
		t.Errorf("expected session value to be %q, but got %q", `{"test":"flashvalue"}`, sess.value)
	}
	if len(w.cookies) != 1 {
		t.Error("expected cookies len 1, got:", len(w.cookies))
	}
}

func TestGetFlash(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := newSessionsResponseWriter(httptest.NewRecorder())

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
	if sess.value != `{"Value":null,"Flash":{}}` {
		t.Errorf("expected session value to be %q, but got %q", `{"Value":{},"Flash":null}`, sess.value)
	}

	val, err = GetFlash(s, w, r, "test")
	if val != "" {
		t.Errorf("Expected value to be nothing but got %q", val)
	}
	if err == nil {
		t.Error("Expected an error, but did not get one")
	}
}

type TestFlashObj struct {
	Name  string
	Valid bool
}

func TestFlashAddAndGetObj(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := newSessionsResponseWriter(httptest.NewRecorder())

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	var testObj TestFlashObj
	err := GetFlashObj(s, w, r, "test", &testObj)
	if err == nil {
		t.Error("expected an error, but got none")
	}

	myObj := TestFlashObj{Name: "test", Valid: true}
	err = AddFlashObj(s, w, r, "test", myObj)
	if err != nil {
		t.Error(err)
	}

	err = GetFlashObj(s, w, r, "test", &testObj)
	if err != nil {
		t.Error(err)
	}

	if testObj.Name != "test" {
		t.Errorf("expected obj Name to be %q, got %q", "test", testObj.Name)
	}
	if testObj.Valid != true {
		t.Error("expected obj Name to be true, got false")
	}

	var testObjTwo TestFlashObj
	err = GetFlashObj(s, w, r, "test", &testObjTwo)
	if err == nil {
		t.Error("expected error, but got none")
	}
}

func TestFlashCombined(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := newSessionsResponseWriter(httptest.NewRecorder())

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	var testObj TestFlashObj
	err := Set(s, w, r, "thing", "stuff")
	if err != nil {
		t.Error(err)
	}

	myObj := TestFlashObj{Name: "test", Valid: true}
	err = AddFlashObj(s, w, r, "test", myObj)
	if err != nil {
		t.Error(err)
	}

	result, err := Get(s, w, r, "thing")
	if result != "stuff" {
		t.Errorf("Expected %q, got %q", "stuff", result)
	}

	err = GetFlashObj(s, w, r, "test", &testObj)
	if err != nil {
		t.Error(err)
	}

	if testObj.Name != "test" {
		t.Errorf("expected obj Name to be %q, got %q", "test", testObj.Name)
	}
	if testObj.Valid != true {
		t.Error("expected obj Name to be true, got false")
	}

}

func TestFlashCombinedTwo(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := newSessionsResponseWriter(httptest.NewRecorder())

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	var testObj TestFlashObj

	myObj := TestFlashObj{Name: "test", Valid: true}
	err := AddFlashObj(s, w, r, "test", myObj)
	if err != nil {
		t.Error(err)
	}

	err = Set(s, w, r, "thing", "stuff")
	if err != nil {
		t.Error(err)
	}

	err = GetFlashObj(s, w, r, "test", &testObj)
	if err != nil {
		t.Error(err)
	}

	if testObj.Name != "test" {
		t.Errorf("expected obj Name to be %q, got %q", "test", testObj.Name)
	}
	if testObj.Valid != true {
		t.Error("expected obj Name to be true, got false")
	}

	result, err := Get(s, w, r, "thing")
	if result != "stuff" {
		t.Errorf("Expected %q, got %q", "stuff", result)
	}
}

func TestAddFlashAndSetCombined(t *testing.T) {
	t.Parallel()

	r := httptest.NewRequest("GET", "http://localhost", nil)
	w := newSessionsResponseWriter(httptest.NewRecorder())

	m, _ := NewDefaultMemoryStorer()
	s := NewStorageOverseer(NewCookieOptions(), m)

	err := AddFlash(s, w, r, "flashone", "fmone")
	if err != nil {
		t.Error(err)
	}

	err = AddFlash(s, w, r, "flashtwo", "fmtwo")
	if err != nil {
		t.Error(err)
	}

	err = Set(s, w, r, "testone", "stuffone")
	if err != nil {
		t.Error(err)
	}

	err = Set(s, w, r, "testtwo", "stufftwo")
	if err != nil {
		t.Error(err)
	}

	res, err := s.Get(w, r)
	if err != nil {
		t.Error(err)
	}

	if res != `{"Value":{"testone":"stuffone","testtwo":"stufftwo"},"Flash":{"flashone":"fmone","flashtwo":"fmtwo"}}` {
		t.Errorf("json serialized value is not as expected in session: %s", res)
	}

	res, err = GetFlash(s, w, r, "flashone")
	if err != nil {
		t.Error(err)
	}
	if res != "fmone" {
		t.Errorf("Expected %q, got %q", "fmone", res)
	}

	res, err = Get(s, w, r, "testone")
	if err != nil {
		t.Error(err)
	}
	if res != "stuffone" {
		t.Errorf("Expected %q, got %q", "stuffone", res)
	}

	res, err = GetFlash(s, w, r, "flashtwo")
	if err != nil {
		t.Error(err)
	}
	if res != "fmtwo" {
		t.Errorf("Expected %q, got %q", "fmtwo", res)
	}

	res, err = Get(s, w, r, "testtwo")
	if err != nil {
		t.Error(err)
	}
	if res != "stufftwo" {
		t.Errorf("Expected %q, got %q", "stufftwo", res)
	}

	res, err = s.Get(w, r)
	if err != nil {
		t.Error(err)
	}

	if res != `{"Value":{"testone":"stuffone","testtwo":"stufftwo"},"Flash":{}}` {
		t.Errorf("json serialized value is not as expected in session: %s", res)
	}
}

func TestValidKey(t *testing.T) {
	t.Parallel()

	// Example:
	keys := map[string]bool{
		"a668b3bb-0cf1-4627-8cd4-7f62d09ebad6": true,
		"a668b3bf-0cf9-a629-fcd0-7aaaaaaaaaaa": true,
		// too short
		"668b3bf-0cf9-a629-fcd0-7aaaaaaaaaaa": false,
		// too long
		"668b3bf-0cf9-a629-fcd0-7aaaaaaaaaaaaa": false,
		// invalid chars and positions
		"/668b3bf-0cf9-a629-fcd0-7aaaaaaaaaaa": false,
		"a668b3bf-0cf9-a6:9-fcd0-7aaaaaaaaaaa": false,
		"a668b3bf-0cf9-a6z9-fcd0-7a`aaaaaaaaa": false,
		"a668b3bf-0cf9-a6z9-fcd0-7aaaaaaaaaag": false,
		"a668b3bfa0cf9aa6z9afcd0a7aaaaaaaaaaa": false,
		"a668b3b-f0cf9-a6z9-fcd0-7aaaaaaaaaaa": false,
		"a668b3bf-0cf-9a6z9-fcd0-7aaaaaaaaaaa": false,
		"a668b3bf-0cf9-a6z-9fcd0-7aaaaaaaaaaa": false,
		"a668b3bf-0cf9-a6z9-fcd-07aaaaaaaaaaa": false,
		"a668b3b-f0cf-fa6z-ffcf-07aaaaaaaaaaa": false,
	}

	for key, valid := range keys {
		v := validKey(key)

		if valid != v {
			t.Errorf("key %s, expected validity to be: %t, got %t", key, valid, v)
		}
	}
}
