
# qlstore 
[![godoc](https://godoc.org/github.com/alexedwards/scs/engine/qlstore?status.png)](https://godoc.org/github.com/alexedwards/scs/engine/qlstore)

Package qlstore is a [ql](https://github.com/cznic/ql) storage engine for the [SCS session package](https://godoc.org/github.com/alexedwards/scs/session).

## Usage

### Installation

Either:

```
$ go get github.com/alexedwards/scs/engine/qlstore
```

Or (recommended) use use [gvt](https://github.com/FiloSottile/gvt) to vendor the `engine/qlstore` and `session` sub-packages:

```
$ gvt fetch github.com/alexedwards/scs/engine/qlstore
$ gvt fetch github.com/alexedwards/scs/session
```

### Setup

A working ql database is required, containing a `sessions` table with the definition:

```sql
CREATE TABLE sessions (
	token string,
	data blob,
	expiry time
);
CREATE INDEX sessions_expiry_idx ON sessions (expiry);
```

### Example

```go
package main

import (
    "database/sql"
    "io"
    "log"
    "net/http"
    "time"

    "github.com/alexedwards/scs/engine/qlstore"
    "github.com/alexedwards/scs/session"
)

func main() {
    // Establish a database/sql pool
    db, err := sql.Open("ql", "example.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Create a new qlstore instance using the existing database/sql pool, 
    // with a cleanup interval of 5 minutes.
    engine := qlstore.New(db, 5*time.Minute)

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

The qlstore package provides a background 'cleanup' goroutine to delete expired session data. This stops the database table from holding on to invalid sessions indefinitely and growing unnecessarily large.

You can specify how frequently to run the cleanup when creating a new qlstore instance:

```go
// Run a cleanup every 30 minutes.
qlstore.New(db, 30*time.Minute)

// Setting the cleanup interval to zero prevents the cleanup from being run.
qlstore.New(db, 0)
```

#### Terminating the cleanup goroutine

It's rare that the cleanup goroutine for a qlstore instance needs to be terminated. It is generally intended to be long-lived and run for the lifetime of your application.

However, there may be occasions when your use of a qlstore instance is transient. A common example would be using it in a short-lived test function. In this scenario, the cleanup goroutine (which will run forever) will prevent the qlstore object from being garbage collected even after the test function has finished. You can prevent this by manually calling `StopCleanup()`.

For example:

```go
func TestExample(t *testing.T) {
    db, err := sql.Open("ql", "test.db")
    if err != nil {
        t.Fatal(err)
    }
    defer db.Close()

    engine := New(db, time.Second)
    defer engine.StopCleanup()

    // Run test...
}
```

## Notes

The qlstore package is underpinned by the excellent [ql](https://github.com/cznic/ql) driver.

Full godoc documentation: [https://godoc.org/github.com/alexedwards/scs/engine/qlstore](https://godoc.org/github.com/alexedwards/scs/engine/qlstore).
