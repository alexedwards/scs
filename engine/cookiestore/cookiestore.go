// Package cookiestore is a cookie-based storage engine for the SCS session package.
//
// It stores session data in AES-encrypted and SHA256-signed cookies on the client.
// It also supports key rotation for increased security.
//
//	// HMAC authentication key (hexadecimal representation of 32 random bytes)
//	var hmacKey = []byte("f71dc7e58abab014ddad2652475056f185164d262869c8931b239de52711ba87")
//	// AES encryption key (hexadecimal representation of 16 random bytes)
//	var blockKey = []byte("911182cec2f206986c8c82440adb7d17")
//
//	func main() {
//	    // Create a new keyset using your authentication and encryption secret keys.
//	    keyset, err := cookiestore.NewKeyset(hmacKey, blockKey)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // Create a new CookieStore instance using the keyset.
//	    engine := cookiestore.New(keyset)
//
//	    sessionManager := session.Manage(engine)
//	    http.ListenAndServe(":4000", sessionManager(http.DefaultServeMux))
//	}
//
// The cookiestore package is a special case for the scs/session package because
// it stores data on the client only, not the server. This means that using the
// session.RegenerateToken() function as a mechanism to prevent session fixation
// attacks is unnecessary when using cookiestore, because the signed and encrypted
// cookie 'token' always changes whenever the session data is modified anyway.
// The only impact of calling session.RegenerateToken() is to reset and restart
// the session lifetime.
package cookiestore

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/databrary/scs/session"
)

var (
	errTokenTooLong = errors.New("encoded token length exceeded 4096 characters")
	errInvalidToken = errors.New("token is invalid")
)

// CookieStore represents the currently configured session storage engine.
type CookieStore struct {
	keyset     *Keyset
	oldKeysets []*Keyset
}

// New returns a new CookieStore instance.
//
// The keyset parameter should contain the Keyset you want to use to sign and
// encrypt session cookies.
//
// Optionally, the variadic oldKeyset parameter can be used to provide an arbitrary
// number of old Keysets. This should be used to ensure that valid cookies continue
// to work correctly after key rotation.
func New(keyset *Keyset, oldKeysets ...*Keyset) *CookieStore {
	return &CookieStore{keyset, oldKeysets}
}

// MakeToken creates a signed, optionally encrypted, cookie token for the provided
// session data. The returned token is limited to 4096 characters in length. An
// error will be returned if this is exceeded.
func (c *CookieStore) MakeToken(b []byte, expiry time.Time) (token string, err error) {
	return encodeToken(c.keyset, b, expiry)
}

// Find returns the session data for given cookie token. It loops through all
// available Keysets (including old Keysets) to try to decode the cookie. If
// the cookie could not be decoded, or has expired, the returned exists flag
// will be set to false.
func (c *CookieStore) Find(token string) (b []byte, exists bool, error error) {
	keysets := append([]*Keyset{c.keyset}, c.oldKeysets...)
	for _, keyset := range keysets {
		b, err := decodeToken(keyset, token)
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
// satisfies the session.Engine interface.
func (c *CookieStore) Save(token string, b []byte, expiry time.Time) error {
	return nil
}

// Delete is a no-op. The function exists only to ensure that a CookieStore instance
// satisfies the session.Engine interface.
func (c *CookieStore) Delete(token string) error {
	return nil
}

func encodeToken(keyset *Keyset, b []byte, expiry time.Time) (string, error) {
	timestamp := []byte(fmt.Sprint(expiry.UnixNano()))

	if keyset.block != nil {
		var err error
		b, err = encrypt(b, keyset.block)
		if err != nil {
			return "", err
		}
	}

	payload := make([]byte, base64.RawURLEncoding.EncodedLen(len(b)))
	base64.RawURLEncoding.Encode(payload, b)

	signature := createSignature(keyset.hmacKey, []byte(session.CookieName), payload, timestamp)

	token := string(bytes.Join([][]byte{payload, timestamp, signature}, []byte("|")))
	if len(token) > 4096 {
		return "", errTokenTooLong
	}
	return token, nil
}

func decodeToken(keyset *Keyset, token string) ([]byte, error) {
	parts := bytes.Split([]byte(token), []byte("|"))
	if len(parts) != 3 {
		return nil, errInvalidToken
	}

	payload := parts[0]
	timestamp := parts[1]
	signature := parts[2]

	newSignature := createSignature(keyset.hmacKey, []byte(session.CookieName), payload, timestamp)
	if hmac.Equal(signature, newSignature) == false {
		return nil, errInvalidToken
	}

	expiry, err := strconv.ParseInt(string(timestamp), 10, 64)
	if err != nil {
		return nil, errInvalidToken
	}
	if expiry < time.Now().UnixNano() {
		return nil, errInvalidToken
	}

	b := make([]byte, base64.RawURLEncoding.DecodedLen(len(payload)))
	base64.RawURLEncoding.Decode(b, payload)

	if keyset.block != nil {
		b = decrypt(b, keyset.block)
	}

	return b, nil
}

func encrypt(plaintext []byte, block cipher.Block) ([]byte, error) {
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	iv := ciphertext[:aes.BlockSize]
	_, err := io.ReadFull(rand.Reader, iv)
	if err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	return ciphertext, nil
}

func decrypt(ciphertext []byte, block cipher.Block) []byte {
	plaintext := make([]byte, len(ciphertext)-aes.BlockSize)
	iv := ciphertext[:aes.BlockSize]

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, ciphertext[aes.BlockSize:])
	return plaintext
}

func createSignature(hmacKey []byte, parts ...[]byte) []byte {
	mac := hmac.New(sha256.New, hmacKey)
	for _, x := range parts {
		mac.Write(x)
	}
	sum := mac.Sum(nil)

	signature := make([]byte, base64.RawURLEncoding.EncodedLen(len(sum)))
	base64.RawURLEncoding.Encode(signature, sum)
	return signature
}
