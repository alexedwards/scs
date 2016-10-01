# boltstore

[![godoc](https://godoc.org/github.com/alexedwards/scs/engine/boltstore?status.png)](https://godoc.org/github.com/alexedwards/scs/engine/boltstore)

Package boltstore is a [boltdb](https://github.com/boltdb/bolt) storage engine for the [SCS session package](https://godoc.org/github.com/alexedwards/scs/session).

This is a good option for local development, or a single-server deployment.
Unlike memstore the data will survive server restarts.

## Usage

### Installation

Either:

```
$ go get github.com/alexedwards/scs/engine/boltstore
```

Or (recommended) use use [gvt](https://github.com/FiloSottile/gvt) to vendor the `engine/boltstore` and `session` sub-packages:

```
$ gvt fetch github.com/alexedwards/scs/engine/boltstore
$ gvt fetch github.com/alexedwards/scs/session
```

### Setup

```go
package main

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/engine/boltstore"
	"github.com/alexedwards/scs/session"
	"github.com/boltdb/bolt"
)

func main() {
	// You provide a bolt.DB instance
	db, err := bolt.Open("testing.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create a new boltstore instance with bolt.DB
	// and a cleanup interval of 5 minutes.
	engine := boltstore.New(db, 5*time.Minute)

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

### Cleaning up expired session data

The boltstore package provides a background 'cleanup' goroutine to delete expired session data.
This stops the database file from holding on to invalid sessions indefinitely and growing unnecessarily large.

You can specify how frequently to run the cleanup when creating a new boltstore instance:

```go
// Run a cleanup every 30 minutes.
boltstore.New(db, 30*time.Minute)

// Setting the cleanup interval to zero prevents the cleanup from being run.
boltstore.New(db, 0)
```

#### Terminating the cleanup goroutine

It's rare that the cleanup goroutine for a boltstore instance needs to be terminated. It is generally intended to be long-lived and run for the lifetime of your application.

However, there may be occasions when your use of a boltstore instance is transient. A common example would be using it in a short-lived test function. In this scenario, the cleanup goroutine (which will run forever) will prevent the boltstore object from being garbage collected even after the test function has finished. You can prevent this by manually calling `StopCleanup()`.

For example:

```go
func TestExample(t *testing.T) {
    db, err := bolt.Open("testing.db", 0600, nil)
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    engine := New(db, time.Second)
    defer engine.StopCleanup()

    // Run test...
}
```

## Notes

The boltstore package is underpinned by the excellent [boltdb](https://github.com/boltdb/bolt).

Full godoc documentation: [https://godoc.org/github.com/alexedwards/scs/engine/boltstore](https://godoc.org/github.com/alexedwards/scs/engine/boltstore).
