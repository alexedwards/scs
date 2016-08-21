// TODO: Document that
//	RenegerateToken and Renew are no-ops
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

	"github.com/alexedwards/scs/session"
)

var (
	errTokenTooLong = errors.New("encoded token length exceeded 4096 characters")
	errInvalidToken = errors.New("token is invalid")
)

type CookieStore struct {
	keyset     *Keyset
	oldKeysets []*Keyset
}

func New(keyset *Keyset, oldKeysets ...*Keyset) *CookieStore {
	return &CookieStore{keyset, oldKeysets}
}

func (c *CookieStore) MakeToken(b []byte, expiry time.Time) (string, error) {
	return encodeToken(c.keyset, b, expiry)
}

func (c *CookieStore) Find(token string) ([]byte, bool, error) {
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

func (c *CookieStore) Save(token string, b []byte, expiry time.Time) error {
	return nil
}

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
