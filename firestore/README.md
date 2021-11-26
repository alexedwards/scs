# firestore

A [Google Cloud Firestore](https://pkg.go.dev/cloud.google.com/go/firestore) based session store for [SCS](https://github.com/alexedwards/scs).

## Setup

You should follow the instructions to [install and open a database](https://cloud.google.com/firestore/docs), and pass the database to `firestore.New()` to establish the session store. 

The default collection is "Sessions". If you want to change that, store a custom CollectionRef in scsfs.Sessions.

## Example

```go
package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	scsfs "github.com/alexedwards/scs/firestore"
	"github.com/alexedwards/scs/v2"
)

var sessionManager *scs.SessionManager

func main() {
	// Establish connection to Google Cloud Firestore.
	db, err := firestore.NewClient(context.Background(), os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		log.Fatal(err)
	}

	// Initialize a new session manager and configure it to use firestore as the session store.
	sessionManager = scs.New()
	sessionManager.Store = scsfs.New(db)

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

The sample can be run by setting the environment variable GOOGLE_CLOUD_PROJECT, which should be proper project id. If you have the local firestore emulator (part of the Google Cloud SDK) installed you can start a local emulator and run the example like this:

Start the emulator first:

```sh
gcloud beta emulators firestore start --host-port=localhost:8041
```

Then test the example program:

```sh
FIRESTORE_EMULATOR_HOST=localhost:8041 GOOGLE_CLOUD_PROJECT=test go run .
```