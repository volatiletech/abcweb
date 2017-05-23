package abcsessions

import (
	"net/http"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// StorageOverseer holds cookie related variables and a session storer
type StorageOverseer struct {
	Storer  Storer
	options CookieOptions
	resetExpiryMiddleware
}

// NewStorageOverseer returns a new storage overseer
func NewStorageOverseer(opts CookieOptions, storer Storer) *StorageOverseer {
	if len(opts.Name) == 0 {
		panic("cookie name must be provided")
	}

	o := &StorageOverseer{
		Storer:  storer,
		options: opts,
	}

	o.resetExpiryMiddleware.resetter = o

	return o
}

// Get looks in the cookie for the session ID and retrieves the value string stored in the session.
func (s *StorageOverseer) Get(w http.ResponseWriter, r *http.Request) (value string, err error) {
	sessID, err := s.options.getCookieValue(w, r)
	if err != nil {
		return "", errors.Wrap(err, "unable to get session id from cookie")
	}

	val, err := s.Storer.Get(sessID)
	if err != nil {
		return "", errors.Wrap(err, "unable to get session value")
	}

	return val, nil
}

// Set looks in the cookie for the session ID and modifies the session with the new value.
// If the session does not exist it creates a new one.
func (s *StorageOverseer) Set(w http.ResponseWriter, r *http.Request, value string) error {
	// Reuse the existing cookie ID if it exists
	sessID, _ := s.options.getCookieValue(w, r)

	if len(sessID) == 0 {
		sessID = uuid.NewV4().String()
	}

	err := s.Storer.Set(sessID, value)
	if err != nil {
		return errors.Wrap(err, "unable to set session value")
	}

	w.(cookieWriter).SetCookie(s.options.makeCookie(sessID))

	return nil
}

// Del deletes the session if it exists and sets the session cookie to expire instantly.
func (s *StorageOverseer) Del(w http.ResponseWriter, r *http.Request) error {
	sessID, err := s.options.getCookieValue(w, r)
	if err != nil {
		return nil
	}

	s.options.deleteCookie(w)

	err = s.Storer.Del(sessID)
	if IsNoSessionError(err) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "unable to delete server-side session")
	}

	return nil
}

// Regenerate a new session ID for your current session
func (s *StorageOverseer) Regenerate(w http.ResponseWriter, r *http.Request) error {
	id, err := s.options.getCookieValue(w, r)
	if err != nil {
		return errors.Wrap(err, "unable to get session id from cookie")
	}

	val, err := s.Storer.Get(id)
	if err != nil {
		return errors.Wrap(err, "unable to get session value")
	}

	// Delete the old session
	_ = s.Storer.Del(id)

	// Generate a new ID
	id = uuid.NewV4().String()

	// Create a new session with the old value
	if err = s.Storer.Set(id, val); err != nil {
		return errors.Wrap(err, "unable to set session value")
	}

	// Override the old cookie with the new cookie
	w.(cookieWriter).SetCookie(s.options.makeCookie(id))

	return nil
}

// SessionID returns the session ID stored in the cookie's value field.
// It will return a errNoSession error if no session exists.
func (s *StorageOverseer) SessionID(w http.ResponseWriter, r *http.Request) (string, error) {
	return s.options.getCookieValue(w, r)
}

// ResetExpiry resets the age of the session to time.Now(), so that
// MaxAge calculations are renewed
func (s *StorageOverseer) ResetExpiry(w http.ResponseWriter, r *http.Request) error {
	sessID, err := s.options.getCookieValue(w, r)
	if err != nil {
		return errors.Wrap(err, "unable to get session id from cookie")
	}

	// Reset the expiry of the server-side session
	err = s.Storer.ResetExpiry(sessID)
	if err != nil {
		return errors.Wrap(err, "unable to reset expiry of server side session")
	}

	// Reset the expiry in the client-side cookie
	if s.options.MaxAge != 0 {
		w.(cookieWriter).SetCookie(s.options.makeCookie(sessID))
	}

	return nil
}
