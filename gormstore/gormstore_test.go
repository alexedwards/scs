package gormstore

import (
	"bytes"
	"os"
	"reflect"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func initWithCleanupInterval(t *testing.T, cleanupInterval time.Duration) (*GORMStore, *gorm.DB) {
	var db *gorm.DB
	var err error

	dialect := os.Getenv("SCS_GORM_TEST_DIALECT")
	switch dialect {
	case "postgres":
		dsn := os.Getenv("SCS_POSTGRES_TEST_DSN")
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	case "mssql":
		dsn := os.Getenv("SCS_MSSQL_TEST_DSN")
		db, err = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
	case "mysql":
		dsn := os.Getenv("SCS_MYSQL_TEST_DSN")
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	default:
		dsn := "./testSQL3lite.db"
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	}
	if err != nil {
		t.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		sqlDB.Close()
		t.Fatal(err)
	}
	if err = sqlDB.Ping(); err != nil {
		sqlDB.Close()
		t.Fatal(err)
	}

	if dialect == "mssql" {
		if err := db.Exec("IF OBJECT_ID('sessions', 'U') IS NOT NULL DROP TABLE sessions").Error; err != nil {
			t.Fatal(err)
		}
	} else {
		if err := db.Exec("DROP TABLE IF EXISTS sessions").Error; err != nil {
			t.Fatal(err)
		}
	}

	g, err := NewWithCleanupInterval(db, cleanupInterval)
	if err != nil {
		sqlDB.Close()
		t.Fatal(err)
	}

	return g, db
}

func TestFind(t *testing.T) {
	g, db := initWithCleanupInterval(t, 0)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	row := db.Create(&session{Token: "session_token", Data: []byte("encoded_data"), Expiry: time.Now().Add(1 * time.Minute)})
	if row.Error != nil {
		t.Fatal(err)
	}

	b, found, err := g.Find("session_token")
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
	g, db := initWithCleanupInterval(t, 0)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	_, found, err := g.Find("missing_session_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestSaveNew(t *testing.T) {
	g, db := initWithCleanupInterval(t, 0)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	err = g.Commit("session_token", []byte("encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	row := sqlDB.QueryRow("SELECT data FROM sessions WHERE token = 'session_token'")
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
	g, db := initWithCleanupInterval(t, 0)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	row1 := db.Create(&session{Token: "session_token", Data: []byte("encoded_data"), Expiry: time.Now().Add(1 * time.Minute)})
	if row1.Error != nil {
		t.Fatal(row1.Error)
	}

	err = g.Commit("session_token", []byte("new_encoded_data"), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	row2 := sqlDB.QueryRow("SELECT data FROM sessions WHERE token = 'session_token'")
	var data []byte
	err = row2.Scan(&data)
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(data, []byte("new_encoded_data")) == false {
		t.Fatalf("got %v: expected %v", data, []byte("new_encoded_data"))
	}
}

func TestExpiry(t *testing.T) {
	g, db := initWithCleanupInterval(t, 0)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	err = g.Commit("session_token", []byte("encoded_data"), time.Now().Add(1*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	_, found, _ := g.Find("session_token")
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(2 * time.Second)
	_, found, _ = g.Find("session_token")
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestDelete(t *testing.T) {
	g, db := initWithCleanupInterval(t, 0)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	err = g.Commit("session_token", []byte("encoded_data"), time.Now().Add(1*time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	err = g.Delete("session_token")
	if err != nil {
		t.Fatal(err)
	}

	row := sqlDB.QueryRow("SELECT COUNT(*) FROM sessions WHERE token = 'session_token'")
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
	g, db := initWithCleanupInterval(t, 2*time.Second)
	defer g.StopCleanup()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	err = g.Commit("session_token", []byte("encoded_data"), time.Now().Add(1*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	row := sqlDB.QueryRow("SELECT COUNT(*) FROM sessions WHERE token = 'session_token'")
	var count int
	err = row.Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("got %d: expected %d", count, 1)
	}

	time.Sleep(3 * time.Second)
	row = sqlDB.QueryRow("SELECT COUNT(*) FROM sessions WHERE token = 'session_token'")
	err = row.Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("got %d: expected %d", count, 0)
	}
}

func TestStopNilCleanup(t *testing.T) {
	g, db := initWithCleanupInterval(t, 0)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	time.Sleep(100 * time.Millisecond)
	// A send to a nil channel will block forever
	g.StopCleanup()
}
