package cookiestore

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strconv"
	"time"

	"golang.org/x/crypto/nacl/secretbox"
)

var (
	errTokenTooLong  = errors.New("cookiestore: encoded token length exceeded 4096 characters")
	errInvalidToken  = errors.New("cookiestore: token is invalid")
	errInvalidExpiry = errors.New("cookiestore: expiry time is invalid")
)

// CookieStore represents the currently configured session store.
type CookieStore struct {
	keys [][32]byte
}

// New returns a new CookieStore instance.
//
// The key parameter should contain the secret you want to use to authenticate and
// encrypt session cookies. This should be exactly 32 bytes long.
//
// Optionally, the variadic oldKeys parameter can be used to provide an arbitrary
// number of old Keys. This should be used to ensure that valid cookies continue
// to work correctly after key rotation.
func New(key []byte, oldKeys ...[]byte) *CookieStore {
	keys := make([][32]byte, 1)
	copy(keys[0][:], key)

	for _, key := range oldKeys {
		var newKey [32]byte
		copy(newKey[:], key)
		keys = append(keys, newKey)
	}

	return &CookieStore{
		keys: keys,
	}
}

// MakeToken creates a signed, optionally encrypted, cookie token for the provided
// session data. The returned token is limited to 4096 characters in length. An
// error will be returned if this is exceeded.
func (c *CookieStore) MakeToken(b []byte, expiry time.Time) (token string, err error) {
	return encodeToken(c.keys[0], b, expiry)
}

// Find returns the session data for given cookie token. It loops through all
// available keys (including old keys) to try to decode the cookie. If
// the cookie could not be decoded, or has expired, the returned exists flag
// will be set to false.
func (c *CookieStore) Find(token string) (b []byte, exists bool, error error) {
	for _, key := range c.keys {
		b, err := decodeToken(key, token)
		switch err {
		case nil:
			return b, true, nil
		case errInvalidToken:
			continue
		default:
			return nil, false, err
		}
	}
	return nil, false, nil
}

// Save is a no-op. The function exists only to ensure that a CookieStore instance
// satisfies the scs.Store interface.
func (c *CookieStore) Save(token string, b []byte, expiry time.Time) error {
	return nil
}

// Delete is a no-op. The function exists only to ensure that a CookieStore instance
// satisfies the scs.Store interface.
func (c *CookieStore) Delete(token string) error {
	return nil
}

func encodeToken(key [32]byte, b []byte, expiry time.Time) (string, error) {
	expiryTimestamp := []byte(strconv.FormatInt(expiry.UnixNano(), 10))
	if len(expiryTimestamp) != 19 {
		return "", errInvalidExpiry
	}

	message := append(expiryTimestamp, b...)

	var nonce [24]byte
	_, err := rand.Read(nonce[:])
	if err != nil {
		return "", err
	}

	box := secretbox.Seal(nonce[:], message, &nonce, &key)

	token := base64.RawURLEncoding.EncodeToString(box)
	if len(token) > 4096 {
		return "", errTokenTooLong
	}

	return token, nil
}

func decodeToken(key [32]byte, token string) ([]byte, error) {
	box, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, errInvalidToken
	}
	if len(box) < 24 {
		return nil, errInvalidToken
	}

	var nonce [24]byte
	copy(nonce[:], box[:24])
	message, ok := secretbox.Open(nil, box[24:], &nonce, &key)
	if !ok {
		return nil, errInvalidToken
	}

	expiryTimestamp, err := strconv.ParseInt(string(message[:19]), 10, 64)
	if err != nil {
		return nil, errInvalidToken
	}
	if expiryTimestamp < time.Now().UnixNano() {
		return nil, errInvalidToken
	}

	return message[19:], nil
}
