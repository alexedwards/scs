package bunstore

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/uptrace/bun/driver/pgdriver"
)

func initWithCleanupInterval(t *testing.T, cleanupInterval time.Duration) *BunStore {
	var db *bun.DB
	var err error

	dialect := os.Getenv("SCS_BUN_TEST_DIALECT")
	switch dialect {
	default:
		dialect = "sqlite3"
		fallthrough
	case "postgres":
		dsn := os.Getenv("SCS_POSTGRES_TEST_DSN")
		sqldb, err := sql.Open("pg", dsn)
		if err != nil {
			sqldb.Close()
			t.Fatal(err)
		}

		sqldb.Exec(`DROP TABLE IF EXISTS sessions`)
		sqldb.Exec(`CREATE TABLE sessions (token TEXT PRIMARY KEY,data BYTEA NOT NULL,expiry TIMESTAMPTZ NOT NULL);`)
		sqldb.Exec(`CREATE INDEX sessions_expiry_idx ON sessions (expiry);`)

		db = bun.NewDB(sqldb, pgdialect.New())
	case "mysql":
		dsn := os.Getenv("SCS_MYSQL_TEST_DSN")
		sqldb, err := sql.Open("mysql", dsn)
		if err != nil {
			sqldb.Close()
			t.Fatal(err)
		}

		sqldb.Exec(`DROP TABLE IF EXISTS sessions`)
		sqldb.Exec(`CREATE TABLE sessions (token CHAR(43) PRIMARY KEY,data BLOB NOT NULL,expiry TIMESTAMP(6) NOT NULL);`)
		sqldb.Exec(`CREATE INDEX sessions_expiry_idx ON sessions (expiry);`)

		db = bun.NewDB(sqldb, mysqldialect.New())
	case "sqlite3":
		dsn := os.Getenv("./testSQL3lite.db")
		sqldb, err := sql.Open(sqliteshim.ShimName, dsn)
		if err != nil {
			sqldb.Close()
			t.Fatal(err)
		}

		sqldb.Exec(`DROP TABLE IF EXISTS sessions`)
		sqldb.Exec(`CREATE TABLE sessions (token TEXT PRIMARY KEY,data BLOB NOT NULL,expiry REAL NOT NULL);`)
		sqldb.Exec(`CREATE INDEX sessions_expiry_idx ON sessions(expiry);`)

		db = bun.NewDB(sqldb, sqlitedialect.New())
	}
	if err != nil {
		t.Fatal(err)
	}

	if db.Ping(); err != nil {
		db.Close()
		t.Fatal(err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1000)
	db.SetConnMaxLifetime(0)

	b, err := NewWithCleanupInterval(db, cleanupInterval)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	return b
}

func TestFind(t *testing.T) {
	b := initWithCleanupInterval(t, 0)
	ctx := context.Background()

	values := &map[string]interface{}{"token": "session_token", "data": []byte("encoded_data"), "expiry": time.Now().Add(1 * time.Minute)}
	if _, err := b.db.NewInsert().Model(values).TableExpr("sessions").Exec(ctx); err != nil {
		t.Fatal(err)
	}

	bb, found, err := b.FindCtx(ctx, "session_token")
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if bytes.Equal(bb, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", b, []byte("encoded_data"))
	}
}

func TestFindMissing(t *testing.T) {
	b := initWithCleanupInterval(t, 0)
	ctx := context.Background()

	_, found, err := b.FindCtx(ctx, "missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestSaveNew(t *testing.T) {
	b := initWithCleanupInterval(t, 0)
	ctx := context.Background()

	err := b.CommitCtx(ctx, "session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	row := b.db.QueryRow("SELECT data FROM sessions WHERE token = 'session_token'")
	var data []byte
	err = row.Scan(&data)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(data, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", data, []byte("encoded_data"))
	}
}

func TestSaveUpdated(t *testing.T) {
	b := initWithCleanupInterval(t, 0)
	ctx := context.Background()

	values := &map[string]interface{}{"token": "session_token", "data": []byte("encoded_data"), "expiry": time.Now().Add(1 * time.Minute)}
	if _, err := b.db.NewInsert().Model(values).TableExpr("sessions").Exec(ctx); err != nil {
		t.Fatal(err)
	}

	err := b.CommitCtx(ctx, "session_token", []byte("new_encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	row := b.db.QueryRow("SELECT data FROM sessions WHERE token = 'session_token'")
	var data []byte
	err = row.Scan(&data)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(data, []byte("new_encoded_data")) == false {
		t.Fatalf("got %v: expected %v", data, []byte("new_encoded_data"))
	}
}

func TestExpiry(t *testing.T) {
	b := initWithCleanupInterval(t, 0)
	ctx := context.Background()

	err := b.CommitCtx(ctx, "session_token", []byte("encoded_data"), time.Now().Add(1*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	_, found, _ := b.FindCtx(ctx, "session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(2 * time.Second)
	_, found, _ = b.FindCtx(ctx, "session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestDelete(t *testing.T) {
	b := initWithCleanupInterval(t, 0)
	ctx := context.Background()

	err := b.CommitCtx(ctx, "session_token", []byte("encoded_data"), time.Now().Add(1*time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	err = b.DeleteCtx(ctx, "session_token")
	if err != nil {
		t.Fatal(err)
	}

	row := b.db.QueryRow("SELECT COUNT(*) FROM sessions WHERE token = 'session_token'")
	var count int
	err = row.Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("got %d: expected %d", count, 0)
	}
}

func TestCleanup(t *testing.T) {
	b := initWithCleanupInterval(t, 2*time.Second)
	defer b.StopCleanup()
	ctx := context.Background()

	err := b.CommitCtx(ctx, "session_token", []byte("encoded_data"), time.Now().Add(1*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	row := b.db.QueryRow("SELECT COUNT(*) FROM sessions WHERE token = 'session_token'")
	var count int
	err = row.Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("got %d: expected %d", count, 1)
	}

	time.Sleep(3 * time.Second)
	row = b.db.QueryRow("SELECT COUNT(*) FROM sessions WHERE token = 'session_token'")
	err = row.Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("got %d: expected %d", count, 0)
	}
}

func TestStopNilCleanup(t *testing.T) {
	b := initWithCleanupInterval(t, 0)

	time.Sleep(100 * time.Millisecond)
	// A send to a nil channel will block forever
	b.StopCleanup()
}
