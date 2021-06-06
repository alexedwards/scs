package gormstore

import (
	"bytes"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func initWithCleanupInterval(t *testing.T, cleanupInterval time.Duration) (*GORMStore, *gorm.DB) {
	dialect := os.Getenv("SCS_GORM_TEST_DIALECT")
	var dsn string
	switch dialect {
	case "postgres":
		dsn = os.Getenv("SCS_POSTGRES_TEST_DSN")
	case "mysql":
		dsn = os.Getenv("SCS_MYSQL_TEST_DSN")
	default:
		dialect = "sqlite3"
		fallthrough
	case "sqlite3":
		dsn = os.Getenv("./testSQL3lite.db")
	}

	db, err := gorm.Open(dialect, dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.DB().Ping(); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if _, err := db.DB().Exec("DROP TABLE IF EXISTS sessions"); err != nil {
		t.Fatal(err)
	}

	p, err := NewWithCleanupInterval(db, cleanupInterval)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}

	return p, db
}

func TestFind(t *testing.T) {
	p, db := initWithCleanupInterval(t, 0)
	defer db.Close()

	sess := db.Create(&session{Token: "session_token", Data: []byte("encoded_data"), Expiry: time.Now().Add(1 * time.Minute)})
	if errs := sess.GetErrors(); len(errs) != 0 {
		t.Fatal(errs[0])
	}

	b, found, err := p.Find("session_token")
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
	p, db := initWithCleanupInterval(t, 0)
	defer db.Close()

	_, found, err := p.Find("missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestSaveNew(t *testing.T) {
	p, db := initWithCleanupInterval(t, 0)
	defer db.Close()

	err := p.Commit("session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	row := db.DB().QueryRow("SELECT data FROM sessions WHERE token = 'session_token'")
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
	p, db := initWithCleanupInterval(t, 0)
	defer db.Close()

	sess := db.Create(&session{Token: "session_token", Data: []byte("encoded_data"), Expiry: time.Now().Add(1 * time.Minute)})
	if errs := sess.GetErrors(); len(errs) != 0 {
		t.Fatal(errs[0])
	}

	err := p.Commit("session_token", []byte("new_encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	row := db.DB().QueryRow("SELECT data FROM sessions WHERE token = 'session_token'")
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
	p, db := initWithCleanupInterval(t, 0)
	defer db.Close()

	err := p.Commit("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	_, found, _ := p.Find("session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(100 * time.Millisecond)
	_, found, _ = p.Find("session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestDelete(t *testing.T) {
	p, db := initWithCleanupInterval(t, 0)
	defer db.Close()

	err := p.Commit("session_token", []byte("encoded_data"), time.Now().Add(1*time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	err = p.Delete("session_token")
	if err != nil {
		t.Fatal(err)
	}

	row := db.DB().QueryRow("SELECT COUNT(*) FROM sessions WHERE token = 'session_token'")
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
	p, db := initWithCleanupInterval(t, 200*time.Millisecond)
	defer p.StopCleanup()
	defer db.Close()

	err := p.Commit("session_token", []byte("encoded_data"), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	row := db.DB().QueryRow("SELECT COUNT(*) FROM sessions WHERE token = 'session_token'")
	var count int
	err = row.Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("got %d: expected %d", count, 1)
	}

	time.Sleep(300 * time.Millisecond)
	row = db.DB().QueryRow("SELECT COUNT(*) FROM sessions WHERE token = 'session_token'")
	err = row.Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("got %d: expected %d", count, 0)
	}
}

func TestStopNilCleanup(t *testing.T) {
	p, db := initWithCleanupInterval(t, 0)
	defer db.Close()

	time.Sleep(100 * time.Millisecond)
	// A send to a nil channel will block forever
	p.StopCleanup()
}
