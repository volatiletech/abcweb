package sessions

import (
	"net/http"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
)

// MemoryStorer is a session storer implementation for saving sessions
// to memory.
type MemoryStorer struct {
	clientExpiry time.Duration
	serverExpiry time.Duration

	secure   bool
	httpOnly bool

	// session storage mutex
	mut sync.RWMutex
	// sessions is the memory storage for the sessions. The map key is the id.
	sessions memorySessions
}

type memorySession struct {
	expires time.Time
	value   string
}

type memorySessions map[string]*memorySession

// NewDefaultMemoryStorer returns a MemoryStorer object with default values.
// The default values are:
// secure: User supplied (only transmit cookies over HTTPS connections)
// httpOnly: Enabled (disable client-side scripts accessing cookies)
// clientExpiry: 0 (browser expires client-side cookie on close)
// serverExpiry: 1 week (clear session stored on server after 1 week)
// cleanInterval: 2 days (delete sessions older than serverExpiry every hour)
func NewDefaultMemoryStorer(secure bool) (*MemoryStorer, error) {
	return NewMemoryStorer(secure, true, 0, time.Hour*24*7, time.Hour)
}

// NewMemoryStorer initializes and returns a new MemoryStorer object. It
// takes the clientExpiry of how long each session cookie should live in
// the clients browser, the serverExpiry of how long each session should live
// in memory, and a cleanInterval duration which defines how often the clean
// task should check for server expired sessions to be removed from memory.
func NewMemoryStorer(secure, httpOnly bool, clientExpiry, serverExpiry, cleanInterval time.Duration) (*MemoryStorer, error) {
	if (serverExpiry != 0 && cleanInterval == 0) || (cleanInterval != 0 && serverExpiry == 0) {
		panic("if server expiry or clean interval is set, the other must also be set")
	}

	m := &MemoryStorer{
		sessions:     make(map[string]*memorySession),
		clientExpiry: clientExpiry,
		serverExpiry: serverExpiry,
		secure:       secure,
		httpOnly:     httpOnly,
	}

	// If server expiry is not set, do not start the cleaner routine
	if serverExpiry == 0 {
		return m, nil
	}

	// Start the memory cleaner go routine
	go m.cleaner(cleanInterval)

	return m, nil
}

// Get returns the value string saved in the session pointed to by the headers
// SessionKey.
func (m *MemoryStorer) Get(r *http.Request) (value string, err error) {
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
func (m *MemoryStorer) Put(r *http.Request, w http.ResponseWriter, value string) {
	var cookie *http.Cookie
	var err error

	expires := time.Now().Add(m.clientExpiry)

	cookie, err = r.Cookie(SessionKey)
	if err != nil {
		cookie = m.makeCookie()
		http.SetCookie(w, cookie)
		// add cookie to request as well in case subsequent calls to Put are made
		r.AddCookie(cookie)
	} else if cookie.Value == "" {
		cookie = m.makeCookie()
		http.SetCookie(w, cookie)
	}

	m.mut.Lock()
	defer m.mut.Unlock()

	sess, ok := m.sessions[cookie.Value]
	if ok {
		sess.value = value
	} else {
		m.sessions[cookie.Value] = &memorySession{
			value:   value,
			expires: expires,
		}
	}
}

// Del the session pointed to by the headers SessionKey and remove it from
// the header.
func (m *MemoryStorer) Del(r *http.Request, w http.ResponseWriter) {
	cookie, err := r.Cookie(SessionKey)
	if err != nil {
		return
	}

	// Set cookie to expire on browser side
	id := cookie.Value
	// If the browser refuses to delete it, set value to "" so subsequent
	// requests replace it when it does not point to a valid session id.
	cookie.Value = ""
	cookie.MaxAge = -1
	cookie.Expires = time.Now().AddDate(-1, 0, 0)
	cookie.HttpOnly = m.httpOnly
	cookie.Secure = m.secure
	http.SetCookie(w, cookie)

	// id is expected to be the session id
	if id == "" {
		return
	}

	m.mut.Lock()
	defer m.mut.Unlock()

	delete(m.sessions, id)
}

// sleepFunc is a test harness
var sleepFunc = time.Sleep

func (m *MemoryStorer) cleaner(loop time.Duration) {
	for {
		sleepFunc(loop)
		t := time.Now()
		m.mut.Lock()
		for id, session := range m.sessions {
			if t.After(session.expires) {
				delete(m.sessions, id)
			}
		}
		m.mut.Unlock()
	}
}

func (m *MemoryStorer) makeCookie() *http.Cookie {
	return &http.Cookie{
		Name:     SessionKey,
		Value:    uuid.NewV4().String(),
		MaxAge:   int(m.clientExpiry.Seconds()),
		Expires:  time.Now().Add(m.clientExpiry),
		HttpOnly: m.httpOnly,
		Secure:   m.secure,
	}
}
