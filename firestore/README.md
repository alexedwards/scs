# firestore

A [Google Cloud Firestore](https://pkg.go.dev/cloud.google.com/go/firestore) based session store for [SCS](https://github.com/alexedwards/scs).

## Setup

You should follow the instructions to [install and open a database](https://cloud.google.com/firestore/docs), and pass the database to `firestore.New()` to establish the session store.

## Example

```go
package main

import (
	"io"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/firestore"
)

var sessionManager *scs.SessionManager

func main() {
	// Establish connection to Google Cloud Firestore.
	// ...
	
	// Initialize a new session manager and configure it to use firestore as the session store.
	sessionManager = scs.New()
	sessionManager.Store = firestore.New(db)

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
