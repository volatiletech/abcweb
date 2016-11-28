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

// NewStorageOverseer returns a new overseer
func NewStorageOverseer(opts CookieOptions, storer Storer) *StorageOverseer {
	return &StorageOverseer{
		storer:  storer,
		options: opts,
	}
}

func (s *StorageOverseer) Get(w http.ResponseWriter, r *http.Request) (value string, err error) {
	cookieID := s.getCookieID(r)
	if len(cookieID) == 0 {
		return "", ErrNoSession
	}

	val, err := s.storer.Get(cookieID)
	if err != nil {
		return "", err
	}

	return val, nil
}

func (s *StorageOverseer) Put(w http.ResponseWriter, r *http.Request, value string) (*http.Request, error) {
	cookieID := s.getCookieID(r)
	if len(cookieID) == 0 {
		cookieID = uuid.NewV4().String()
		cookie := s.makeCookie(cookieID)
		http.SetCookie(w, cookie)

		// Assign the cookie to the request context so that it can be used
		// again in subsequent calls to Put(). This is required so that
		// subsequent calls to put can locate the session ID that was generated
		// for this cookie, otherwise you will get a new session every time Put()
		// is called.
		ctx := context.WithValue(r.Context(), s.options.Name, cookieID)
		r = r.WithContext(ctx)
	}

	err := s.storer.Put(cookieID, value)
	return r, err
}

func (s *StorageOverseer) Del(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(s.options.Name)
	if err != nil {
		return err
	}

	// Set cookie to expire on browser side
	id := cookie.Value
	// If the browser refuses to delete it, set value to "" so subsequent
	// requests replace it when it does not point to a valid session id.
	cookie.Value = ""
	cookie.MaxAge = -1
	cookie.Expires = time.Now().UTC().AddDate(-1, 0, 0)
	cookie.HttpOnly = s.options.HTTPOnly
	cookie.Secure = s.options.Secure
	http.SetCookie(w, cookie)

	// id is expected to be the session id
	if id == "" {
		return nil
	}

	return s.storer.Del(id)
}

func (s *StorageOverseer) makeCookie(cookieID string) *http.Cookie {
	return &http.Cookie{
		Name:     s.options.Name,
		Value:    cookieID,
		MaxAge:   int(s.options.MaxAge.Seconds()),
		Expires:  time.Now().UTC().Add(s.options.MaxAge),
		HttpOnly: s.options.HTTPOnly,
		Secure:   s.options.Secure,
	}
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
