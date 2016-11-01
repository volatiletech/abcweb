package sessions

import (
	"net/http"

	uuid "github.com/satori/go.uuid"
)

// DiskSession is a session storer implementation for saving sessions
// to disk.
type DiskSession struct {
	folderPath string
}

// New initializes and returns a new DiskSession.
func (d DiskSession) New(folderPath string) (DiskSession, error) {

	return DiskSession{}, nil
}

// Get returns the value string saved in the session pointed to by the headers
// SessionKey.
func (d DiskSession) Get(header http.Header) (value string, err error) {
	id := header.Get(SessionKey)
	if id == "" {
		return id, ErrNoSession
	}

	return d.GetID(id)
}

// Put saves the value string to the session pointed to by the headers
// SessionKey. If SessionKey does not exist, Put creates a new session
// with a random unique id.
func (d DiskSession) Put(header http.Header, value string) error {
	id := header.Get(SessionKey)
	if id == "" {
		return d.PutID(uuid.NewV4().String(), value)
	}

	return d.PutID(id, value)
}

// Del the session pointed to by the headers SessionKey and remove it from
// the header.
func (d DiskSession) Del(header http.Header) error {
	return d.DelID(header, header.Get(SessionKey))
}

// GetID returns the value string saved in the session pointed to by id.
func (d DiskSession) GetID(id string) (value string, err error) {
	return "", nil
}

// PutID saves the value string to the session pointed to by id. If id
// does not exist, PutID creates a new session with a random unique id.
func (d DiskSession) PutID(id string, value string) error {
	return nil
}

// DelID deletes the session pointed to by id and removes it from the header.
func (d DiskSession) DelID(header http.Header, id string) error {
	return nil
}
