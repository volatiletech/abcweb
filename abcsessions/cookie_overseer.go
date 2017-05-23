package abcsessions

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/pkg/errors"
)

// CookieOverseer oversees cookie operations that are encrypted and verified
// but does store all data client side which means it is a possible attack
// vector. Uses GCM to verify and encrypt data.
type CookieOverseer struct {
	options CookieOptions

	secretKey    []byte
	gcmBlockMode cipher.AEAD

	resetExpiryMiddleware
}

// NewCookieOverseer creates an overseer from cookie options and a secret key
// for use in encryption. Panic's on any errors that deal with cryptography.
func NewCookieOverseer(opts CookieOptions, secretKey []byte) *CookieOverseer {
	if len(opts.Name) == 0 {
		panic("cookie name must be provided")
	}

	block, err := aes.NewCipher(secretKey)
	if err != nil {
		panic(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}

	o := &CookieOverseer{
		options:      opts,
		secretKey:    secretKey,
		gcmBlockMode: gcm,
	}

	o.resetExpiryMiddleware.resetter = o

	return o
}

// Get a value from the cookie overseer
func (c *CookieOverseer) Get(w http.ResponseWriter, r *http.Request) (string, error) {
	val, err := c.options.getCookieValue(w, r)
	if err != nil {
		return "", errors.Wrap(err, "unable to get session value from cookie")
	}

	return c.decode(val)
}

// Set a value into the cookie overseer
func (c *CookieOverseer) Set(w http.ResponseWriter, r *http.Request, value string) error {
	ev, err := c.encode(value)
	if err != nil {
		return errors.Wrap(err, "unable to encode session value into cookie")
	}

	w.(cookieWriter).SetCookie(c.options.makeCookie(ev))

	return nil
}

// Del a value from the cookie overseer
func (c *CookieOverseer) Del(w http.ResponseWriter, r *http.Request) error {
	c.options.deleteCookie(w)
	return nil
}

// Regenerate for the cookie overseer will panic because cookie sessions
// do not have session IDs, only values.
func (c *CookieOverseer) Regenerate(w http.ResponseWriter, r *http.Request) error {
	panic("cookie sessions do not use session ids")
}

// SessionID for the cookie overseer will panic because cookie sessions
// do not have session IDs, only values.
func (c *CookieOverseer) SessionID(w http.ResponseWriter, r *http.Request) (string, error) {
	panic("cookie sessions do not use session ids")
}

// ResetExpiry resets the age of the session to time.Now(), so that
// MaxAge calculations are renewed
func (c *CookieOverseer) ResetExpiry(w http.ResponseWriter, r *http.Request) error {
	if c.options.MaxAge == 0 {
		return nil
	}

	val, err := c.options.getCookieValue(w, r)
	if err != nil {
		return errors.Wrap(err, "unable to get session value from cookie")
	}

	w.(cookieWriter).SetCookie(c.options.makeCookie(val))

	return nil
}

// encode into base64'd aes-gcm
func (c *CookieOverseer) encode(plaintext string) (string, error) {
	nonce := make([]byte, c.gcmBlockMode.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", errors.Wrap(err, "failed to encode session cookie value")
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
		return "", errors.New("failed to decode session cookie value")
	}

	// Nonce comes from the first n bytes (n = NonceSize)
	plaintext, err := c.gcmBlockMode.Open(nil,
		ct[:c.gcmBlockMode.NonceSize()],
		ct[c.gcmBlockMode.NonceSize():],
		nil)

	if err != nil {
		return "", errors.Wrap(err, "unable to open gcm block mode")
	}

	return string(plaintext), nil
}
