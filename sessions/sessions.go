package sessions

import (
	"errors"
	"net/http"
)

// Storer provides methods to retrieve, add and delete session keys
// and their corresponding values.
type Storer interface {
	Get(key string) (value string, err error)
	Put(key, value string) error
	Del(key string) error
}

// Overseer of session cookies
type Overseer interface {
	Get(w http.ResponseWriter, r *http.Request) (value string, err error)
	Put(w http.ResponseWriter, r *http.Request, value string) (cr *http.Request, err error)
	Del(w http.ResponseWriter, r *http.Request) (err error)
}

// Key is the key name in the cookie that holds the session id.
const Key = "_SESSION_ID"

// ErrNoSession is returned when a session does not exist.
var ErrNoSession = errors.New("cookie has no session id")
