// TODO: Document that
//  * Keysets must be entirely unique (i.e HMAC keys should not be reused)
//	* The blockKey argument should be the AES key, either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256
// 	* HMAC keys are recommended to be 32 random bytes. If an encoded string is used the underlying entropy should still be 32 random bytes (i.e a 64 character hex string or 43 character base64 string)
package cookiestore

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

var (
	errHMACKeyLength  = errors.New("hmacKey length must be at least 32 bytes")
	errBlockKeyLength = errors.New("blockKey length must be 16, 24 or 32 bytes")
)

type Keyset struct {
	hmacKey []byte
	block   cipher.Block
}

func NewKeyset(hmacKey, blockKey []byte) (*Keyset, error) {
	if len(hmacKey) < 32 {
		return nil, errHMACKeyLength
	}

	switch len(blockKey) {
	case 16, 24, 32:
		break
	default:
		return nil, errBlockKeyLength
	}

	block, err := aes.NewCipher(blockKey)
	if err != nil {
		return nil, err
	}

	return &Keyset{hmacKey, block}, nil
}

func NewUnencryptedKeyset(hmacKey []byte) (*Keyset, error) {
	if len(hmacKey) < 32 {
		return nil, errHMACKeyLength
	}

	return &Keyset{hmacKey: hmacKey}, nil
}
