package abcsessions

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareWrite(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	response := newSessionsResponseWriter(w)

	if response.wroteHeader {
		t.Error("expected false")
	}

	_, err := response.Write([]byte("hi"))
	if err != nil {
		t.Error(err)
	}

	if !response.wroteHeader {
		t.Error("expected true")
	}

	if w.Result().StatusCode != http.StatusOK {
		t.Error("expected status ok, got:", w.Result().StatusCode)
	}
}

func TestMiddlewareWriteHeader(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	response := newSessionsResponseWriter(w)

	response.cookies = make(map[string]*http.Cookie)

	response.cookies["lol"] = &http.Cookie{Name: "lol", Value: "test1"}
	response.cookies["hehe"] = &http.Cookie{Name: "hehe", Value: "test2"}
	response.WriteHeader(200)

	cookies := w.Result().Cookies()
	if len(cookies) != 2 {
		t.Error("expected cookies len 2, got:", len(cookies))
	}

	for _, c := range cookies {
		var found bool
		for _, rc := range response.cookies {
			if c.Name == rc.Name {
				found = true
			}
		}
		if !found {
			t.Errorf("could not find cookie with name %s in cookies", c.Name)
		}
	}
}

func TestMiddlewareSetCookie(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	response := newSessionsResponseWriter(w)

	response.SetCookie(&http.Cookie{Name: "lolcats"})
	response.SetCookie(&http.Cookie{Name: "catlollers"})

	if len(response.cookies) != 2 {
		t.Error("expected len 2, got:", len(response.cookies))
	}

	if _, ok := response.cookies["lolcats"]; !ok {
		t.Error("expected lolcats to be set")
	}
	if _, ok := response.cookies["catlollers"]; !ok {
		t.Error("expected catlollers to be set")
	}
}

func TestMiddleware(t *testing.T) {
	t.Parallel()

	fn := func(w http.ResponseWriter, r *http.Request) {
		if _, ok := w.(*sessionsResponseWriter); !ok {
			t.Error("was not of type response")
		}
	}

	hf := Middleware(http.HandlerFunc(fn))

	hf.ServeHTTP(nil, nil)
}

type middlewareOverseerMock struct {
	called bool
	resetExpiryMiddleware
}

func (m *middlewareOverseerMock) ResetExpiry(w http.ResponseWriter, r *http.Request) error {
	m.called = true
	return nil
}

func TestResetMiddleware(t *testing.T) {
	t.Parallel()

	o := &middlewareOverseerMock{}
	o.resetExpiryMiddleware.resetter = o

	fn := func(w http.ResponseWriter, r *http.Request) {}

	hf := o.ResetMiddleware(http.HandlerFunc(fn))

	hf.ServeHTTP(nil, nil)
	if !o.called {
		t.Error("expected called true")
	}
}

func TestMiddlewareWithReset(t *testing.T) {
	t.Parallel()

	o := &middlewareOverseerMock{}
	o.resetExpiryMiddleware.resetter = o

	fn := func(w http.ResponseWriter, r *http.Request) {
		if _, ok := w.(*sessionsResponseWriter); !ok {
			t.Error("was not of type response")
		}
	}

	hf := o.MiddlewareWithReset(http.HandlerFunc(fn))

	hf.ServeHTTP(nil, nil)
	if !o.called {
		t.Error("expected called true")
	}
}
