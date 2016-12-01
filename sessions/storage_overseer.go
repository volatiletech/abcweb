package sessions

import (
	"context"
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"
)

// StorageOverseer holds cookie related variables and a session storer
type StorageOverseer struct {
	storer  Storer
	options CookieOptions
}

// NewStorageOverseer returns a new storage overseer
func NewStorageOverseer(opts CookieOptions, storer Storer) *StorageOverseer {
	return &StorageOverseer{
		storer:  storer,
		options: opts,
	}
}

// Get looks in the cookie for the session ID and retrieves the value string stored in the session.
func (s *StorageOverseer) Get(w http.ResponseWriter, r *http.Request) (value string, err error) {
	sessID, err := s.options.getCookieValue(r)
	if err != nil {
		return "", err
	}

	val, err := s.storer.Get(sessID)
	if err != nil {
		return "", err
	}

	return val, nil
}

// Put looks in the cookie for the session ID and modifies the session with the new value.
// If the session does not exist it creates a new one.
func (s *StorageOverseer) Put(w http.ResponseWriter, r *http.Request, value string) (*http.Request, error) {
	// Reuse the existing cookie ID if it exists
	sessID, _ := s.options.getCookieValue(r)

	if len(sessID) == 0 {
		sessID = uuid.NewV4().String()
	}

	cookie := s.options.makeCookie(sessID)
	http.SetCookie(w, cookie)

	// Assign the cookie to the request context so that it can be used
	// again in subsequent calls to Put(). This is required so that
	// subsequent calls to put can locate the session ID that was generated
	// for this cookie, otherwise you will get a new session every time Put()
	// is called.
	ctx := context.WithValue(r.Context(), s.options.Name, sessID)
	r = r.WithContext(ctx)

	err := s.storer.Put(sessID, value)
	return r, err
}

// Del deletes the session if it exists and sets the session cookie to expire instantly.
func (s *StorageOverseer) Del(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	sessID, err := s.options.getCookieValue(r)
	if err != nil {
		return r, nil
	}

	err = s.storer.Del(sessID)
	if IsNoSessionError(err) {
		return r, nil
	} else if err != nil {
		return r, err
	}

	cookie := &http.Cookie{
		// If the browser refuses to delete it, set value to "" so subsequent
		// requests replace it when it does not point to a valid session id.
		Value:    "",
		Name:     s.options.Name,
		MaxAge:   -1,
		Expires:  time.Now().UTC().AddDate(-1, 0, 0),
		HttpOnly: s.options.HTTPOnly,
		Secure:   s.options.Secure,
	}

	http.SetCookie(w, cookie)
	return r, nil
}
