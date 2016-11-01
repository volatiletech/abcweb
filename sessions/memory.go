package sessions

import (
	"net/http"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
)

// MemorySessions is a session storer implementation for saving sessions
// to memory.
type MemorySessions struct {
	expiry   time.Duration
	secure   bool
	httpOnly bool

	// session storage mutex
	mut sync.RWMutex
	// sessions is the memory storage for the sessions. The map key is the id.
	sessions map[string]memorySession
}

type memorySession struct {
	expires time.Time
	value   string
}

// NewMemorySessions initializes and returns a new MemorySessions object. It
// takes the expiry of how long each session should live in memory, and it
// takes a clean duration which defines how often the clean task should check
// for expired sessions to be removed from memory.
func NewMemorySessions(secure, httpOnly bool, expiry, clean time.Duration) (*MemorySessions, error) {
	if (expiry == 0 && clean != 0) || (expiry != 0 && clean == 0) {
		panic("if clean or expiry is set, the other must also be set")
	}

	m := &MemorySessions{}

	// If expiry is not set, do not start the cleaner routine
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
func (m *MemorySessions) Get(r *http.Request) (value string, err error) {
	cookie, err := r.Cookie(SessionKey)
	if err != nil {
		return "", ErrNoSession
	}

	// cookie.Value is expected to be the session id
	if cookie.Value == "" {
		return "", ErrNoSession
	}

	m.mut.RLock()
	defer m.mut.RUnlock()

	session, ok := m.sessions[cookie.Value]
	if !ok {
		return "", ErrNoSession
	}

	return session.value, nil
}

// Put saves the value string to the session pointed to by the headers
// SessionKey. If SessionKey does not exist, Put creates a new session
// with a random unique id.
func (m *MemorySessions) Put(r *http.Request, value string) {
	var cookie *http.Cookie
	var err error

	cookie, err = r.Cookie(SessionKey)
	if err != nil {
		cookie = m.makeCookie(value)
	} else if cookie.Value == "" {
		cookie.Value = uuid.NewV4().String()
	}

	r.AddCookie(cookie)

	m.mut.Lock()
	defer m.mut.Unlock()

	m.sessions[cookie.Value] = memorySession{
		value:   value,
		expires: cookie.Expires,
	}
}

// Del the session pointed to by the headers SessionKey and remove it from
// the header.
func (m *MemorySessions) Del(r *http.Request) {
	cookie, err := r.Cookie(SessionKey)
	if err != nil {
		return
	}

	// Set cookie to expire on browser side
	cookie.MaxAge = -1
	cookie.Expires = time.Now().AddDate(-1, 0, 0)
	// Re-add the cookie with the new expiration settings
	r.AddCookie(cookie)

	// cookie.Value is expected to be the session id
	if cookie.Value == "" {
		return
	}

	m.mut.Lock()
	defer m.mut.Unlock()

	delete(m.sessions, cookie.Value)
}

// sleepFunc is a test harness
var sleepFunc = time.Sleep

func (m *MemorySessions) cleaner(loop time.Duration) {
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

func (m *MemorySessions) makeCookie(value string) *http.Cookie {
	return &http.Cookie{
		Name:     uuid.NewV4().String(),
		Value:    value,
		MaxAge:   int(m.expiry.Seconds()),
		Expires:  time.Now().Add(m.expiry),
		HttpOnly: m.httpOnly,
		Secure:   m.secure,
	}
}
