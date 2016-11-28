package sessions

import (
	"context"
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
	maxAge       int

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
		sessions:     make(map[string]memorySession),
		clientExpiry: clientExpiry,
		serverExpiry: serverExpiry,
		maxAge:       int(clientExpiry.Seconds()),
		secure:       secure,
		httpOnly:     httpOnly,
	}

	// If server expiry is set start the memory cleaner go routine
	if serverExpiry != 0 {
		go m.cleaner(cleanInterval)
	}

	return m, nil
}

// Get returns the value string saved in the session pointed to by the headers
// SessionKey.
func (m *MemoryStorer) Get(r *http.Request) (value string, err error) {
	cookie, err := r.Cookie(SessionKey)
	if err != nil || len(cookie.Value) == 0 {
		return "", ErrNoSession
	}

	m.mut.RLock()
	session, ok := m.sessions[cookie.Value]
	m.mut.RUnlock()
	if !ok {
		return "", ErrNoSession
	}

	return session.value, nil
}

// Put saves the value string to the session pointed to by the headers
// SessionKey. If SessionKey does not exist, Put creates a new session
// with a random unique id.
func (m *MemoryStorer) Put(w http.ResponseWriter, r *http.Request, value string) *http.Request {
	cookie := m.getCookie(r)
	var request *http.Request

	if cookie == nil || cookie.Value == "" {
		cookie = m.makeCookie()
		http.SetCookie(w, cookie)
		// Assign the cookie to the request context so that it can be used
		// again in subsequent calls to Put(). This is required so that
		// subsequent calls to put can locate the session ID that was generated
		// for this cookie, otherwise you will get a new session every time Put()
		// is called.
		ctx := context.WithValue(r.Context(), SessionKey, cookie)
		request = r.WithContext(ctx)
	}

	m.mut.Lock()
	defer m.mut.Unlock()

	m.sessions[cookie.Value] = memorySession{
		expires: time.Now().UTC().Add(m.serverExpiry),
		value:   value,
	}

	return request
}

// Del the session pointed to by the headers SessionKey and remove it from
// the header.
func (m *MemoryStorer) Del(w http.ResponseWriter, r *http.Request) {
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
	cookie.Expires = time.Now().UTC().AddDate(-1, 0, 0)
	cookie.HttpOnly = m.httpOnly
	cookie.Secure = m.secure
	http.SetCookie(w, cookie)

	// id is expected to be the session id
	if id == "" {
		return
	}

	m.mut.Lock()
	delete(m.sessions, id)
	m.mut.Unlock()
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

func (m *MemoryStorer) makeCookie() *http.Cookie {
	return &http.Cookie{
		Name:     SessionKey,
		Value:    uuid.NewV4().String(),
		MaxAge:   m.maxAge,
		Expires:  time.Now().UTC().Add(m.clientExpiry),
		HttpOnly: m.httpOnly,
		Secure:   m.secure,
	}
}

// getCookie returns the cookie stored in the request context. If it does not
// exist in the request context it will attempt to fetch it from the request
// headers. If this fails it will return nil.
func (m *MemoryStorer) getCookie(r *http.Request) *http.Cookie {
	ctxCookie, ok := r.Context().Value(SessionKey).(*http.Cookie)
	if ok && ctxCookie != nil {
		return ctxCookie
	}

	reqCookie, err := r.Cookie(SessionKey)
	if err != nil {
		return nil
	}

	return reqCookie
}
