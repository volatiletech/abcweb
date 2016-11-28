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

	val := map[string]string{
		"hi": "hello",
	}

	r, err := Put(s, w, r, val)
	if err != nil {
		t.Error(err)
	}

	ret, err := Get(s, w, r)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(val, ret) {
		t.Errorf("Expected ret to match val, got: %#v", ret)
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

	val := &TestSessJSON{
		Test: "hello",
	}

	r, err := PutObj(s, w, r, val)
	if err != nil {
		t.Error(err)
	}

	testptr := &TestSessJSON{}

	err = GetObj(s, w, r, testptr)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(val, testptr) {
		t.Errorf("Expected testptr to match val, got: %#v", testptr)
	}

}
