# xormstore

A [Xorm](https://xorm.io/) based session store for [SCS](https://github.com/alexedwards/scs).

## Setup

`xormstore` will auto migrate `sessions` table
## Example

```go
package main

import (
	"io"
	"log"
	"net/http"

	"github.com/alexedwards/scs/xormstore"
	"github.com/alexedwards/scs/v2"

	"xorm.io/xorm"

  _ "github.com/go-sql-driver/mysql"
)

var sessionManager *scs.SessionManager

func main() {
	// Establish connection to your store.
	db, err := xorm.NewEngine("mysql", "root:123@/test?charset=utf8") // MySQL
	// db, err := xorm.NewEngine("sqlite3", "data.db") // SQLite3
	if err != nil {
		log.Fatal(err)
	}

	// Initialize a new session manager and configure it to use gormstore as the session store.
	sessionManager = scs.New()
	if sessionManager.Store, err = xormstore.New(db); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/put", putHandler)
	mux.HandleFunc("/get", getHandler)

	log.Println("OK")
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

This package provides a background 'cleanup' goroutine to delete expired session data. This stops the database table from holding on to invalid sessions indefinitely and growing unnecessarily large. By default the cleanup runs every 5 minutes. You can change this by using the `NewWithCleanupInterval()` function to initialize your session store. For example:

```go
// Run a cleanup every 30 minutes.
xormstore.NewWithCleanupInterval(db, 30*time.Minute)

// Disable the cleanup goroutine by setting the cleanup interval to zero.
xormstore.NewWithCleanupInterval(db, 0)
```

### Terminating the Cleanup Goroutine

It's rare that the cleanup goroutine needs to be terminated --- it is generally intended to be long-lived and run for the lifetime of your application.

However, there may be occasions when your use of a session store instance is transient. A common example would be using it in a short-lived test function. In this scenario, the cleanup goroutine (which will run forever) will prevent the session store instance from being garbage collected even after the test function has finished. You can prevent this by either disabling the cleanup goroutine altogether (as described above) or by stopping it using the `StopCleanup()` method. For example:

```go
func TestExample(t *testing.T) {
	db, err := xorm.NewEngine("mysql", "root:123@/test?charset=utf8")
	if err != nil {
	    t.Fatal(err)
	}
	defer db.Close()

  store, err := xormstore.New(db)
  if err != nil {
	  t.Fatal(err)
  }
	defer store.StopCleanup()

	sessionManager = scs.New()
	sessionManager.Store = store

	// Run test...
}
```
