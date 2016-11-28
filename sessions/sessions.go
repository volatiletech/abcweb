package sessions

import (
	"errors"
	"net/http"
)

// SessionStorer provides methods to retrieve, add and delete session keys
// and their corresponding values.
type SessionStorer interface {
	Get(r *http.Request) (value string, err error)
	Put(r *http.Request, value string) error
	Del(r *http.Request) error
}

// SessionKey is the key name in the cookie that holds the session id.
const SessionKey = "_SESSION_ID"

// ErrNoSession is returned when a session does not exist.
var ErrNoSession = errors.New("cookie has no session id")
