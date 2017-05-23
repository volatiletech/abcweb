package abcsessions

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// session holds the session value and the flash messages key/value mapping
type session struct {
	// value is the session value stored as a json encoded string
	Value *json.RawMessage
	// flash is the key/value storage for flash messages. Depending on whether
	// you're calling Get/SetFlash or Get/SetFlashObj it will either store
	// a json string or a json object.
	Flash map[string]*json.RawMessage
}

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
	// MiddlewareWithReset converts the writer to a sessionsResponseWriter
	// and will reset the users session expiry on every request
	MiddlewareWithReset(next http.Handler) http.Handler
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
	if ok {
		return ok
	}

	_, ok = errors.Cause(err).(noSessionInterface)
	return ok
}

// IsNoMapKeyError checks an error to see if it means that there was
// no session map key
func IsNoMapKeyError(err error) bool {
	_, ok := err.(noMapKeyInterface)
	if ok {
		return ok
	}

	_, ok = errors.Cause(err).(noMapKeyInterface)
	return ok
}

// timerTestHarness allows us to control the timer channels manually in the
// disk and memory storer tests so that we can trigger cleans at will
var timerTestHarness = func(d time.Duration) (timer, <-chan time.Time) {
	t := time.NewTimer(d)
	return t, t.C
}

// validKey returns true if the session key is a valid UUIDv4 format:
// 8chars-4chars-4chars-4chars-12chars (chars are a-f 0-9)
// Example: a668b3bb-0cf1-4627-8cd4-7f62d09ebad6
func validKey(key string) bool {
	// UUIDv4's are 36 chars (16 bytes not including dashes)
	if len(key) != 36 {
		return false
	}

	// 0 indexed dash positions
	dashPos := []int{8, 13, 18, 23}
	for i := 0; i < len(key); i++ {
		atDashPos := false
		for _, pos := range dashPos {
			if i == pos {
				atDashPos = true
				break
			}
		}

		if atDashPos == true {
			if key[i] != '-' {
				return false
			}
			// continue the loop if dash is found
			continue
		}

		// if not a dash, make sure char is a-f or 0-9
		// 48 == '0', 57 == '9', 97 == 'a', 102 == 'f'
		if key[i] < 48 || (key[i] > 57 && key[i] < 97) || key[i] > 102 {
			return false
		}
	}

	return true
}

// Set is a JSON helper used for storing key-value session values.
// Set modifies the marshalled map stored in the session to include the key value pair passed in.
func Set(overseer Overseer, w http.ResponseWriter, r *http.Request, key string, value string) error {
	var sess session
	sessMap := make(map[string]string)

	val, err := overseer.Get(w, r)
	if err != nil && !IsNoSessionError(err) {
		return errors.Wrap(err, "unable to get session")
	}

	if !IsNoSessionError(err) {
		err = json.Unmarshal([]byte(val), &sess)
		if err != nil {
			return errors.Wrap(err, "unable to unmarshal session object")
		}

		if sess.Value != nil {
			err = json.Unmarshal(*sess.Value, &sessMap)
			if err != nil {
				return errors.Wrap(err, "unable to unmarshal session map value")
			}
		}
	}

	sessMap[key] = value

	mv, err := json.Marshal(sessMap)
	if err != nil {
		return errors.Wrap(err, "unable to marshal session map value")
	}
	sess.Value = (*json.RawMessage)(&mv)

	ret, err := json.Marshal(sess)
	if err != nil {
		return errors.Wrap(err, "unable to marshal session object")
	}

	return overseer.Set(w, r, string(ret))
}

// Get is a JSON helper used for retrieving key-value session values.
// Get returns the value pointed to by the key of the marshalled map stored in the session.
func Get(overseer Overseer, w http.ResponseWriter, r *http.Request, key string) (string, error) {
	val, err := overseer.Get(w, r)
	if err != nil {
		return "", errors.Wrap(err, "unable to get session")
	}

	var sess session
	err = json.Unmarshal([]byte(val), &sess)
	if err != nil {
		return "", errors.Wrap(err, "unable to unmarshal session object")
	}

	var sessMap map[string]string
	err = json.Unmarshal(*sess.Value, &sessMap)
	if err != nil {
		return "", errors.Wrap(err, "unable to unmarshal session map value")
	}

	mapVal, ok := sessMap[key]
	if !ok {
		return "", errNoMapKey{}
	}

	return mapVal, nil
}

// Del is a JSON helper used for deleting keys from a key-value session values store.
// Del is a noop on nonexistent keys, but will error if the session does not exist.
func Del(overseer Overseer, w http.ResponseWriter, r *http.Request, key string) error {
	val, err := overseer.Get(w, r)
	if err != nil {
		return errors.Wrap(err, "unable to get session")
	}

	var sess session
	err = json.Unmarshal([]byte(val), &sess)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal session object")
	}

	var sessMap map[string]string
	err = json.Unmarshal(*sess.Value, &sessMap)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal session map value")
	}

	delete(sessMap, key)

	mv, err := json.Marshal(sessMap)
	if err != nil {
		return errors.Wrap(err, "unable to marshal session map value")
	}
	sess.Value = (*json.RawMessage)(&mv)

	ret, err := json.Marshal(sess)
	if err != nil {
		return errors.Wrap(err, "unable to marshal session object")
	}

	return overseer.Set(w, r, string(ret))
}

// SetObj is a JSON helper used for storing object or variable session values.
// Set stores in the session a marshaled version of the passed in value pointed to by value.
func SetObj(overseer Overseer, w http.ResponseWriter, r *http.Request, value interface{}) error {
	val, err := overseer.Get(w, r)
	// If it's a no session error because a session hasn't been created yet
	// then we can skip this return statement and create a fresh map
	if err != nil && !IsNoSessionError(err) {
		return errors.Wrap(err, "unable to get session")
	}

	var sess session

	// If there's an existing session then unmarshal it so we can copy over
	// the flash messages to the new marshalled session
	if !IsNoSessionError(err) {
		// json unmarshal the outter session struct
		err = json.Unmarshal([]byte(val), &sess)
		if err != nil {
			return errors.Wrap(err, "unable to unmarshal session object")
		}

	}

	mv, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "unable to marshal value")
	}
	sess.Value = (*json.RawMessage)(&mv)

	ret, err := json.Marshal(sess)
	if err != nil {
		return errors.Wrap(err, "unable to marshal session object")
	}

	return overseer.Set(w, r, string(ret))
}

// GetObj is a JSON helper used for retrieving object or variable session values.
// GetObj unmarshals the session value into the pointer pointed to by pointer.
func GetObj(overseer Overseer, w http.ResponseWriter, r *http.Request, pointer interface{}) error {
	val, err := overseer.Get(w, r)
	if err != nil {
		return errors.Wrap(err, "unable to get session")
	}

	var sess session
	// json unmarshal the outter session struct
	err = json.Unmarshal([]byte(val), &sess)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal session object")
	}

	// json unmarshal the RawMessage value into the users pointer
	err = json.Unmarshal(*sess.Value, pointer)
	return errors.Wrap(err, "unable to unmarshal session value into pointer")
}

// AddFlash adds a flash message to the session that will be deleted when it is retrieved with GetFlash
func AddFlash(overseer Overseer, w http.ResponseWriter, r *http.Request, key string, value string) error {
	var sess session

	val, err := overseer.Get(w, r)
	if err != nil && !IsNoSessionError(err) {
		return errors.Wrap(err, "unable to get session")
	} else if !IsNoSessionError(err) {
		err = json.Unmarshal([]byte(val), &sess)
		if err != nil {
			return errors.Wrap(err, "unable to unmarshal session object")
		}
	}

	if sess.Flash == nil {
		sess.Flash = make(map[string]*json.RawMessage)
	}

	mv, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "unable to marshal session value")
	}
	sess.Flash[key] = (*json.RawMessage)(&mv)

	ret, err := json.Marshal(sess)
	if err != nil {
		return errors.Wrap(err, "unable to marshal session object")
	}

	return overseer.Set(w, r, string(ret))
}

// GetFlash retrieves a flash message from the session then deletes it
func GetFlash(overseer Overseer, w http.ResponseWriter, r *http.Request, key string) (string, error) {
	var sess session

	val, err := overseer.Get(w, r)
	if err != nil {
		return "", errors.Wrap(err, "unable to get session")
	}

	err = json.Unmarshal([]byte(val), &sess)
	if err != nil {
		return "", errors.Wrap(err, "unable to unmarshal session object")
	}

	fv, ok := sess.Flash[key]
	if !ok {
		return "", errNoMapKey{}
	}

	var ret string
	err = json.Unmarshal(*fv, &ret)
	if err != nil {
		return ret, errors.Wrap(err, "unable to unmarshal flash value")
	}

	delete(sess.Flash, key)

	mv, err := json.Marshal(sess)
	if err != nil {
		return ret, errors.Wrap(err, "unable to marshal session object")
	}

	err = overseer.Set(w, r, string(mv))
	return ret, errors.Wrap(err, "unable to set flash session object")
}

// AddFlashObj adds a flash message to the session that will be deleted when it is retrieved with GetFlash
func AddFlashObj(overseer Overseer, w http.ResponseWriter, r *http.Request, key string, value interface{}) error {
	mv, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "unable to marshal flash value")
	}

	return AddFlash(overseer, w, r, key, string(mv))
}

// GetFlashObj unmarshals a flash message from the session into the users pointer
// then deletes it from the session.
func GetFlashObj(overseer Overseer, w http.ResponseWriter, r *http.Request, key string, pointer interface{}) error {
	val, err := GetFlash(overseer, w, r, key)
	if err != nil {
		return errors.Wrap(err, "unable to get flash object")
	}

	return json.Unmarshal([]byte(val), pointer)
}
