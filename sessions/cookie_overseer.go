package sessions

import "crypto/rand"

// CookieOverseer oversees cookie operations that are encrypted and verified
// but does store all data client side which means it is a possible attack
// vector.
type CookieOverseer struct {
	options CookieOptions

	secretKeyAES  []byte
	secretKeyHMAC []byte
}

// NewCookieOverseer creates
func NewCookieOverseer(opts CookieOptions, secretKeyAES, secretKeyHMAC []byte) *CookieOverseer {
	return &CookieOverseer{
		options:       opts,
		secretKeyAES:  secretKeyAES,
		secretKeyHMAC: secretKeyHMAC,
	}
}

// MakeAESKey creates a randomized key securely for use with the AES algorithm
// the size of the key is 32 bytes in order to use AES-256. It must be persisted
// somewhere in order to re-use it across restarts of the sessions.
func MakeAESKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

// MakeHashKey creates a
func MakeHashKey() ([]byte, error) {
	key := make([]byte, 64)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}
