# bunstore

A [Bun](https://github.com/uptrace/bun) based session store for [SCS](https://github.com/alexedwards/scs).

## Setup

You should have a working database containing a `sessions` table with the definition (for PostgreSQL):

```sql
CREATE TABLE sessions (
	token TEXT PRIMARY KEY,
	data BYTEA NOT NULL,
	expiry TIMESTAMPTZ NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);
```
For other stores you can find the setup here: [MySQL](https://github.com/alexedwards/scs/tree/master/mysqlstore), [SQLite3](https://github.com/alexedwards/scs/tree/master/sqlite3store).

If no table is present, a new one will be automatically created.

The database user for your application must have `CREATE TABLE`, `SELECT`, `INSERT`, `UPDATE` and `DELETE` permissions on this table.

## Example

```go
package main

import (
	"database/sql"
	"io"
	"log"
	"net/http"

	"github.com/alexedwards/scs/bunstore"
	"github.com/alexedwards/scs/v2"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	_ "github.com/uptrace/bun/driver/pgdriver"
)

var sessionManager *scs.SessionManager

func main() {
	// Establish connection to your store.
	sqldb, err := sql.Open("pg", "postgres://username:password@host/dbname") // PostgreSQL
	//sqldb, err := sql.Open("mysql", "username:password@tcp(host)/dbname?parseTime=true") // MySQL
	//sqldb, err := sql.Open(sqliteshim.ShimName, "sqlite3_database.db") // SQLite3
	if err != nil {
		log.Fatal(err)
	}

	db := bun.NewDB(sqldb, pgdialect.New()) // PostgreSQL
	//db := bun.NewDB(sqldb, mysqldialect.New()) // MySQL
	//db := bun.NewDB(sqldb, sqlitedialect.New()) // SQLite3
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1000)
	db.SetConnMaxLifetime(0)

	// Initialize a new session manager and configure it to use bunstore as the session store.
	sessionManager = scs.New()
	if sessionManager.Store, err = bunstore.New(db); err != nil {
        log.Fatal(err)
    }

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
bunstore.NewWithCleanupInterval(db, 30*time.Minute)

// Disable the cleanup goroutine by setting the cleanup interval to zero.
bunstore.NewWithCleanupInterval(db, 0)
```

### Terminating the Cleanup Goroutine

It's rare that the cleanup goroutine needs to be terminated --- it is generally intended to be long-lived and run for the lifetime of your application.

However, there may be occasions when your use of a session store instance is transient. A common example would be using it in a short-lived test function. In this scenario, the cleanup goroutine (which will run forever) will prevent the session store instance from being garbage collected even after the test function has finished. You can prevent this by either disabling the cleanup goroutine altogether (as described above) or by stopping it using the `StopCleanup()` method. For example:

```go
func TestExample(t *testing.T) {
	sqldb, err := sql.Open("pg", "postgres://username:password@host/dbname")
	if err != nil {
		log.Fatal(err)
	}
	
	db := bun.NewDB(sqldb, pgdialect.New())
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1000)
	db.SetConnMaxLifetime(0)

    store, err := bunstore.New(db)
    if err != nil {
	    t.Fatal(err)
    }
	defer store.StopCleanup()

	sessionManager = scs.New()
	sessionManager.Store = store

	// Run test...
}
```