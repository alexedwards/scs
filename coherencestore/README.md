# coherence

A [Coherence](https://github.com/oracle/coherence) backed session store for [SCS](https://github.com/alexedwards/scs) using the [Coherence Go Client](https://github.com/oracle/coherence-go-client).

> Note: This implementation requires Go 1.19 or above.

# Setup

Before running or testing this implementation, you must ensure a Coherence cluster is available. 
For local development, we recommend using the Coherence CE Docker image; it contains everything 
necessary for the client to operate correctly.

To start a Coherence cluster using Docker, issue the following:

```bash
docker run -d -p 1408:1408 ghcr.io/oracle/coherencestore-ce:22.06.4
```

See the documentation [here](https://pkg.go.dev/github.com/oracle/coherence-go-client/coherence#hdr-Obtaining_a_Session) on connection options
when creating a Coherence session.

# Example

```go
package main

import (
	"context"
	"github.com/alexedwards/scs/coherencestore"
	"github.com/alexedwards/scs/v2"
	"github.com/oracle/coherence-go-client/coherence"
	"io"
	"log"
	"net/http"
)

var sessionManager *scs.SessionManager

func main() {
	ctx := context.Background()

	// Create a Coherence session connecting to the default gRPC port of 1408 using plain text
	session, err := coherence.NewSession(ctx, coherence.WithPlainText())
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	
	// Initialize a new session manager and configure it to use Coherence as the session store
	sessionManager = scs.New()

	store, err := coherencestore.New(session)
	if err != nil {
		log.Fatal(err)
	}
	sessionManager.Store = store

	mux := http.NewServeMux()
	mux.HandleFunc("/put", putHandler)
	mux.HandleFunc("/get", getHandler)

	_ = http.ListenAndServe(":4000", sessionManager.LoadAndSave(mux))
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	sessionManager.Put(r.Context(), "message", "Hello from a Coherence backed session!")
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	msg := sessionManager.GetString(r.Context(), "message")
	_, err := io.WriteString(w, msg)
	if err != nil {
		return
	}
}
```


# Expired Session Cleanup

Coherence will automatically clean up expired sessions.

# Key Collisions

By default, Coherence will use a cache named `default-session-store` to store tokens.
They keys should be unique enough for a single backend cache, but if you would like to have multiple, unique session managers you can specify a cache
name on startup. E.g.

Using the **default** session cache name:

```go
sessionManager = scs.New()
coherenceStore, err := New(session)
if err != nil {
    log.Fatal(err)
}
sessionManager.Store = coherenceStore
```

Using **your own** session cache name:

```go
coherenceStore, err := NewWithCache(session, "my-session-cache")
if err != nil {
    log.Fatal(err)
}
sessionManager.Store = coherenceStore
```