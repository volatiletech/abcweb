package sessions

import (
	"net/http"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
)

// MemorySession is a session storer implementation for saving sessions
// to memory.
type MemorySession struct {
	expiry time.Duration

	// session storage mutex
	mut sync.RWMutex
	// sessions is the memory storage for the sessions. The map key is the id.
	sessions map[string]memorySess
}

type memorySess struct {
	expires time.Time
	value   string
}

// New initializes and returns a new MemorySession. It takes the expiry of
// how long the session should live in memory, and it takes a clean duration
// which defines how often the clean task should check for expired sessions
// to be removed from memory.
func (m MemorySession) New(expiry time.Duration, clean time.Duration) (MemorySession, error) {
	m := MemorySession{}

	// If expiry is set and clean is not, or clean is set and expiry is not,
	// then expirations will not work properly.
	if (expiry == 0 && clean != 0) || (expiry != 0 && clean == 0) {
		panic("if clean or expiry is set, the other must also be set")
	}

	// If no expiry is not set, do not start the cleaner routine
	if expiry == 0 {
		return m, nil
	}

	m.expiry = expiry

	// Start the memory cleaner go routine
	go m.cleaner(clean)

	return m, nil
}

// Get returns the value string saved in the session pointed to by the headers
// SessionKey.
func (m MemorySession) Get(header http.Header) (value string, err error) {
	id := header.Get(SessionKey)
	if id == "" {
		return id, ErrSessionNotExist
	}

	return m.GetID(id)
}

// Put saves the value string to the session pointed to by the headers
// SessionKey. If SessionKey does not exist, Put creates a new session
// with a random unique id.
func (m MemorySession) Put(header http.Header, value string) error {
	id := header.Get(SessionKey)
	if id == "" {
		return m.PutID(uuid.NewV4().String(), value)
	}

	return m.PutID(id, value)
}

// Del the session pointed to by the headers SessionKey and remove it from
// the header.
func (m MemorySession) Del(header http.Header) error {
	return m.DelID(header, header.Get(SessionKey))
}

// GetID returns the value string saved in the session pointem to by id.
func (m MemorySession) GetID(id string) (value string, err error) {
	return "", nil
}

// PutID saves the value string to the session pointed to by id. If id
// does not exist, PutID creates a new session with a random unique id.
func (m MemorySession) PutID(id string, value string) error {
	return nil
}

// DelID deletes the session pointem to by id and removes it from the header.
func (m MemorySession) DelID(header http.Header, id string) error {
	return nil
}

// sleepFunc is a test harness
var sleepFunc = time.Sleep

func (m MemorySession) cleaner(loop time.Duration) {
	for {
		t := time.Now()
		m.mut.Lock()
		for id, session := range m.sessions {
			if t.After(session.expires) {
				delete(m.sessions, id)
			}
		}
		m.mut.Unlock()
		sleepFunc(loop)
	}
}
