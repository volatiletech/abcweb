package sessions

import (
	"errors"
	"net/http"
	"time"
)

// Storer provides methods to retrieve, add and delete session keys
// and their corresponding values.
type Storer interface {
	Get(key string) (value string, err error)
	Put(key, value string) error
	Del(key string) error
}

// Key is the key name in the cookie that holds the session id.
const Key = "_SESSION_ID"

// ErrNoSession is returned when a session does not exist.
var ErrNoSession = errors.New("cookie has no session id")

// Overseer holds cookie related variables and a session storer
type Overseer struct {
	storer Storer

	clientExpiry time.Duration
	maxAge       int

	secure   bool
	httpOnly bool
}

// NewDefaultOverseer returns a new overseer with smart default values
func NewDefaultOverseer(secure bool) (*Overseer, error) {
	return NewOverseer(true, true, 0)
}

// NewOverseer returns a new overseer
func NewOverseer(storer Storer, secure, httpOnly bool, clientExpiry time.Duration) (*Overseer, error) {
	return &Overseer{
		storer:       storer,
		secure:       secure,
		httpOnly:     httpOnly,
		clientExpiry: clientExpiry,
		maxAge:       int(clientExpiry.Seconds()),
	}
}

func (o *Overseer) Get(r *http.Request) (value string, err error) {
	cookie, err := r.Cookie(Key)
	if err != nil || len(cookie.Value) == 0 {
		return "", ErrNoSession
	}

	o.storer.Get(cookie.Value)
	val, err := o.storer.Get(cookie.Value)
	if err != nil {
		return "", err
	}

	return val, nil
}

func (o *Overseer) Put(r *http.Request, value string) (err error) {

}

func (o *Overseer) Del(r *http.Request) error {

}
