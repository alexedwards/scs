# buntdbstore

A [BuntDB](https://github.com/tidwall/buntdb)-based session store for [SCS](https://github.com/alexedwards/scs).

## Example

You should follow the instructions to [install and open a database](https://github.com/tidwall/buntdb#installing), and pass the database to `buntdbstore.New()` to establish the session store.

```go
package main

import (
	"io"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/buntdbstore"
	"github.com/tidwall/buntdb"
)

var sessionManager *scs.SessionManager

func main() {
	// Create a BuntDB database.
	db, err := buntdb.Open("tmp/buntdb.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize a new session manager and configure it to use buntdbstore as
	// the session store.
	sessionManager = scs.New()
	sessionManager.Store = buntdbstore.New(db)

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

BuntDB will [automatically remove](https://github.com/tidwall/buntdb#data-expiration) expired session keys.
