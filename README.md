# SCS: A HTTP Session Manager
[![godoc](https://godoc.org/github.com/alexedwards/scs?status.png)](https://godoc.org/github.com/alexedwards/scs) [![go report card](https://goreportcard.com/badge/github.com/alexedwards/scs)](https://goreportcard.com/report/github.com/alexedwards/scs)

SCS is a fast and lightweight HTTP session manager for Go. It features:

* Built-in PostgreSQL, MySQL, Redis, Memcached, encrypted cookie and in-memory storage engines. Custom storage engines are also supported.
* Supports OWASP good-practices, including absolute and idle session timeouts and easy regeneration of session tokens.
* Fast and very memory-efficient performance.
* Type-safe and sensible API for managing session data. Safe for concurrent use.
* Automatic saving of session data.

**Recent changes:** Release v1.0.0 made breaking changes to the package layout and API. If you need the old version please vendor [release v0.1.1](https://github.com/alexedwards/scs/releases/tag/v0.1.1).

## Installation &amp; Usage

Install with `go get`:

```sh
$ go get github.com/alexedwards/scs
```

### Basic use

```go
package main

import (
    "io"
    "net/http"

    "github.com/alexedwards/scs"
)

// Initialize a new encrypted-cookie based session manager and store it in a global
// variable. In a real application, you might inject the session manager as a
// dependency to your handlers instead. The parameter to the NewCookieManager()
// function is a 32 character long random key, which is used to encrypt and
// authenticate the session cookies.
var sessionManager = scs.NewCookieManager("u46IpCV9y5Vlur8YvODJEhgOY8m9JVE4")

func main() {
    // Set up your HTTP handlers in the normal way.
    mux := http.NewServeMux()
    mux.HandleFunc("/put", putHandler)
    mux.HandleFunc("/get", getHandler)

    // Wrap your handlers with the session manager middleware.
    http.ListenAndServe(":4000", sessionManager.Use(mux))
}

func putHandler(w http.ResponseWriter, r *http.Request) {
    // Load the session data for the current request. Any errors are deferred
    // until you actually use the session data.
    session := sessionManager.Load(r)

    // Use the PutString() method to add a new key and associated string value
    // to the session data. Methods for many other common data types are also
    // provided. The session data is automatically saved.
    err := session.PutString(w, "message", "Hello world!")
    if err != nil {
        http.Error(w, err.Error(), 500)
    }
}

func getHandler(w http.ResponseWriter, r *http.Request) {
    // Load the session data for the current request.
    session := sessionManager.Load(r)

    // Use the GetString() helper to retrieve the string value for the "message"
    // key from the session data. The zero value for a string is returned if the
    // key does not exist.
    message, err := session.GetString("message")
    if err != nil {
        http.Error(w, err.Error(), 500)
    }

    io.WriteString(w, message)
}
```

SCS provides a wide range of functions for working with session data.

* `Put…` and `Get…` methods for storing and retrieving a variety of common data types and custom objects.
* `Pop…` methods for one-time retrieval of common data types (and custom objects) from the session data.
* `Keys` returns an alphabetically-sorted slice of all keys in the session data.
* `Exists` returns whether a specific key exists in the session data.
* `Remove` removes an individual key and value from the session data.
* `Clear` removes all data for the current session.
* `RenewToken` creates a new session token. This should be used before privilege changes to help avoid session fixation.
* `Destroy` deletes the current session and instructs the browser to delete the session cookie.

A full list of available functions can be found in [the GoDoc](https://godoc.org/github.com/alexedwards/scs/#pkg-index).

### Customizing the session manager

The session manager can be configured to customize its behavior. For example:

```go
sessionManager = scs.NewCookieManager("u46IpCV9y5Vlur8YvODJEhgOY8m9JVE4")
sessionManager.Lifetime(time.Hour) // Set the maximum session lifetime to 1 hour.
sessionManager.Persist(true) // Persist the session after a user has closed their browser.
sessionManager.Secure(true) // Set the Secure flag on the session cookie.
```

A full list of available settings can be found in [the GoDoc](https://godoc.org/github.com/alexedwards/scs/#pkg-index).

### Using a different session store

The above examples use encrypted cookies to store session data, but SCS also supports a range of server-side stores.

| Package                                                                               |                                                                                   |
|:------------------------------------------------------------------------------------- |-----------------------------------------------------------------------------------|
| [stores/boltstore](https://godoc.org/github.com/alexedwards/scs/stores/boltstore)     | BoltDB-based session store                                                        |
| [stores/buntstore](https://godoc.org/github.com/alexedwards/scs/stores/buntstore)     | BuntDB based session store                                                        |
| [stores/cookiestore](https://godoc.org/github.com/alexedwards/scs/stores/cookiestore) | Encrypted-cookie session store                                                    |
| [stores/dynamostore](https://godoc.org/github.com/alexedwards/scs/stores/dynamostore) | DynamoDB-based session store                                                      |
| [stores/memstore](https://godoc.org/github.com/alexedwards/scs/stores/memstore)       | In-memory session store                                                           |
| [stores/mysqlstore](https://godoc.org/github.com/alexedwards/scs/stores/mysqlstore)   | MySQL-based session store                                                         |
| [stores/pgstore](https://godoc.org/github.com/alexedwards/scs/stores/pgstore)         | PostgreSQL-based storage eninge                                                   |
| [stores/qlstore](https://godoc.org/github.com/alexedwards/scs/stores/qlstore)         | QL-based session store                                                            |
| [stores/redisstore](https://godoc.org/github.com/alexedwards/scs/stores/redisstore)   | Redis-based session store                                                         |
| [stores/memcached](https://godoc.org/github.com/alexedwards/scs/stores/memcachedstore)| Memcached-based session store                                                     |

### Compatibility

SCS is designed to be compatible with Go's `net/http` package and the `http.Handler` interface.

If you're using the [Echo](https://echo.labstack.com/) framework, the [official session middleware](https://echo.labstack.com/middleware/session) for Echo is likely to be a better fit for your application.

### Examples

* [RequireLogin middleware](https://gist.github.com/alexedwards/6eac2f19b9b5c064ca90f756c32f94cc)