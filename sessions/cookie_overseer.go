package sessions

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
)

// CookieOverseer oversees cookie operations that are encrypted and verified
// but does store all data client side which means it is a possible attack
// vector. Uses GCM to verify and encrypt data.
type CookieOverseer struct {
	options CookieOptions

	secretKey    [32]byte
	gcmBlockMode cipher.AEAD
}

// NewCookieOverseer creates an overseer from cookie options and a secret key
// for use in encryption. Panic's on any errors that deal with cryptography.
func NewCookieOverseer(opts CookieOptions, secretKey [32]byte) *CookieOverseer {
	if len(opts.Name) == 0 {
		panic("cookie name must be provided")
	}

	block, err := aes.NewCipher(secretKey[:])
	if err != nil {
		panic(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	return &CookieOverseer{
		options:      opts,
		secretKey:    secretKey,
		gcmBlockMode: gcm,
	}
}

// MakeSecretKey creates a randomized key securely for use with the AES-GCM
// algorithm the size of the key is 32 bytes in order to use AES-256. It must
// be persisted somewhere in order to re-use it across restarts of the app.
func MakeSecretKey() ([32]byte, error) {
	var key [32]byte
	if _, err := rand.Read(key[:]); err != nil {
		return key, err
	}
	return key, nil
}

// Get a value from the cookie overseer
func (c *CookieOverseer) Get(w http.ResponseWriter, r *http.Request) (string, error) {
	val, err := c.options.getCookieValue(r)
	if err != nil {
		return "", err
	}

	return c.decode(val)
}

// Set a value into the cookie overseer
func (c *CookieOverseer) Set(w http.ResponseWriter, r *http.Request, value string) (*http.Request, error) {
	ev, err := c.encode(value)
	if err != nil {
		return nil, err
	}

	http.SetCookie(w, c.options.makeCookie(ev))

	// Store the cookie value in context so it can be retrieved from context
	// in subsequent Set calls.
	r = r.WithContext(context.WithValue(r.Context(), c.options.Name, ev))

	return r, nil
}

// Del a value from the cookie overseer
func (c *CookieOverseer) Del(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	cookie := &http.Cookie{
		// If the browser refuses to delete it, set value to "" so subsequent
		// requests replace it when it does not point to a valid session id.
		Value:    "",
		Name:     c.options.Name,
		MaxAge:   -1,
		Expires:  time.Now().UTC().AddDate(-1, 0, 0),
		HttpOnly: c.options.HTTPOnly,
		Secure:   c.options.Secure,
	}

	http.SetCookie(w, cookie)

	// Reset the context so it doesn't re-use the old deleted session value
	r = r.WithContext(context.WithValue(r.Context(), c.options.Name, ""))

	return r, nil
}

// Regenerate for the cookie overseer will panic because cookie sessions
// do not have session IDs, only values.
func (c *CookieOverseer) Regenerate(w http.ResponseWriter, r *http.Request) (*http.Request, error) {
	panic("cookie sessions do not use session ids")
}

// SessionID for the cookie overseer will panic because cookie sessions
// do not have session IDs, only values.
func (c *CookieOverseer) SessionID(r *http.Request) (string, error) {
	panic("cookie sessions do not use session ids")
}

// ResetExpiry resets the age of the session to time.Now(), so that
// MaxAge calculations are renewed
func (c *CookieOverseer) ResetExpiry(w http.ResponseWriter, r *http.Request) error {
	if c.options.MaxAge == 0 {
		return nil
	}

	val, err := c.options.getCookieValue(r)
	if err != nil {
		return err
	}

	spew.Dump(w)
	cookie := c.options.makeCookie(val)
	fmt.Printf("\n\n%#v\n\n", cookie)
	http.SetCookie(w, cookie)

	spew.Dump(w)

	return nil
}

// encode into base64'd aes-gcm
func (c *CookieOverseer) encode(plaintext string) (string, error) {
	nonce := make([]byte, c.gcmBlockMode.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", errors.Wrap(err, "failed to encode session cookie")
	}

	// Append ciphertext to the end of nonce so we have the nonce for decrypt
	ciphertext := c.gcmBlockMode.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decode base64'd aes-gcm
func (c *CookieOverseer) decode(ciphertext string) (string, error) {
	ct, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", nil
	}

	if len(ct) <= c.gcmBlockMode.NonceSize() {
		return "", errors.New("failed to decode in cookie overseer")
	}

	// Nonce comes from the first n bytes (n = NonceSize)
	plaintext, err := c.gcmBlockMode.Open(nil,
		ct[:c.gcmBlockMode.NonceSize()],
		ct[c.gcmBlockMode.NonceSize():],
		nil)

	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
