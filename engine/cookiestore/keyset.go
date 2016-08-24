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

// Keyset holds the secrets for signing and encrypting/decrypting session cookies.
// It should be instantiated using the NewKeyset and NewUnencryptedKeyset functions
// only.
type Keyset struct {
	hmacKey []byte
	block   cipher.Block
}

// NewKeyset returns a pointer to a Keyset, which contains your secret keys used
// for encrypting/decrypting the session data and signing the session cookie.
//
// The hmacKey parameter is used to create the HMAC hash to sign the session cookie.
// Because cookiestore uses SHA256 as the HMAC hashing algorithm, the recommended
// minimum length hmacKey parameter is at least 32 random bytes. If you're storing
// the key as an encoded string for convenience, the underlying entropy should
// still be 32 bytes (i.e you should use a 64 character hex string or 43 character
// base64 string).
//
// The blockKey parameter is used to encrypt/decrypt the session data. It must be
// 16, 24 or 32 bytes long. The byte length you use will control which AES implementation
// is used. A 16 byte `blockKey` will mean that  AES-128 is used to encrypt the
// session data, 24 bytes means AES-192 will be used, and 32 bytes means that
// AES-256 will be used.
//
// When rotating Keysets, it is essential that Keysets are entirely unique. You
// must not change the the blockKey for a Keyset without also changing the hmacKey.
// Re-using `hmacKey` values will result in some valid cookies not being able to
// be decoded.
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

// NewUnencryptedKeyset returns a pointer to a Keyset which will sign, but not
// encrypt, the session cookie. The cookie will be tamper-proof, but an user or
// attacker will be able to read the session data in the cookie. Using unencrypted
// session cookies is marginally faster.
func NewUnencryptedKeyset(hmacKey []byte) (*Keyset, error) {
	if len(hmacKey) < 32 {
		return nil, errHMACKeyLength
	}

	return &Keyset{hmacKey: hmacKey}, nil
}
