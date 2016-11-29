package sessions

import (
	"sync"
	"time"
)

// MemoryStorer is a session storer implementation for saving sessions
// to memory.
type MemoryStorer struct {
	maxAge time.Duration

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
// maxAge: 1 week (clear session stored on server after 1 week)
// cleanInterval: 1 hour (delete sessions older than maxAge every hour)
func NewDefaultMemoryStorer() (*MemoryStorer, error) {
	return NewMemoryStorer(time.Hour*24*7, time.Hour)
}

// NewMemoryStorer initializes and returns a new MemoryStorer object.
// It takes the maxAge of how long each session should live in memory,
// and a cleanInterval duration which defines how often the clean
// task should check for maxAge sessions to be removed from memory.
// Persistent storage can be attained by setting maxAge and cleanInterval
// to zero, however the memory will be wiped when the server is restarted.
func NewMemoryStorer(maxAge, cleanInterval time.Duration) (*MemoryStorer, error) {
	if (maxAge != 0 && cleanInterval == 0) || (cleanInterval != 0 && maxAge == 0) {
		panic("if max age or clean interval is set, the other must also be set")
	}

	m := &MemoryStorer{
		sessions: make(map[string]memorySession),
		maxAge:   maxAge,
	}

	// If max age is set start the memory cleaner go routine
	if maxAge != 0 {
		go m.cleaner(cleanInterval)
	}

	return m, nil
}

// Get returns the value string saved in the session pointed to by the
// session id key.
func (m *MemoryStorer) Get(key string) (value string, err error) {
	m.mut.RLock()
	session, ok := m.sessions[key]
	m.mut.RUnlock()
	if !ok {
		return "", errNoSession{}
	}

	return session.value, nil
}

// Put saves the value string to the session pointed to by the session id key.
func (m *MemoryStorer) Put(key, value string) error {
	m.mut.Lock()
	m.sessions[key] = memorySession{
		expires: time.Now().UTC().Add(m.maxAge),
		value:   value,
	}
	m.mut.Unlock()

	return nil
}

// Del the session pointed to by the session id key and remove it.
func (m *MemoryStorer) Del(key string) error {
	m.mut.Lock()
	delete(m.sessions, key)
	m.mut.Unlock()

	return nil
}

// memorySleepFunc is a test harness
var memorySleepFunc = time.Sleep

func (m *MemoryStorer) cleaner(loop time.Duration) {
	for {
		memorySleepFunc(loop)
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
