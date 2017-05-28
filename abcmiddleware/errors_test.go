package abcmiddleware

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"go.uber.org/zap"

	"github.com/volatiletech/abcweb/abcrender"
)

func TestRemoveAndAdd(t *testing.T) {
	t.Parallel()

	ea := errors.New("error1")
	eb := errors.New("error2")

	eaExpected := ErrorContainer{
		Err:      ea,
		Template: "errors/404",
		Code:     404,
		Handler:  nil,
	}

	ebExpected := ErrorContainer{
		Err:      eb,
		Template: "errors/404",
		Code:     404,
		Handler:  nil,
	}

	m := NewErrorManager(&abcrender.Render{})

	m.Add(NewError(ea, 404, "errors/404", nil))
	m.Add(NewError(eb, 404, "errors/404", nil))

	if len(m.errors) != 2 {
		t.Errorf("expected len 2, got %d", len(m.errors))
	}
	if !reflect.DeepEqual(m.errors[0], eaExpected) {
		t.Errorf("expected:\n%#v\ngot:\n%#v", eaExpected, m.errors[0])
	}
	if !reflect.DeepEqual(m.errors[1], ebExpected) {
		t.Errorf("expected:\n%#v\ngot:\n%#v", ebExpected, m.errors[1])
	}

	m.Remove(m.errors[0])

	if len(m.errors) != 1 {
		t.Errorf("expected len 1, got %d", len(m.errors))
	}

	if !reflect.DeepEqual(m.errors[0], ebExpected) {
		t.Errorf("expected:\n%#v\ngot:\n%#v", ebExpected, m.errors[1])
	}

	m.Remove(m.errors[0])
	if len(m.errors) != 0 {
		t.Errorf("expected len 0, got %d", len(m.errors))
	}
}

func TestCustomErrorHandler(t *testing.T) {
	t.Parallel()

	var globalTest bool
	myHandler := func(w http.ResponseWriter, r *http.Request, e ErrorContainer, render abcrender.Renderer) error {
		globalTest = true
		return nil
	}

	// test handler route
	ea := errors.New("error1")
	m := NewErrorManager(&abcrender.Render{})
	m.Add(NewError(ea, 404, "errors/404", myHandler))
	fn := m.Errors(func(w http.ResponseWriter, r *http.Request) error {
		return ea
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("get", "/", nil)
	fn.ServeHTTP(w, r)
	if !globalTest {
		t.Error("expected handler to set globalTest")
	}

	// test non-handler non-custom error route
	rndr := &mockRender{}
	m = NewErrorManager(rndr)
	fn = m.Errors(func(w http.ResponseWriter, r *http.Request) error {
		// generic error that isnt added to error manager
		// this should test default case
		return errors.New("test")
	})
	w = httptest.NewRecorder()
	r = httptest.NewRequest("get", "/", nil)
	r = r.WithContext(context.WithValue(context.Background(), CtxLoggerKey, zap.NewNop()))
	fn.ServeHTTP(w, r)
	if rndr.status != http.StatusInternalServerError {
		t.Errorf("expected StatusInternalServerError, got %d", rndr.status)
	}
	if rndr.name != "errors/500" {
		t.Errorf("expected %q, got %q", "errors/500", rndr.name)
	}

	// test non-handler but custom error route
	e1 := errors.New("100 error")
	rndr = &mockRender{}
	m = NewErrorManager(rndr)
	m.Add(NewError(e1, 100, "errors/100", nil))
	fn = m.Errors(func(w http.ResponseWriter, r *http.Request) error {
		// generic error that isnt added to error manager
		// this should test default case
		return e1
	})
	w = httptest.NewRecorder()
	r = httptest.NewRequest("get", "/", nil)
	r = r.WithContext(context.WithValue(context.Background(), CtxLoggerKey, zap.NewNop()))
	fn.ServeHTTP(w, r)
	if rndr.status != 100 {
		t.Errorf("expected 100, got %d", rndr.status)
	}
	if rndr.name != "errors/100" {
		t.Errorf("expected %q, got %q", "errors/100", rndr.name)
	}
}

type mockRender struct {
	status int
	name   string
}

func (mockRender) Data(w io.Writer, status int, v []byte) error      { return nil }
func (mockRender) JSON(w io.Writer, status int, v interface{}) error { return nil }
func (mockRender) Text(w io.Writer, status int, v string) error      { return nil }
func (m *mockRender) HTML(w io.Writer, status int, name string, binding interface{}) error {
	m.status = status
	m.name = name
	return nil
}
func (mockRender) HTMLWithLayout(w io.Writer, status int, name string, binding interface{}, layout string) error {
	return nil
}
