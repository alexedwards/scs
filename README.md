# SCS: HTTP Session Management for Go

[![GoDoc](https://godoc.org/github.com/alexedwards/scs?status.png)](https://pkg.go.dev/github.com/alexedwards/scs/v2?tab=doc)
[![Go report card](https://goreportcard.com/badge/github.com/alexedwards/scs)](https://goreportcard.com/report/github.com/alexedwards/scs)
[![Test coverage](http://gocover.io/_badge/github.com/alexedwards/scs)](https://gocover.io/github.com/alexedwards/scs)

## Features

- Automatic loading and saving of session data via middleware.
- Choice of 19 different server-side session stores including PostgreSQL, MySQL, MSSQL, SQLite, Redis and many others. Custom session stores are also supported.
- Supports multiple sessions per request, 'flash' messages, session token regeneration, idle and absolute session timeouts, and 'remember me' functionality.
- Easy to extend and customize. Communicate session tokens to/from clients in HTTP headers or request/response bodies.
- Efficient design. Smaller, faster and uses less memory than [gorilla/sessions](https://github.com/gorilla/sessions).

## Instructions

- [SCS: HTTP Session Management for Go](#scs-http-session-management-for-go)
  - [Features](#features)
  - [Instructions](#instructions)
    - [Installation](#installation)
    - [Basic Use](#basic-use)
    - [Configuring Session Behavior](#configuring-session-behavior)
    - [Working with Session Data](#working-with-session-data)
    - [Loading and Saving Sessions](#loading-and-saving-sessions)
    - [Configuring the Session Store](#configuring-the-session-store)
    - [Using Custom Session Stores](#using-custom-session-stores)
      - [Using Custom Session Stores (with context.Context)](#using-custom-session-stores-with-contextcontext)
    - [Multiple Sessions per Request](#multiple-sessions-per-request)
    - [Enumerate All Sessions](#enumerate-all-sessions)
    - [Flushing and Streaming Responses](#flushing-and-streaming-responses)
    - [Compatibility](#compatibility)
    - [Contributing](#contributing)

### Installation

This package requires Go 1.12 or newer.

```sh
go get github.com/alexedwards/scs/v2
```

Note: If you're using the traditional `GOPATH` mechanism to manage dependencies, instead of modules, you'll need to `go get` and `import` `github.com/alexedwards/scs` without the `v2` suffix.

Please use [versioned releases](https://github.com/alexedwards/scs/releases). Code in tip may contain experimental features which are subject to change.

### Basic Use

SCS implements a session management pattern following the [OWASP security guidelines](https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/Session_Management_Cheat_Sheet.md). Session data is stored on the server, and a randomly-generated unique session token (or _session ID_) is communicated to and from the client in a session cookie.

```go
package main

import (
	"io"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

var sessionManager *scs.SessionManager

func main() {
	// Initialize a new session manager and configure the session lifetime.
	sessionManager = scs.New()
	sessionManager.Lifetime = 24 * time.Hour

	mux := http.NewServeMux()
	mux.HandleFunc("/put", putHandler)
	mux.HandleFunc("/get", getHandler)

	// Wrap your handlers with the LoadAndSave() middleware.
	http.ListenAndServe(":4000", sessionManager.LoadAndSave(mux))
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	// Store a new key and value in the session data.
	sessionManager.Put(r.Context(), "message", "Hello from a session!")
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	// Use the GetString helper to retrieve the string value associated with a
	// key. The zero value is returned if the key does not exist.
	msg := sessionManager.GetString(r.Context(), "message")
	io.WriteString(w, msg)
}
```

```
$ curl -i --cookie-jar cj --cookie cj localhost:4000/put
HTTP/1.1 200 OK
Cache-Control: no-cache="Set-Cookie"
Set-Cookie: session=lHqcPNiQp_5diPxumzOklsSdE-MJ7zyU6kjch1Ee0UM; Path=/; Expires=Sat, 27 Apr 2019 10:28:20 GMT; Max-Age=86400; HttpOnly; SameSite=Lax
Vary: Cookie
Date: Fri, 26 Apr 2019 10:28:19 GMT
Content-Length: 0

$ curl -i --cookie-jar cj --cookie cj localhost:4000/get
HTTP/1.1 200 OK
Date: Fri, 26 Apr 2019 10:28:24 GMT
Content-Length: 21
Content-Type: text/plain; charset=utf-8

Hello from a session!
```

### Configuring Session Behavior

Session behavior can be configured via the `SessionManager` fields.

```go
sessionManager = scs.New()
sessionManager.Lifetime = 3 * time.Hour
sessionManager.IdleTimeout = 20 * time.Minute
sessionManager.Cookie.Name = "session_id"
sessionManager.Cookie.Domain = "example.com"
sessionManager.Cookie.HttpOnly = true
sessionManager.Cookie.Path = "/example/"
sessionManager.Cookie.Persist = true
sessionManager.Cookie.SameSite = http.SameSiteStrictMode
sessionManager.Cookie.Secure = true
sessionManager.Cookie.Partitioned = true
```

Documentation for all available settings and their default values can be [found here](https://pkg.go.dev/github.com/alexedwards/scs/v2#SessionManager).

### Working with Session Data

Data can be set using the [`Put()`](https://pkg.go.dev/github.com/alexedwards/scs/v2#SessionManager.Put) method and retrieved with the [`Get()`](https://pkg.go.dev/github.com/alexedwards/scs/v2#SessionManager.Get) method. A variety of helper methods like [`GetString()`](https://pkg.go.dev/github.com/alexedwards/scs/v2#SessionManager.GetString), [`GetInt()`](https://pkg.go.dev/github.com/alexedwards/scs/v2#SessionManager.GetInt) and [`GetBytes()`](https://pkg.go.dev/github.com/alexedwards/scs/v2#SessionManager.GetBytes) are included for common data types. Please see [the documentation](https://pkg.go.dev/github.com/alexedwards/scs/v2#pkg-index) for a full list of helper methods.

The [`Pop()`](https://pkg.go.dev/github.com/alexedwards/scs/v2#SessionManager.Pop) method (and accompanying helpers for common data types) act like a one-time `Get()`, retrieving the data and removing it from the session in one step. These are useful if you want to implement 'flash' message functionality in your application, where messages are displayed to the user once only.

Some other useful functions are [`Exists()`](https://pkg.go.dev/github.com/alexedwards/scs/v2#SessionManager.Exists) (which returns a `bool` indicating whether or not a given key exists in the session data) and [`Keys()`](https://pkg.go.dev/github.com/alexedwards/scs/v2#SessionManager.Keys) (which returns a sorted slice of keys in the session data).

Individual data items can be deleted from the session using the [`Remove()`](https://pkg.go.dev/github.com/alexedwards/scs/v2#SessionManager.Remove) method. Alternatively, all session data can be deleted by using the [`Destroy()`](https://pkg.go.dev/github.com/alexedwards/scs/v2#SessionManager.Destroy) method. After calling `Destroy()`, any further operations in the same request cycle will result in a new session being created --- with a new session token and a new lifetime.

Behind the scenes SCS uses gob encoding to store session data, so if you want to store custom types in the session data they must be [registered](https://golang.org/pkg/encoding/gob/#Register) with the encoding/gob package first. Struct fields of custom types must also be exported so that they are visible to the encoding/gob package. Please [see here](https://gist.github.com/alexedwards/d6eca7136f98ec12ad606e774d3abad3) for a working example.

### Loading and Saving Sessions

Most applications will use the [`LoadAndSave()`](https://pkg.go.dev/github.com/alexedwards/scs/v2#SessionManager.LoadAndSave) middleware. This middleware takes care of loading and committing session data to the session store, and communicating the session token to/from the client in a cookie as necessary.

If you want to customize the behavior (like communicating the session token to/from the client in a HTTP header, or creating a distributed lock on the session token for the duration of the request) you are encouraged to create your own alternative middleware using the code in [`LoadAndSave()`](https://pkg.go.dev/github.com/alexedwards/scs/v2#SessionManager.LoadAndSave) as a template. An example is [given here](https://gist.github.com/alexedwards/cc6190195acfa466bf27f05aa5023f50).

Or for more fine-grained control you can load and save sessions within your individual handlers (or from anywhere in your application). [See here](https://gist.github.com/alexedwards/0570e5a59677e278e13acb8ea53a3b30) for an example.

### Configuring the Session Store

By default SCS uses an in-memory store for session data. This is convenient (no setup!) and very fast, but all session data will be lost when your application is stopped or restarted. Therefore it's useful for applications where data loss is an acceptable trade off for fast performance, or for prototyping and testing purposes. In most production applications you will want to use a persistent session store like PostgreSQL or MySQL instead.

The session stores currently included are shown in the table below. Please click the links for usage instructions and examples.

| Package                                                                             | Backend                                                                         | Embedded | In-Memory | Multi-Process |
| :---------------------------------------------------------------------------------- | --------------------------------------------------------------------------------|----------|-----------|---------------| 
| [badgerstore](https://github.com/alexedwards/scs/tree/master/badgerstore)           | [BadgerDB](https://dgraph.io/docs/badger/)                                      | Y | N | N |
| [boltstore](https://github.com/alexedwards/scs/tree/master/boltstore)               | [BBolt](https://go.etcd.io/bbolt)                                               | Y | N | N |
| [bunstore](https://github.com/alexedwards/scs/tree/master/bunstore)                 | [Bun](https://bun.uptrace.dev/) ORM for PostgreSQL/MySQL/MSSQL/SQLite           | N | N | Y | 
| [buntdbstore](https://github.com/alexedwards/scs/tree/master/buntdbstore)           | [BuntDB](https://github.com/tidwall/buntdb)                                     | Y | Y | N |
| [cockroachdbstore](https://github.com/alexedwards/scs/tree/master/cockroachdbstore) | [CockroachDB](https://www.cockroachlabs.com/)                                   | N | N | Y |
| [consulstore](https://github.com/alexedwards/scs/tree/master/consulstore)           | [Consul](https://www.consul.io/)                                                | N | Y | Y |
| [etcdstore](https://github.com/alexedwards/scs/tree/master/etcdstore)               | [Etcd](https://etcd.io/)                                                        | N | N | Y |
| [firestore](https://github.com/alexedwards/scs/tree/master/firestore)               | [Google Cloud Firestore](https://cloud.google.com/firestore)                    | N | ? | Y |
| [gormstore](https://github.com/alexedwards/scs/tree/master/gormstore)               | [GORM](https://gorm.io/index.html) ORM for PostgreSQL/MySQL/SQLite/MSSQL/TiDB   | N | N | Y |
| [leveldbstore](https://github.com/alexedwards/scs/tree/master/leveldbstore)         | [LevelDB](https://github.com/syndtr/goleveldb)                                  | Y | N | N |
| [memstore](https://github.com/alexedwards/scs/tree/master/memstore)                 | In-memory (default)                                                             | Y | Y | N |
| [mongodbstore](https://github.com/alexedwards/scs/tree/master/mongodbstore)         | [MongoDB](https://www.mongodb.com/)                                             | N | N | Y |
| [mssqlstore](https://github.com/alexedwards/scs/tree/master/mssqlstore)             | [Microsoft SQL Server](https://www.microsoft.com/en-us/sql-server)              | N | N | Y | 
| [mysqlstore](https://github.com/alexedwards/scs/tree/master/mysqlstore)             | [MySQL](https://www.mysql.com/)                                                 | N | N | Y |
| [pgxstore](https://github.com/alexedwards/scs/tree/master/pgxstore)                 | [PostgreSQL](https://www.postgresql.org/) (using the [pgx](https://github.com/jackc/pgx) driver) | N | N | Y |
| [postgresstore](https://github.com/alexedwards/scs/tree/master/postgresstore)       | [PostgreSQL](https://www.postgresql.org/) (using the [pq](https://github.com/lib/pq) driver)     | N | N | Y |
| [redisstore](https://github.com/alexedwards/scs/tree/master/redisstore)             | [Redis](https://redis.io/)                                                      | N | Y | Y |
| [sqlite3store](https://github.com/alexedwards/scs/tree/master/sqlite3store)         | [SQLite3](https://sqlite.org/) (using the [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) CGO-based driver) | Y | N | Y |

Custom session stores are also supported. Please [see here](#using-custom-session-stores) for more information.

### Using Custom Session Stores

[`scs.Store`](https://pkg.go.dev/github.com/alexedwards/scs/v2#Store) defines the interface for custom session stores. Any object that implements this interface can be set as the store when configuring the session.

```go
type Store interface {
	// Delete should remove the session token and corresponding data from the
	// session store. If the token does not exist then Delete should be a no-op
	// and return nil (not an error).
	Delete(token string) (err error)

	// Find should return the data for a session token from the store. If the
	// session token is not found or is expired, the found return value should
	// be false (and the err return value should be nil). Similarly, tampered
	// or malformed tokens should result in a found return value of false and a
	// nil err value. The err return value should be used for system errors only.
	Find(token string) (b []byte, found bool, err error)

	// Commit should add the session token and data to the store, with the given
	// expiry time. If the session token already exists, then the data and
	// expiry time should be overwritten.
	Commit(token string, b []byte, expiry time.Time) (err error)
}

type IterableStore interface {
	// All should return a map containing data for all active sessions (i.e.
	// sessions which have not expired). The map key should be the session
	// token and the map value should be the session data. If no active
	// sessions exist this should return an empty (not nil) map.
	All() (map[string][]byte, error)
}
```

#### Using Custom Session Stores (with context.Context)

[`scs.CtxStore`](https://pkg.go.dev/github.com/alexedwards/scs/v2#CtxStore) defines the interface for custom session stores (with methods take context.Context parameter).

```go
type CtxStore interface {
	Store

	// DeleteCtx is the same as Store.Delete, except it takes a context.Context.
	DeleteCtx(ctx context.Context, token string) (err error)

	// FindCtx is the same as Store.Find, except it takes a context.Context.
	FindCtx(ctx context.Context, token string) (b []byte, found bool, err error)

	// CommitCtx is the same as Store.Commit, except it takes a context.Context.
	CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) (err error)
}

type IterableCtxStore interface {
	// AllCtx is the same as IterableStore.All, expect it takes a
	// context.Context.
	AllCtx(ctx context.Context) (map[string][]byte, error)
}
```

### Preventing Session Fixation

To help prevent session fixation attacks you should [renew the session token after any privilege level change](https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/Session_Management_Cheat_Sheet.md#renew-the-session-id-after-any-privilege-level-change). Commonly, this means that the session token must to be changed when a user logs in or out of your application. You can do this using the [`RenewToken()`](https://pkg.go.dev/github.com/alexedwards/scs/v2#SessionManager.RenewToken) method like so:

```go
func loginHandler(w http.ResponseWriter, r *http.Request) {
	userID := 123

	// First renew the session token...
	err := sessionManager.RenewToken(r.Context())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	// Then make the privilege-level change.
	sessionManager.Put(r.Context(), "userID", userID)
}
```

### Multiple Sessions per Request

It is possible for an application to support multiple sessions per request, with different lifetime lengths and even different stores. Please [see here for an example](https://gist.github.com/alexedwards/22535f758356bfaf96038fffad154824).

### Enumerate All Sessions


To iterate throught all sessions, SCS offers to all data stores an `All()` function where they can return their own sessions.

Essentially, in your code, you pass the `Iterate()` method a closure with the signature `func(ctx context.Context) error` which contains the logic that you want to execute against each session. For example, if you want to revoke all sessions with contain a `userID` value equal to `4` you can do the following:

```go
err := sessionManager.Iterate(r.Context(), func(ctx context.Context) error {
	userID := sessionManager.GetInt(ctx, "userID")

	if userID == 4 {
		return sessionManager.Destroy(ctx)
	}

	return nil
})
if err != nil {
	log.Fatal(err)
}
```

### Flushing and Streaming Responses

Flushing responses is supported via the `http.NewResponseController` type (available in Go >= 1.20).

```go
func flushingHandler(w http.ResponseWriter, r *http.Request) {
	sessionManager.Put(r.Context(), "message", "Hello from a flushing handler!")

	rc := http.NewResponseController(w)

	for i := 0; i < 5; i++ {
		fmt.Fprintf(w, "Write %d\n", i)

		err := rc.Flush()
		if err != nil {
			log.Println(err)
			return
		}

		time.Sleep(time.Second)
	}
}
```

For a complete working example, please see [this comment](https://github.com/alexedwards/scs/issues/141#issuecomment-1774050802).

Note that the `http.ResponseWriter` passed on by the [`LoadAndSave()`](https://pkg.go.dev/github.com/alexedwards/scs/v2#SessionManager.LoadAndSave) middleware does not support the `http.Flusher` interface directly. This effectively means that flushing/streaming is only supported by SCS if you are using Go >= 1.20.

### Compatibility

You may have some problems using this package with Go frameworks that do not propagate the request context from standard-library compatible middleware, like [Echo](https://github.com/alexedwards/scs/issues/57) and [Fiber](https://github.com/alexedwards/scs/issues/106). If you are using Echo, you may wish to evaluate using the [echo-scs-session](https://github.com/canidam/echo-scs-session) package for session management.

### Contributing

Bug fixes and documentation improvements are very welcome! For feature additions or behavioral changes, please open an issue to discuss the change before submitting a PR. Additional store implementations will not merged to this repository (unless there is very significant demand) --- but please feel free to host the store implementation yourself and open a PR to link to it from this README.
