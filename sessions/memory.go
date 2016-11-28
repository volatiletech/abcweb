package sessions

import (
	"sync"
	"time"
)

// MemoryStorer is a session storer implementation for saving sessions
// to memory.
type MemoryStorer struct {
	serverExpiry time.Duration

	// session storage mutex
	mut sync.RWMutex
	// sessions is the memory storage for the sessions. The map key is the id.
	sessions map[string]memorySession
}

type memorySession struct {
	expires time.Time
	value   string
}

// NewDefaultMemoryStorer returns a MemoryStorer object with default values.
// The default values are:
// serverExpiry: 1 week (clear session stored on server after 1 week)
// cleanInterval: 2 days (delete sessions older than serverExpiry every hour)
func NewDefaultMemoryStorer() (*MemoryStorer, error) {
	return NewMemoryStorer(time.Hour*24*7, time.Hour)
}

// NewMemoryStorer initializes and returns a new MemoryStorer object. It
// takes the clientExpiry of how long each session cookie should live in
// the clients browser, the serverExpiry of how long each session should live
// in memory, and a cleanInterval duration which defines how often the clean
// task should check for server expired sessions to be removed from memory.
// Persistent storage can be attained by setting serverExpiry and cleanInterval
// to zero, however the memory will be wiped when the server is restarted.
func NewMemoryStorer(serverExpiry, cleanInterval time.Duration) (*MemoryStorer, error) {
	if (serverExpiry != 0 && cleanInterval == 0) || (cleanInterval != 0 && serverExpiry == 0) {
		panic("if server expiry or clean interval is set, the other must also be set")
	}

	m := &MemoryStorer{
		sessions:     make(map[string]memorySession),
		serverExpiry: serverExpiry,
	}

	// If server expiry is set start the memory cleaner go routine
	if serverExpiry != 0 {
		go m.cleaner(cleanInterval)
	}

	return m, nil
}

// Get returns the value string saved in the session pointed to by the headers
// Key.
func (m *MemoryStorer) Get(key string) (value string, err error) {
	m.mut.RLock()
	session, ok := m.sessions[key]
	m.mut.RUnlock()
	if !ok {
		return "", ErrNoSession
	}

	return session.value, nil
}

// Put saves the value string to the session pointed to by the headers
// Key. If Key does not exist, Put creates a new session
// with a random unique id.
func (m *MemoryStorer) Put(key, value string) error {
	m.mut.Lock()
	m.sessions[key] = memorySession{
		expires: time.Now().UTC().Add(m.serverExpiry),
		value:   value,
	}
	m.mut.Unlock()

	return nil
}

// Del the session pointed to by the headers Key and remove it from
// the header.
func (m *MemoryStorer) Del(key string) error {
	m.mut.Lock()
	delete(m.sessions, key)
	m.mut.Unlock()

	return nil
}

// sleepFunc is a test harness
var sleepFunc = time.Sleep

func (m *MemoryStorer) cleaner(loop time.Duration) {
	for {
		sleepFunc(loop)
		t := time.Now().UTC()
		m.mut.Lock()
		for id, session := range m.sessions {
			if t.After(session.expires) {
				delete(m.sessions, id)
			}
		}
		m.mut.Unlock()
	}
}
