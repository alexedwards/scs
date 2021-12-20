# leveldbstore

A [LevelDB](https://github.com/syndtr/goleveldb) based session store for [SCS](https://github.com/alexedwards/scs).

## Setup

You should follow the instructions to [install and open a database](https://github.com/syndtr/goleveldb#installation), and pass the database to `leveldbstore.New()` to establish the session store.

## Example

```go
package main

import (
	"io"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/leveldbstore"
	"github.com/syndtr/goleveldb/leveldb"
)

var sessionManager *scs.SessionManager

func main() {
	// Open a LevelDB database.
	db, err := leveldb.OpenFile("tmp/leveldb.db", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	// Initialize a new session manager and configure it to use leveldbstore as the session store.
	sessionManager = scs.New()
	sessionManager.Store = leveldbstore.New(db)

	mux := http.NewServeMux()
	mux.HandleFunc("/put", putHandler)
	mux.HandleFunc("/get", getHandler)

	http.ListenAndServe(":4000", sessionManager.LoadAndSave(mux))
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	sessionManager.Put(r.Context(), "message", "Hello from a session!")
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	msg := sessionManager.GetString(r.Context(), "message")
	io.WriteString(w, msg)
}
```

## Expired Session Cleanup

This package provides a background 'cleanup' goroutine to delete expired session data. This stops the database table from holding on to invalid sessions indefinitely and growing unnecessarily large. By default the cleanup runs every 1 minute. You can change this by using the `NewWithCleanupInterval()` function to initialize your session store. For example:

```go
// Run a cleanup every 5 minutes.
leveldbstore.NewWithCleanupInterval(db, 5*time.Minute)

// Disable the cleanup goroutine by setting the cleanup interval to zero.
leveldbstore.NewWithCleanupInterval(db, 0)
```

### Terminating the Cleanup Goroutine

It's rare that the cleanup goroutine needs to be terminated --- it is generally intended to be long-lived and run for the lifetime of your application.

However, there may be occasions when your use of a session store instance is transient. A common example would be using it in a short-lived test function. In this scenario, the cleanup goroutine (which will run forever) will prevent the session store instance from being garbage collected even after the test function has finished. You can prevent this by either disabling the cleanup goroutine altogether (as described above) or by stopping it using the `StopCleanup()` method. For example:

```go
func TestExample(t *testing.T) {
	db, err := leveldb.OpenFile("tmp/leveldb.db", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := leveldbstore.New(db)
	defer store.StopCleanup()

	sessionManager = scs.New()
	sessionManager.Store = store

	// Run test...
}
```
