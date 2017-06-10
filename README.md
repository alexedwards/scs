# SCS
[![godoc](https://godoc.org/github.com/alexedwards/scs?status.png)](https://godoc.org/github.com/alexedwards/scs) [![go report card](https://goreportcard.com/badge/github.com/alexedwards/scs)](https://goreportcard.com/report/github.com/alexedwards/scs) 



Session management for Go 1.7+

## Features

* Automatic loading and saving of session data via middleware.
* Fast and very memory-efficient performance. See [the benchmarks](#benchmarks).
* Choice of PostgreSQL, MySQL, Redis, encrypted cookie and in-memory storage engines. Custom storage engines are also supported.
* Type-safe and sensible API. Designed to be safe for concurrent use.
* Supports OWASP good-practices, including absolute and idle session timeouts and easy regeneration of session tokens.

## Installation

SCS is broken up into small single-purpose packages for ease of use. You should install the `session` package and your choice of storage engine from the following table:

| Package                                                                               |                                                                                   |
|:------------------------------------------------------------------------------------- |-----------------------------------------------------------------------------------|
| [session](https://godoc.org/github.com/alexedwards/scs/session)                       | Provides session management middleware and helpers for manipulating session data  |
| [engine/memstore](https://github.com/alexedwards/scs/tree/master/engine/memstore)       | In-memory storage engine                                                          |
| [engine/cookiestore](https://github.com/alexedwards/scs/tree/master/engine/cookiestore) | Encrypted-cookie based storage engine                                             |
| [engine/pgstore](https://github.com/alexedwards/scs/tree/master/engine/pgstore)         | PostgreSQL based storage eninge                                                   |
| [engine/mysqlstore](https://github.com/alexedwards/scs/tree/master/engine/mysqlstore)   | MySQL based storage engine                                                        |
| [engine/redisstore](https://github.com/alexedwards/scs/tree/master/engine/redisstore)   | Redis based storage engine                                                        |
| [engine/boltstore](https://github.com/alexedwards/scs/tree/master/engine/boltstore)     | BoltDB based storage engine                                                       |
| [engine/buntstore](https://github.com/alexedwards/scs/tree/master/engine/buntstore)     | BuntDB based storage engine                                                       |

For example:

```sh
$ go get github.com/alexedwards/scs/session
$ go get github.com/alexedwards/scs/engine/memstore
```

Or (recommended) use use [gvt](https://github.com/FiloSottile/gvt) to vendor the packages you need. For example:

```
$ gvt fetch github.com/alexedwards/scs/session
$ gvt fetch github.com/alexedwards/scs/engine/memstore
```

## Examples

* [Basic use](#basic-use)
* [Setting options](#setting-options)
* [Storing data](#storing-data)
* [Flash data](#flash-data)
* [Preventing session fixation](#preventing-session-fixation)
* [Destroying data and sessions](#destroying-data-and-sessions)

### Basic use

Working with SCS is straightforward: use the `session.Manage` function to initialise a new session management middleware, then wrap your handlers or router with it.

```go
package main

import (
    "io"
    "net/http"

    "github.com/alexedwards/scs/engine/memstore"
    "github.com/alexedwards/scs/session"
)

func main() {
    // Initialise a new storage engine. Here we use the memstore package, but the approach  
    // is the same no matter which back-end store you choose.
    engine := memstore.New(0)

    // Initialise the session manager middleware, passing in the storage engine as
    // the first parameter. This middleware will automatically handle loading and 
    // saving of session data for you.
    sessionManager := session.Manage(engine)

    // Set up your HTTP handlers in the normal way.
    mux := http.NewServeMux()
    mux.HandleFunc("/put", putHandler)
    mux.HandleFunc("/get", getHandler)

    // Wrap your handlers with the session manager middleware.
    http.ListenAndServe(":4000", sessionManager(mux))
}

func putHandler(w http.ResponseWriter, r *http.Request) {
    // Use the PutString helper to store a new key and associated string value in
    // the session data. Helpers are also available for many other data types.
    err := session.PutString(r, "message", "Hello from a session!")
    if err != nil {
        http.Error(w, err.Error(), 500)
    }
}

func getHandler(w http.ResponseWriter, r *http.Request) {
    // Use the GetString helper to retreive the string value associated with a key. 
    // The zero value is returned if the key does not exist.
    msg, err := session.GetString(r, "message")
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    io.WriteString(w, msg)
}
```

### Setting options

The `session.Manage` function accepts a range of [functional options](https://commandcenter.blogspot.co.at/2014/01/self-referential-functions-and-design.html). You can specify any mixture of options, or none at all if you're happy with the defaults.

You can control how and when a session expires:

```go
sessionManager := session.Manage(engine,
    // IdleTimeout sets the maximum length of time a session can be inactive
    // before it expires. By default IdleTimeout is not set (i.e. there is
    // no inactivity timeout).
    session.IdleTimeout(30*time.Minute),

    // Lifetime sets the maximum length of time that a session is valid for
    // before it expires. This is an 'absolute expiry' and is set when the
    // session is first created. The default value is 24 hours.
    session.Lifetime(3*time.Hour),

    // Persist sets whether the session cookie should be persistent or not
    // (i.e. whether it should be retained after a user closes their browser).
    // The default value is false.
    session.Persist(true),
)
```

You can control how the session cookie behaves:

```go
sessionManager := session.Manage(engine,
    session.Domain("example.org"),  // Domain is not set by default.
    session.HttpOnly(false),        // HttpOnly attribute is true by default.
    session.Path("/account"),       // Path is set to "/" by default.
    session.Secure(true),           // Secure attribute is false by default.
)
```

And also set a custom error handler:

```go
sessionManager := session.Manage(engine,
    // ErrorFunc allows you to control behavior when an error is encountered
    // loading or saving a session. The default behavior is for a HTTP 500
    // status code to be written to the ResponseWriter along with the plain-text
    // error string.
    session.ErrorFunc(ServerError),
)
…

func ServerError(w http.ResponseWriter, r *http.Request, err error) {
    log.Println(err.Error())
    http.Error(w, "Sorry, the application encountered an error", 500)
}
```

### Storing data

SCS comes with built-in functions for storing and retreiving various types of data:

* `PutBool`, `GetBool`, `PopBool` &ndash; for use with `bool` types
* `PutBytes`, `GetBytes`, `PopBytes` &ndash; for use with byte slice `[]byte` types
* `PutFloat`, `GetFloat`, `PopFloat` &ndash; for use with `float64` types
* `PutInt`, `GetInt`, `PopInt` &ndash; for use with `int` types
* `PutInt64`, `GetInt64`, `PopInt64` &ndash; for use with `int64` types
* `PutString`, `GetString`, `PopString` &ndash; for use with `string` types
* `PutTime`, `GetTime`, `PopTime` &ndash; for use with `time.Time` types

* `Keys` &ndash; returns a alphabetically-sorted slice of all key names.


#### Custom types

Custom types can be stored and retreived using the `PutObject` and `GetObject` helpers.

Behind the scenes SCS uses gob encoding to store custom data types. For this to work properly:

* Your custom type must first be [registered](https://golang.org/pkg/encoding/gob/#Register) with the `encoding/gob` package.
* The fields of your custom types must be exported.

The `GetObject` function is computationally expensive, compared with the other built-in getters. Use it sparingly if performance is a major concern.

```go
package main

import (
    "encoding/gob"
    "fmt"
    "net/http"

    "github.com/alexedwards/scs/engine/memstore"
    "github.com/alexedwards/scs/session"
)

// Note that the fields on the custom type are all exported.
type User struct {
    Name  string
    Email string
}

func main() {
    // Register the type with the encoding/gob package.
    gob.Register(User{})

    engine := memstore.New(0)
    sessionManager := session.Manage(engine)

    mux := http.NewServeMux()
    mux.HandleFunc("/put", putHandler)
    mux.HandleFunc("/get", getHandler)
    http.ListenAndServe(":4000", sessionManager(mux))
}

func putHandler(w http.ResponseWriter, r *http.Request) {
    // Initialise a pointer to a new custom object.
    user := &User{"Alice", "alice@example.com"}

    // Store the custom object in the session data. Important: you should pass in 
    // a pointer to your object, not the value.
    err := session.PutObject(r, "user", user)
    if err != nil {
        http.Error(w, err.Error(), 500)
    }
}

func getHandler(w http.ResponseWriter, r *http.Request) {
    // Initialise a pointer to a new, empty, custom object.
    user := &User{}

    // Read the custom object data from the session into the pointer.
    err := session.GetObject(r, "user", user)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    fmt.Fprintf(w, "Name: %s, Email: %s", user.Name, user.Email)
}
```

### Flash data

The `PopString` function (and similar helpers for other data types) provide one-time 'read and remove' operations on session data. This is useful for implementing flash-message style functions, such as displaying a one-time notification message after processing a form.

```go
func putHandler(w http.ResponseWriter, r *http.Request) {
    // Use the PutString helper to add the flash data as normal.
    err := session.PutString(r, "flashMessage", "This will be a one-time message!")
    if err != nil {
        http.Error(w, err.Error(), 500)
    }
}

func popHandler(w http.ResponseWriter, r *http.Request) {
    // Use the PopString helper to retrieve the string and delete it from the
    // session.
    msg, err := session.PopString(r, "flashMessage")
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
    io.WriteString(w, msg)
}
```

### Preventing session fixation

To help prevent session fixation attacks you should [renew the session token after any privilege level change](https://www.owasp.org/index.php/Session_Management_Cheat_Sheet#Renew_the_Session_ID_After_Any_Privilege_Level_Change).

SCS provides a `RegenerateToken` helper, which should be called before making any changes to the session data that affect user privileges (such as login or logout operations).

`RegenerateToken` creates a new session token (while retaining the session data), deletes the old session token from the storage engine, and sends the new session token to the client.

```go
func loginHandler(w http.ResponseWriter, r *http.Request) {
    userID := 123

    // First regenerate the session token…
    err := session.RegenerateToken(r)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }

    // Then make the privilege-level change.
    err = session.PutInt(r, "userID", userID)
    if err != nil {
        http.Error(w, err.Error(), 500)
        return
    }
}
```

### Destroying data and sessions

There are four different functions for deleting session data:

* [Remove](https://godoc.org/github.com/alexedwards/scs/session#Remove) - Deletes a single key and corresponding value from the session data.
* [Clear](https://godoc.org/github.com/alexedwards/scs/session#Clear) - Deletes all keys and values in the session data.
* [Destroy](https://godoc.org/github.com/alexedwards/scs/session#Destroy) - Deletes all keys and values in the session data and removes the session from the storage engine. The client is instructed to delete the session cookie. 
* [Renew](https://godoc.org/github.com/alexedwards/scs/session#Renew) - Establishes a new, empty session. The old session is deleted from the storage engine. This is essentially a a concurrency-safe amalgamation of the `RegenerateToken` and `Clear` functions. 

## Custom storage engines

[`session.Engine`](https://godoc.org/github.com/alexedwards/scs/session#Engine) defines the interface for custom storage engines. Any object that implements this interface can be used as a storage engine when setting up the session manager middleware.

```go
type Engine interface {
    // Delete should remove the session token and corresponding data from the
    // session engine. If the token does not exist then Delete should be a no-op
    // and return nil (not an error).
    Delete(token string) (err error)

    // Find should return the data for a session token from the storage engine.
    // If the session token is not found or is expired, the found return value
    // should be false (and the err return value should be nil). Similarly, tampered
    // or malformed tokens should result in a found return value of false and a
    // nil err value. The err return value should be used for system errors only.
    Find(token string) (b []byte, found bool, err error)

    // Save should add the session token and data to the storage engine, with
    // the given expiry time. If the session token already exists, then the data
    // and expiry time should be overwritten.
    Save(token string, b []byte, expiry time.Time) (err error)
}
```

## Benchmarks

Performance of SCS is heavily influenced by the choice of storage engine. The following benchmarks simulate a HTTP request during which an existing session is loaded, an integer value is retreived, modified and the session is saved.

```
BenchmarkSCSMemstore-8                200000          8463 ns/op        3644 B/op         49 allocs/op
BenchmarkSCSCookies-8                 100000         20675 ns/op        7518 B/op         83 allocs/op
BenchmarkSCSRedis-8                    30000         43636 ns/op        3229 B/op         64 allocs/op
BenchmarkSCSPostgres-8                   500       3787304 ns/op        5584 B/op         96 allocs/op
BenchmarkSCSMySQL-8                      300       5511906 ns/op        4382 B/op         73 allocs/op
BenchmarkSCSBoltstore-8                  300       4086699 ns/op       12331 B/op        117 allocs/op
```

These benchmarks can be run from the `benchmark_test.go` file.

#### Comparisons

Trying to compare against other packages is difficult. Not only is real-world usage tough to simulate with simple benchmarks, things like community support and quality of tests are probably more important than raw performance in the long-term. 

That said, SCS stacks up pretty well. For the benchmarked operations it used around a quarter of the memory that [Gorilla Sessions](http://www.gorillatoolkit.org/pkg/sessions) did and operated between 1.5 and 3 times faster depending on the storage engine. 

```
BenchmarkGorillaCookies-8          20000         63678 ns/op       16987 B/op        296 allocs/op
BenchmarkGorillaRedis-8            10000        109229 ns/op       17877 B/op        336 allocs/op
BenchmarkGorillaPostgres-8           300       5460733 ns/op       24498 B/op        485 allocs/op
```

A big part of this performance difference is due to SCS's 'on-demand' use of Gob decoding. Accordingly, for operations which do need to call `GetObject` the performance difference is significantly less pronounced.

```
BenchmarkSCSObjectCookies-8            30000         60773 ns/op       17700 B/op        300 allocs/op
BenchmarkSCSObjectRedis-8              10000        104259 ns/op       13883 B/op        293 allocs/op
BenchmarkSCSObjectPostgres-8             500       3926530 ns/op       15124 B/op        313 allocs/op

BenchmarkGorillaObjectCookies-8        20000         67899 ns/op       19302 B/op        320 allocs/op
BenchmarkGorillaObjectRedis-8          10000        123880 ns/op       18976 B/op        360 allocs/op
BenchmarkGorillaObjectPostgres-8         300       4073790 ns/op       26589 B/op        509 allocs/op
```

The code for all the above benchmarks is available [in this gist](https://gist.github.com/alexedwards/9ba9a9f2ebd4e735713f8e9995f5640c).

## Notes

Full godoc documentation: [https://godoc.org/github.com/alexedwards/scs](https://godoc.org/github.com/alexedwards/scs).
