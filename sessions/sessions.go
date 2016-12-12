package sessions

import (
	"encoding/json"
	"net/http"
	"time"
)

// Storer provides methods to retrieve, add and delete sessions.
type Storer interface {
	// All returns all keys in the store
	All() (keys []string, err error)
	Get(key string) (value string, err error)
	Set(key, value string) error
	Del(key string) error
	ResetExpiry(key string) error
}

// Overseer of session cookies
type Overseer interface {
	Resetter
	// Get the value stored in a session
	Get(w http.ResponseWriter, r *http.Request) (value string, err error)
	// Set creates or updates a session with value
	Set(w http.ResponseWriter, r *http.Request, value string) error
	// Delete a session
	Del(w http.ResponseWriter, r *http.Request) error
	// Regenerate a new session id for your session
	Regenerate(w http.ResponseWriter, r *http.Request) error
	// SessionID returns the session id for your session
	SessionID(w http.ResponseWriter, r *http.Request) (id string, err error)
}

// Resetter has reset functions
type Resetter interface {
	// ResetExpiry resets the age of the session to time.Now(), so that
	// MaxAge calculations are renewed
	ResetExpiry(w http.ResponseWriter, r *http.Request) error
	// ResetMiddleware will reset the users session expiry on every request
	ResetMiddleware(next http.Handler) http.Handler
}

// timer interface is used to mock the test harness for disk and memory storers
type timer interface {
	Stop() bool
	Reset(time.Duration) bool
}

type noSessionInterface interface {
	NoSession()
}
type noMapKeyInterface interface {
	NoMapKey()
}

type errNoSession struct{}
type errNoMapKey struct{}

func (errNoSession) NoSession() {}
func (errNoMapKey) NoMapKey()   {}

func (errNoSession) Error() string {
	return "session does not exist"
}

func (errNoMapKey) Error() string {
	return "session map key does not exist"
}

// IsNoSessionError checks an error to see if it means that there was no session
func IsNoSessionError(err error) bool {
	_, ok := err.(noSessionInterface)
	return ok
}

// IsNoMapKeyError checks an error to see if it means that there was no session map key
func IsNoMapKeyError(err error) bool {
	_, ok := err.(noMapKeyInterface)
	return ok
}

// timerTestHarness allows us to control the timer channels manually in the
// disk and memory storer tests so that we can trigger cleans at will
var timerTestHarness = func(d time.Duration) (timer, <-chan time.Time) {
	t := time.NewTimer(d)
	return t, t.C
}

// Set is a JSON helper used for storing key-value session values.
// Set modifies the marshalled map stored in the session to include the key value pair passed in.
func Set(overseer Overseer, w http.ResponseWriter, r *http.Request, key string, value string) error {
	sessMap := map[string]string{}
	err := GetObj(overseer, w, r, &sessMap)
	// If it's a no session error because a session hasn't been created yet
	// then we can skip this return statement and create a fresh map
	if err != nil && !IsNoSessionError(err) {
		return err
	}

	sessMap[key] = value
	ret, err := json.Marshal(sessMap)
	if err != nil {
		return err
	}

	return overseer.Set(w, r, string(ret))
}

// Get is a JSON helper used for retrieving key-value session values.
// Get returns the value pointed to by the key of the marshalled map stored in the session.
func Get(overseer Overseer, w http.ResponseWriter, r *http.Request, key string) (string, error) {
	var ret map[string]string
	val, err := overseer.Get(w, r)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal([]byte(val), &ret)
	if err != nil {
		return "", err
	}

	mapVal, ok := ret[key]
	if !ok {
		return "", errNoMapKey{}
	}

	return mapVal, nil
}

// Del is a JSON helper used for deleting keys from a key-value session values store.
// Del is a noop on nonexistent keys, but will error if the session does not exist.
func Del(overseer Overseer, w http.ResponseWriter, r *http.Request, key string) error {
	sessMap := map[string]string{}
	err := GetObj(overseer, w, r, &sessMap)
	if err != nil {
		return err
	}

	delete(sessMap, key)

	ret, err := json.Marshal(sessMap)
	if err != nil {
		return err
	}

	return overseer.Set(w, r, string(ret))
}

// SetObj is a JSON helper used for storing object or variable session values.
// Set stores in the session a marshaled version of the passed in value pointed to by v.
func SetObj(overseer Overseer, w http.ResponseWriter, r *http.Request, v interface{}) error {
	ret, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return overseer.Set(w, r, string(ret))
}

// GetObj is a JSON helper used for retrieving object or variable session values.
// GetObj unmarshals the session value into the value pointed to by v.
func GetObj(overseer Overseer, w http.ResponseWriter, r *http.Request, v interface{}) error {
	val, err := overseer.Get(w, r)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(val), v)
	if err != nil {
		return err
	}

	return nil
}
