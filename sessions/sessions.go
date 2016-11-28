package sessions

import "net/http"

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

type noSessionInterface interface {
	NoSession()
}

type errNoSession struct{}

func (errNoSession) NoSession() {}
func (errNoSession) Error() string {
	return "no session found"
}

// IsNoSessionError checks an error to see if it means that there was no session
func IsNoSessionError(err error) bool {
	_, ok := err.(noSessionInterface)
	return ok
}

// Get is a JSON helper used for retrieving key-value session values.
// Get returns the unmarshaled session value as a map[string]string.
func Get(overseer Overseer, w http.ResponseWriter, r *http.Request) (map[string]string, error) {

}

// Put is a JSON helper used for storing key-value session values.
// Put stores in the session a marshaled version of the passed in map[string]string.
func Put(overseer Overseer, w http.ResponseWriter, r *http.Request, value map[string]string) (*http.Request, error) {

}

// GetObj is a JSON helper used for retrieving object or variable session values.
// GetObj unmarshals the session value into the value pointed to by v.
func GetObj(overseer Overseer, w http.ResponseWriter, r *http.Request, v interface{}) error {

}

// PutObj is a JSON helper used for storing object or variable session values.
// Put stores in the session a marshaled version of the passed in value pointed to by v.
func PutObj(overseer Overseer, w http.ResponseWriter, r *http.Request, v interface{}) (*http.Request, error) {

}
