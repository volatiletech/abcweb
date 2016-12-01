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
	cookieID := s.getCookieID(r)
	if len(cookieID) == 0 {
		return "", errNoSession{}
	}

	val, err := s.storer.Get(cookieID)
	if err != nil {
		return "", err
	}

	return val, nil
}

// Put looks in the cookie for the session ID and modifies the session with the new value.
// If the session does not exist it creates a new one.
func (s *StorageOverseer) Put(w http.ResponseWriter, r *http.Request, value string) (*http.Request, error) {
	// Reuse the existing cookie ID if it exists
	cookieID := s.getCookieID(r)
	if len(cookieID) == 0 {
		cookieID = uuid.NewV4().String()
	}

	cookie := s.options.makeCookie(cookieID)
	http.SetCookie(w, cookie)

	// Assign the cookie to the request context so that it can be used
	// again in subsequent calls to Put(). This is required so that
	// subsequent calls to put can locate the session ID that was generated
	// for this cookie, otherwise you will get a new session every time Put()
	// is called.
	ctx := context.WithValue(r.Context(), s.options.Name, cookieID)
	r = r.WithContext(ctx)

	err := s.storer.Put(cookieID, value)
	return r, err
}

// Del deletes the session if it exists and sets the session cookie to expire instantly.
func (s *StorageOverseer) Del(w http.ResponseWriter, r *http.Request) error {
	err := s.storer.Del(s.getCookieID(r))
	if IsNoSessionError(err) {
		return nil
	} else if err != nil {
		return err
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
	return nil
}

// getCookie returns the cookie ID stored in the request context. If it does not
// exist in the request context it will attempt to fetch it from the request
// headers. If this fails it will return nil.
func (s *StorageOverseer) getCookieID(r *http.Request) string {
	cookieID, ok := r.Context().Value(s.options.Name).(string)
	if ok && cookieID != "" {
		return cookieID
	}

	reqCookie, err := r.Cookie(s.options.Name)
	if err != nil {
		return ""
	}

	return reqCookie.Value
}
