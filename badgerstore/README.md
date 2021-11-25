# badgerstore

A [Badger](https://github.com/dgraph-io/badger) based session store for [SCS](https://github.com/alexedwards/scs).

## Setup

You should follow the instructions to [install and open a database](https://github.com/dgraph-io/badger#installing), and pass the database to `badgerstore.New()` to establish the session store.

## Example

```go
package main

import (
	"io"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/badgerstore"
	"github.com/dgraph-io/badger"
)

var sessionManager *scs.SessionManager

func main() {
	// Open a Badger database.
	db, err := badger.Open(badger.DefaultOptions("tmp/badger"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize a new session manager and configure it to use badgerstore as the session store.
	sessionManager = scs.New()
	sessionManager.Store = badgerstore.New(db)

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

Badger will [automatically remove](https://github.com/dgraph-io/badger#setting-time-to-livettl-and-user-metadata-on-keys) expired session keys.

## Key Collisions

By default keys are in the form `scs:session:<token>`. For example:

```
"scs:session:ZnirGwi2FiLwXeVlP5nD77IpfJZMVr6un9oZu2qtJrg"
```

Because the token is highly unique, key collisions are not a concern. But if you're configuring *multiple session managers*, both of which use `badgerstore`, then you may want the keys to have a different prefix depending on which session manager wrote them. You can do this by using the `NewWithPrefix()` method like so:

```go
db, err := badger.Open(badger.DefaultOptions("tmp/badger"))
if err != nil {
	log.Fatal(err)
}
defer db.Close()

sessionManagerOne = scs.New()
sessionManagerOne.Store = badgerstore.NewWithPrefix(db, "scs:session:1:")

sessionManagerTwo = scs.New()
sessionManagerTwo.Store = badgerstore.NewWithPrefix(db, "scs:session:2:")
```
