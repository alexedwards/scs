package pgxstore

import (
	"bytes"
	"context"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestFind(t *testing.T) {
	ctx := context.Background()

	dsn := os.Getenv("SCS_POSTGRES_TEST_DSN")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, "TRUNCATE TABLE sessions")
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, "INSERT INTO sessions VALUES('session_token', 'encoded_data', current_timestamp + interval '1 minute')")
	if err != nil {
		t.Fatal(err)
	}

	p := NewWithCleanupInterval(pool, 0)

	b, found, err := p.FindCtx(ctx, "session_token")
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if bytes.Equal(b, []byte("encoded_data")) == false {
		t.Fatalf("got %v: expected %v", b, []byte("encoded_data"))
	}
}

func TestFindMissing(t *testing.T) {
	ctx := context.Background()

	dsn := os.Getenv("SCS_POSTGRES_TEST_DSN")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, "TRUNCATE TABLE sessions")
	if err != nil {
		t.Fatal(err)
	}

	p := NewWithCleanupInterval(pool, 0)

	_, found, err := p.FindCtx(ctx, "missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestSaveNew(t *testing.T) {
	ctx := context.Background()

	dsn := os.Getenv("SCS_POSTGRES_TEST_DSN")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, "TRUNCATE TABLE sessions")
	if err != nil {
		t.Fatal(err)
	}

	p := NewWithCleanupInterval(pool, 0)

	err = p.CommitCtx(ctx, "session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	row := pool.QueryRow(ctx, "SELECT data FROM sessions WHERE token = 'session_token'")
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
	ctx := context.Background()

	dsn := os.Getenv("SCS_POSTGRES_TEST_DSN")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, "TRUNCATE TABLE sessions")
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, "INSERT INTO sessions VALUES('session_token', 'encoded_data', current_timestamp + interval '1 minute')")
	if err != nil {
		t.Fatal(err)
	}

	p := NewWithCleanupInterval(pool, 0)

	err = p.CommitCtx(ctx, "session_token", []byte("new_encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	row := pool.QueryRow(ctx, "SELECT data FROM sessions WHERE token = 'session_token'")
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
	ctx := context.Background()

	dsn := os.Getenv("SCS_POSTGRES_TEST_DSN")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, "TRUNCATE TABLE sessions")
	if err != nil {
		t.Fatal(err)
	}

	p := NewWithCleanupInterval(pool, 10*time.Millisecond)
	defer p.StopCleanup()

	err = p.CommitCtx(ctx, "session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	_, found, _ := p.FindCtx(ctx, "session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(100 * time.Millisecond)
	_, found, _ = p.FindCtx(ctx, "session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestDelete(t *testing.T) {
	ctx := context.Background()

	dsn := os.Getenv("SCS_POSTGRES_TEST_DSN")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, "TRUNCATE TABLE sessions")
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, "INSERT INTO sessions VALUES('session_token', 'encoded_data', current_timestamp + interval '1 minute')")
	if err != nil {
		t.Fatal(err)
	}

	p := NewWithCleanupInterval(pool, 0)

	err = p.DeleteCtx(ctx, "session_token")
	if err != nil {
		t.Fatal(err)
	}

	row := pool.QueryRow(ctx, "SELECT COUNT(*) FROM sessions WHERE token = 'session_token'")
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
	ctx := context.Background()

	dsn := os.Getenv("SCS_POSTGRES_TEST_DSN")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, "TRUNCATE TABLE sessions")
	if err != nil {
		t.Fatal(err)
	}

	p := NewWithCleanupInterval(pool, 200*time.Millisecond)
	defer p.StopCleanup()

	err = p.CommitCtx(ctx, "session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	row := pool.QueryRow(ctx, "SELECT COUNT(*) FROM sessions WHERE token = 'session_token'")
	var count int
	err = row.Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("got %d: expected %d", count, 1)
	}

	time.Sleep(300 * time.Millisecond)
	row = pool.QueryRow(ctx, "SELECT COUNT(*) FROM sessions WHERE token = 'session_token'")
	err = row.Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("got %d: expected %d", count, 0)
	}
}

func TestStopCleanup(t *testing.T) {
	ctx := context.Background()

	dsn := os.Getenv("SCS_POSTGRES_TEST_DSN")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	for i := 0; i < 100; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			p := New(pool)
			time.Sleep(100 * time.Millisecond)
			defer p.StopCleanup()
		})
	}
}

func TestStopNilCleanup(t *testing.T) {
	dsn := os.Getenv("SCS_POSTGRES_TEST_DSN")
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	p := NewWithCleanupInterval(pool, 0)
	time.Sleep(100 * time.Millisecond)
	// A send to a nil channel will block forever
	p.StopCleanup()
}
