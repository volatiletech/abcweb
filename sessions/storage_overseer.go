package sessions

import (
	"net/http"
	"time"

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
		return "", err
	}

	val, err := s.Storer.Get(sessID)
	if err != nil {
		return "", err
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
		return err
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

	err = s.Storer.Del(sessID)
	if IsNoSessionError(err) {
		return nil
	} else if err != nil {
		return err
	}

	cookie := &http.Cookie{
		// If the browser refuses to delete it, set value to "" so subsequent
		// requests replace it when it does not point to a valid session id.
		Path:     s.options.Path,
		Domain:   s.options.Domain,
		Value:    "",
		Name:     s.options.Name,
		MaxAge:   -1,
		Expires:  time.Now().UTC().AddDate(-1, 0, 0),
		HttpOnly: s.options.HTTPOnly,
		Secure:   s.options.Secure,
	}

	w.(cookieWriter).SetCookie(cookie)

	return nil
}

// Regenerate a new session ID for your current session
func (s *StorageOverseer) Regenerate(w http.ResponseWriter, r *http.Request) error {
	id, err := s.options.getCookieValue(w, r)
	if err != nil {
		return err
	}

	val, err := s.Storer.Get(id)
	if err != nil {
		return err
	}

	// Delete the old session
	_ = s.Storer.Del(id)

	// Generate a new ID
	id = uuid.NewV4().String()

	// Create a new session with the old value
	if err = s.Storer.Set(id, val); err != nil {
		return err
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
		return err
	}

	// Reset the expiry in the client-side cookie
	if s.options.MaxAge != 0 {
		w.(cookieWriter).SetCookie(s.options.makeCookie(sessID))
	}

	// Reset the expiry of the server-side session
	return s.Storer.ResetExpiry(sessID)
}
