package sessions

import (
	"encoding/json"
	"net/http"
	"time"
)

// Storer provides methods to retrieve, add and delete session keys
// and their corresponding values.
type Storer interface {
	Get(key string) (value string, err error)
	Put(key, value string) error
	Del(key string) error
}

// Overseer of session cookies
type Overseer interface {
	Get(w http.ResponseWriter, r *http.Request) (value string, err error)
	Put(w http.ResponseWriter, r *http.Request, value string) (cr *http.Request, err error)
	Del(w http.ResponseWriter, r *http.Request) (err error)
}

// timer interface is used to mock the test harness for disk and memory storers
type timer interface {
	Stop() bool
	Reset(time.Duration) bool
}

// timerTestHarness allows us to control the timer channels manually in the
// disk and memory storer tests so that we can trigger cleans at will
var timerTestHarness = func(d time.Duration) (timer, <-chan time.Time) {
	t := time.NewTimer(d)
	return t, t.C
}

type noSessionInterface interface {
	NoSession()
}

type errNoSession struct{}

func (errNoSession) NoSession() {}
func (errNoSession) Error() string {
	return "session does not exist"
}

// IsNoSessionError checks an error to see if it means that there was no session
func IsNoSessionError(err error) bool {
	_, ok := err.(noSessionInterface)
	return ok
}

type noMapKeyInterface interface {
	NoMapKey()
}

type errNoMapKey struct{}

func (errNoMapKey) NoMapKey() {}
func (errNoMapKey) Error() string {
	return "session map key does not exist"
}

// IsNoMapKeyError checks an error to see if it means that there was no session map key
func IsNoMapKeyError(err error) bool {
	_, ok := err.(noMapKeyInterface)
	return ok
}

// Put is a JSON helper used for storing key-value session values.
// Put modifies the marshalled map stored in the session to include the key value pair passed in.
func Put(overseer Overseer, w http.ResponseWriter, r *http.Request, key string, value string) (*http.Request, error) {
	sessMap := map[string]string{}
	err := GetObj(overseer, w, r, &sessMap)
	// If it's a no session error because a session hasn't been created yet
	// then we can skip this return statement and create a fresh map
	if err != nil && !IsNoSessionError(err) {
		return nil, err
	}

	sessMap[key] = value
	ret, err := json.Marshal(sessMap)
	if err != nil {
		return nil, err
	}

	return overseer.Put(w, r, string(ret))
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

	_, err = overseer.Put(w, r, string(ret))
	return err
}

// PutObj is a JSON helper used for storing object or variable session values.
// Put stores in the session a marshaled version of the passed in value pointed to by v.
func PutObj(overseer Overseer, w http.ResponseWriter, r *http.Request, v interface{}) (*http.Request, error) {
	ret, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return overseer.Put(w, r, string(ret))
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
