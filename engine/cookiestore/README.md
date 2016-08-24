# cookiestore 
[![godoc](https://godoc.org/github.com/alexedwards/scs/engine/cookiestore?status.png)](https://godoc.org/github.com/alexedwards/scs/engine/cookiestore)

Package cookiestore is a cookie-based storage engine for the [SCS session package](https://godoc.org/github.com/alexedwards/scs/session).

It stores session data in AES-encrypted and SHA256-signed cookies on the client. Key rotation is supported for increased security.

The cookiestore package provides a simple and easy way to implement session functionality, with no external dependencies. 

## Usage

### Installation

Either:

```
$ go get github.com/alexedwards/scs/engine/cookiestore
```

Or (recommended) use use [gvt](https://github.com/FiloSottile/gvt) to vendor the `engine/cookiestore` and `session` sub-packages:

```
$ gvt fetch github.com/alexedwards/scs/engine/cookiestore
$ gvt fetch github.com/alexedwards/scs/session
```

### Example

```go
package main

import (
    "io"
    "log"
    "net/http"

    "github.com/alexedwards/scs/engine/cookiestore"
    "github.com/alexedwards/scs/session"
)

// HMAC authentication key (hexadecimal representation of 32 random bytes)
var hmacKey = []byte("f71dc7e58abab014ddad2652475056f185164d262869c8931b239de52711ba87")

// AES encryption key (hexadecimal representation of 16 random bytes)
var blockKey = []byte("911182cec2f206986c8c82440adb7d17")

func main() {
    // Create a new keyset using your authentication and encryption secret keys.
    keyset, err := cookiestore.NewKeyset(hmacKey, blockKey)
    if err != nil {
        log.Fatal(err)
    }

    // Create a new CookieStore instance using the keyset.
    engine := cookiestore.New(keyset)

    sessionManager := session.Manage(engine)
    http.HandleFunc("/put", putHandler)
    http.HandleFunc("/get", getHandler)
    http.ListenAndServe(":4000", sessionManager(http.DefaultServeMux))
}

func putHandler(w http.ResponseWriter, r *http.Request) {
    err := session.PutString(r, "message", "Hello world!")
    if err != nil {
        http.Error(w, err.Error(), 500)
    }
}

func getHandler(w http.ResponseWriter, r *http.Request) {
    msg, err := session.GetString(r, "message")
    if err != nil {
        http.Error(w, err.Error(), 500)
    }
    io.WriteString(w, msg)
}
```

#### Creating a Keyset

Every CookieStore instance must have a Keyset, which contains your secret keys used for encrypting/decrypting the session data and signing the session cookie.

Keysets are created with the `NewKeyset()` function, which takes an `hmacKey` parameter (used to create the HMAC hash to sign the session cookie) and a `blockKey` parameter (used to encrypt/decrypt the session data). 

```go
var hmacKey = []byte("f71dc7e58abab014ddad2652475056f185164d262869c8931b239de52711ba87")
var blockKey = []byte("911182cec2f206986c8c82440adb7d17")

keyset, err := cookiestore.NewKeyset(hmacKey, blockKey)
if err != nil {
    log.Fatal(err)
}
```

Because cookiestore uses SHA256 as the HMAC hashing algorithm, the [recommended minimum length](https://tools.ietf.org/html/rfc2104) of the `hmacKey` parameter is at least 32 random bytes. If you're storing the key as an encoded string for convenience, the underlying entropy should still be 32 bytes (i.e you should use a 64 character hex string or 43 character base64 string).

The `blockKey` must be 16, 24 or 32 bytes long. The byte length you use will control which AES implementation is used. A 16 byte `blockKey` will mean that  AES-128 is used to encrypt the session data, 24 bytes means AES-192 will be used, and 32 bytes means that AES-256 will be used.

#### Unencrypted session cookies

Session cookies that are signed, but not encrypted, can also be used. The cookies will remain tamper-proof, but an user or attacker will be able to read the session data in the cookie.

Using unencrypted session cookies is marginally faster.

Creating a Keyset with the `NewUnencryptedKeyset()` function will result in unencrypted cookies being used.

```go
var hmacKey = []byte("f71dc7e58abab014ddad2652475056f185164d262869c8931b239de52711ba87")

keyset, err := cookiestore.NewUnencryptedKeyset(hmacKey)
if err != nil {
    log.Fatal(err)
}

engine := cookiestore.New(keyset)
```


#### Key rotation

The cookiestore package supports key rotation for increased security.

An arbitrary number of old Keysets can be provided when creating a new CookieStore instance. For example:

```go
keyset, err := cookiestore.NewKeyset([]byte("f71dc7e58abab014ddad2652475056f185164d262869c8931b239de52711ba87"), []byte("911182cec2f206986c8c82440adb7d17"))
if err != nil {
    log.Fatal(err)
}

oldKeyset, err := cookiestore.NewKeyset([]byte("16bd76c6372363cd9af46f5619cc406776210b6164c48fd1200119d4cfc359e6"), []byte("5f8b7a8efac2a900a0c6be609b2e0241"))
if err != nil {
    log.Fatal(err)
}

veryOldKeyset, err := cookiestore.NewKeyset([]byte("0c03fa487baa82dda09c4f12c7238370c58112a135318a6e3d4a4724a95cd2e0"), []byte("46ee77bfb95a765dfefca83bf53d5914"))
if err != nil {
    log.Fatal(err)
}

engine := cookiestore.New(keyset, oldKeyset, veryOldKeyset)
```

When a session cookie is received from a client, all Keysets (including old Keysets) are looped through to try to decode the cookie.

When rotating Keysets, it is essential that Keysets are entirely unique. You must not change the the `blockKey` for a Keyset without also changing the `hmacKey`. Re-using `hmacKey` values will result in some valid cookies not being able to be decoded.

#### Cookie length

Cookies are limited to 4096 characters in length. Storing large amounts of session data may, when encoded and signed, exceed this length and result in an error. 

This makes cookie-based sessions suitable for applications where the amount of session data is known in advance and small. 

### RegenerateToken function

The cookiestore package is a special case for the `scs/session` package because it stores data on the client only, not the server.

This means that using [`session.RegenerateToken()`](https://godoc.org/github.com/alexedwards/scs/session#RegenerateToken) as a mechanism to prevent session fixation attacks is unnecessary when using cookiestore, because the signed and encrypted cookie 'token' always changes whenever the session data is modified anyway.

The only impact that calling `session.RegenerateToken()` will have is to reset and restart the session lifetime.

## Notes

Full godoc documentation: [https://godoc.org/github.com/alexedwards/scs/engine/cookiestore](https://godoc.org/github.com/alexedwards/scs/engine/cookiestore).