# memstore 
[![godoc](https://godoc.org/github.com/alexedwards/scs/engine/memstore?status.png)](https://godoc.org/github.com/alexedwards/scs/engine/memstore)

Package memstore is an in-memory storage engine for the [SCS session package](https://godoc.org/github.com/alexedwards/scs/session).

Warning: Because memstore uses in-memory storage only, all session data will be lost when your Go program is stopped or restarted. On the upside though, it is blazingly fast.

In production, memstore should only be used where this volatility is an acceptable trade-off for the high performance, and where lost session data will have a negligible impact on users. As an example, a use case could be using it to track which adverts a visitor has already been shown.

## Usage

### Installation

Either:

```
$ go get github.com/alexedwards/scs/engine/memstore
```

Or (recommended) use use [gvt](https://github.com/FiloSottile/gvt) to vendor the `engine/memstore` and `session` sub-packages:

```
$ gvt fetch github.com/alexedwards/scs/engine/memstore
$ gvt fetch github.com/alexedwards/scs/session
```

### Example

```go
package main

import (
    "io"
    "net/http"
    "time"

    "github.com/alexedwards/scs/engine/memstore"
    "github.com/alexedwards/scs/session"
)

func main() {
    // Create a new MemStore instance with a cleanup interval of 5 minutes.
    engine := memstore.New(5 * time.Minute)

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

The memstore package provides a background 'cleanup' goroutine to delete expired session data. This stops the underlying cache from holding on to invalid sessions forever and taking up unnecessary memory.

You can specify how frequently to run the cleanup when creating a new MemStore instance:

```go
// Run a cleanup every 30 minutes.
memstore.New(30 * time.Minute)

// Setting the cleanup interval to zero prevents the cleanup from being run.
memstore.New(0)
```

## Notes

The memstore package is underpinned by the excellent [go-cache](https://github.com/patrickmn/go-cache) package.

Full godoc documentation: [https://godoc.org/github.com/alexedwards/scs/engine/memstore](https://godoc.org/github.com/alexedwards/scs/engine/memstore).