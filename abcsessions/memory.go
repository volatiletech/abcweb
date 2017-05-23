package abcsessions

import (
	"sync"
	"time"
)

// MemoryStorer is a session storer implementation for saving sessions
// to memory.
type MemoryStorer struct {
	// sessions is the memory storage for the sessions. The map key is the id.
	sessions map[string]memorySession
	// How long sessions take to expire on disk
	maxAge time.Duration
	// How often the memory map should be polled for maxAge expired sessions
	cleanInterval time.Duration
	// session storage mutex
	mut sync.RWMutex
	// wg is used to manage the cleaner go routines
	wg sync.WaitGroup
	// quit channel for exiting the cleaner loop
	quit chan struct{}
}

type memorySession struct {
	expires time.Time
	value   string
}

// NewDefaultMemoryStorer returns a MemoryStorer object with default values.
// The default values are:
// maxAge: 2 days (clear session stored on server after 2 days)
// cleanInterval: 1 hour (delete sessions older than maxAge every hour)
func NewDefaultMemoryStorer() (*MemoryStorer, error) {
	return NewMemoryStorer(time.Hour*24*2, time.Hour)
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
		sessions:      make(map[string]memorySession),
		maxAge:        maxAge,
		cleanInterval: cleanInterval,
	}

	return m, nil
}

// All keys in the memory store
func (m *MemoryStorer) All() ([]string, error) {
	m.mut.RLock()
	defer m.mut.RUnlock()

	sessions := make([]string, len(m.sessions))

	i := 0
	for id := range m.sessions {
		sessions[i] = id
		i++
	}

	return sessions, nil
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

// Set saves the value string to the session pointed to by the session id key.
func (m *MemoryStorer) Set(key, value string) error {
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

// ResetExpiry resets the expiry of the key
func (m *MemoryStorer) ResetExpiry(key string) error {
	m.mut.RLock()
	session, ok := m.sessions[key]
	m.mut.RUnlock()
	if !ok {
		return errNoSession{}
	}

	session.expires = time.Now().UTC().Add(m.maxAge)
	m.sessions[key] = session
	return nil
}

// Clean checks all sessions in memory to see if they are older than
// maxAge by checking their expiry. If it finds an expired session
// it will remove it from memory.
func (m *MemoryStorer) Clean() {
	t := time.Now().UTC()
	m.mut.Lock()
	for id, session := range m.sessions {
		if t.After(session.expires) {
			delete(m.sessions, id)
		}
	}
	m.mut.Unlock()
}

// StartCleaner starts the memory session cleaner go routine. This go routine
// will delete expired sessions from the memory map on the cleanInterval interval.
func (m *MemoryStorer) StartCleaner() {
	if m.maxAge == 0 || m.cleanInterval == 0 {
		panic("both max age and clean interval must be set to non-zero")
	}

	// init quit chan
	m.quit = make(chan struct{})

	m.wg.Add(1)

	// Start the cleaner infinite loop go routine.
	// StopCleaner() can be used to kill this go routine.
	go m.cleanerLoop()
}

// StopCleaner stops the cleaner go routine
func (m *MemoryStorer) StopCleaner() {
	close(m.quit)
	m.wg.Wait()
}

// cleanerLoop executes the Clean() method every time cleanInterval elapses.
// StopCleaner() can be used to kill this go routine loop.
func (m *MemoryStorer) cleanerLoop() {
	defer m.wg.Done()

	t, c := timerTestHarness(m.cleanInterval)

	select {
	case <-c:
		m.Clean()
		t.Reset(m.cleanInterval)
	case <-m.quit:
		t.Stop()
		return
	}
}
