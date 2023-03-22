# pgxstore

A PostgreSQL based session store for [SCS](https://github.com/alexedwards/scs) using the [pgx](https://github.com/jackc/pgx) driver.

## Setup

You should have a working PostgreSQL database containing a `sessions` table with the definition:

```sql
CREATE TABLE sessions (
	token TEXT PRIMARY KEY,
	data BYTEA NOT NULL,
	expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);
```

The database user for your application must have `SELECT`, `INSERT`, `UPDATE` and `DELETE` permissions on this table.

## Example

```go
package main

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

var sessionManager *scs.SessionManager

func main() {
	// Establish connection pool to PostgreSQL.
	pool, err := pgxpool.New(context.Background(), "postgres://username:password@host/dbname")
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Initialize a new session manager and configure it to use pgxstore as the session store.
	sessionManager = scs.New()
	sessionManager.Store = pgxstore.New(pool)

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

This package provides a background 'cleanup' goroutine to delete expired session data. This stops the database table from holding on to invalid sessions indefinitely and growing unnecessarily large. By default the cleanup runs every 5 minutes. You can change this by using the `NewWithCleanupInterval()` function to initialize your session store. For example:

```go
// Run a cleanup every 30 minutes.
pgxstore.NewWithCleanupInterval(conn, 30*time.Minute)

// Disable the cleanup goroutine by setting the cleanup interval to zero.
pgxstore.NewWithCleanupInterval(conn, 0)
```

### Terminating the Cleanup Goroutine

It's rare that the cleanup goroutine needs to be terminated --- it is generally intended to be long-lived and run for the lifetime of your application.

However, there may be occasions when your use of a session store instance is transient. A common example would be using it in a short-lived test function. In this scenario, the cleanup goroutine (which will run forever) will prevent the session store instance from being garbage collected even after the test function has finished. You can prevent this by either disabling the cleanup goroutine altogether (as described above) or by stopping it using the `StopCleanup()` method. For example:

```go
func TestExample(t *testing.T) {
	pool, err := pgxpool.New(context.Background(), "postgres://username:password@host/dbname")
	if err != nil {
		 t.Fatal(err)
	}
	defer pool.Close()

	store := pgxstore.New(pool)
	defer store.StopCleanup()

	sessionManager = scs.New()
	sessionManager.Store = store

	// Run test...
}
```
