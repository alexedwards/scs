# buntstore

[![godoc](https://godoc.org/github.com/alexedwards/scs/engine/buntstore?status.png)](https://godoc.org/github.com/alexedwards/scs/engine/buntstore)

Package buntstore is a [buntdb](https://github.com/tidwall/buntdb) storage engine for the [SCS session package](https://godoc.org/github.com/alexedwards/scs/session).

This is a good option for local development, or a single-server deployment.
Unlike memstore the data will survive server restarts.

It provides better performance than boltdb, but only syncs to disk every second. *You could lose data.* You have to decide if it is acceptable
for your application to loose the last second of session data.

## Usage

### Installation

Either:

```
$ go get github.com/alexedwards/scs/engine/buntstore
```

Or (recommended) use use some kind of vendoring


### Setup

```go
package main

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/engine/buntstore"
	"github.com/alexedwards/scs/session"
	"github.com/tidwall/buntdb"
)

func main() {
	// You provide a bunt.DB instance
	db, err := buntdb.Open("/tmp/bunt.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create a new buntstore instance with bunt.DB
	engine := buntstore.New(db)

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

The buntdb package provides a background 'cleanup' goroutine to delete expired session data.
This stops the database file from holding on to invalid sessions indefinitely and growing unnecessarily large.

Additionally on every 'Get' the item is checked for expiration.

buntdb uses an AOF to store the data. It provied an autoshrink feature wich does not block the main go routine and reduces the file
to current db contents.


## Notes

The buntstore package is underpinned by the excellent [buntdb](https://github.com/tidwall/buntdb).

Full godoc documentation: [https://godoc.org/github.com/alexedwards/scs/engine/buntstore](https://godoc.org/github.com/alexedwards/scs/engine/buntstore).
