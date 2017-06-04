# filestore
[![godoc](https://godoc.org/github.com/alexedwards/scs/engine/filestore?status.png)](https://godoc.org/github.com/alexedwards/scs/engine/filestore)

Package filestore is a simple disk-based storage engine for the [SCS session package](https://godoc.org/github.com/alexedwards/scs/session).

Warning: Because filestore uses file based storage it is slow and should not be used in production.  It mearly exists to provide the convenience of memstore while maintaining data across server restarts. **dev only don't use in production**

## Usage

### Installation

Either:

```
$ go get github.com/alexedwards/scs/engine/filestore
```

Or (recommended) use use [gvt](https://github.com/FiloSottile/gvt) to vendor the `engine/filestore` and `session` sub-packages:

```
$ gvt fetch github.com/alexedwards/scs/engine/filestore
$ gvt fetch github.com/alexedwards/scs/session
```

### Example

```go
package main

import (
    "io"
    "net/http"
    "time"

    "github.com/alexedwards/scs/engine/filestore"
    "github.com/alexedwards/scs/session"
)

func main() {
    // Create a new filestore instance with a cleanup interval of 5 minutes.
    engine := filestore.New("/tmp/cookies.data", 5 * time.Minute)

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

The filestore package provides a background 'cleanup' goroutine to delete expired session data. This stops the underlying cache from holding on to invalid sessions forever and taking up unnecessary memory.

You can specify how frequently to run the cleanup when creating a new filestore instance:

```go
// Run a cleanup every 30 minutes.
filestore.New("/tmp/cookies.data", 30 * time.Minute)

// Setting the cleanup interval to zero prevents the cleanup from being run.
filestore.New("/tmp/cookies.data", 0)
```

## Notes

The filestore package is underpinned by the excellent [go-cache](https://github.com/patrickmn/go-cache) package.

Full godoc documentation: [https://godoc.org/github.com/alexedwards/scs/engine/filestore](https://godoc.org/github.com/alexedwards/scs/engine/filestore).
