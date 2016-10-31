package sessions

import (
	"errors"
	"net/http"
)

// SessionStorer provides methods to retrieve, add and delete session keys
// and their corresponding values.
type SessionStorer interface {
	Get(header http.Header) (value string, err error)
	Put(header http.Header, value string) error
	Del(header http.Header) error

	GetID(id string) (value string, err error)
	PutID(id string, value string) error
	DelID(header http.Header, id string) error
}

// SessionKey is the key name in the cookie that holds the session id.
const SessionKey = "_SESSION_ID"

// ErrSessionNotExist is returned when a session does not exist.
var ErrSessionNotExist = errors.New("cookie has no session id")
