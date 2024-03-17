// Package xormstore Provide xorm store by xorm
package xormstore

import (
	"bytes"
	"os"
	"reflect"
	"testing"
	"time"

	"xorm.io/xorm"

	_ "github.com/go-sql-driver/mysql"
)

func TestFind(t *testing.T) {
	dsn := os.Getenv("SCS_XORM_TEST_DSN")
	db, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	_, err = db.Truncate(&session{})
	if err != nil {
		t.Fatal(err)
	}

	ts := &session{
		Token:  "scs_test_token",
		Data:   []byte("scs_test_data"),
		Expiry: time.Now().Add(time.Minute),
	}

	_, err = db.Insert(ts)
	if err != nil {
		t.Fatal(err)
	}

	xs, err := NewWithCleanupInterval(db, 0)
	if err != nil {
		t.Fatal(err)
	}

	b, found, err := xs.Find("scs_test_token")
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if bytes.Equal(b, []byte("scs_test_data")) == false {
		t.Fatalf("got %v: expected %v", b, []byte("scs_test_data"))
	}
}

func TestFindMissing(t *testing.T) {
	dsn := os.Getenv("SCS_XORM_TEST_DSN")
	db, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	_, err = db.Truncate(&session{})
	if err != nil {
		t.Fatal(err)
	}

	xs, err := NewWithCleanupInterval(db, 0)
	if err != nil {
		t.Fatal(err)
	}

	_, found, err := xs.Find("scs_test_token")
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestSaveNew(t *testing.T) {
	dsn := os.Getenv("SCS_XORM_TEST_DSN")
	db, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	_, err = db.Truncate(&session{})
	if err != nil {
		t.Fatal(err)
	}

	ts := &session{
		Token:  "scs_test_token",
		Data:   []byte("scs_test_data"),
		Expiry: time.Now().Add(time.Minute),
	}

	_, err = db.Insert(ts)
	if err != nil {
		t.Fatal(err)
	}

	xs, err := NewWithCleanupInterval(db, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := xs.Commit("scs_test_token", []byte("scs_test_data"), time.Now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}

	s := &session{}
	has, err := db.Where("token = ?", "scs_test_token").Get(s)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatalf("got %v: expected %v", has, true)
	}

	if reflect.DeepEqual(s.Data, []byte("scs_test_data")) == false {
		t.Fatalf("got %v: expected %v", s.Data, []byte("scs_test_data"))
	}
}

func TestSaveUpdate(t *testing.T) {
	dsn := os.Getenv("SCS_XORM_TEST_DSN")
	db, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	_, err = db.Truncate(&session{})
	if err != nil {
		t.Fatal(err)
	}

	ts := &session{
		Token:  "scs_test_token",
		Data:   []byte("scs_test_data"),
		Expiry: time.Now().Add(time.Minute),
	}
	_, err = db.Insert(ts)
	if err != nil {
		t.Fatal(err)
	}

	xs, err := NewWithCleanupInterval(db, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := xs.Commit("scs_test_token", []byte("scs_test_data_new"), time.Now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}

	s := &session{}
	has, err := db.Where("token = ?", "scs_test_token").Get(s)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatalf("got %v: expected %v", has, true)
	}

	if reflect.DeepEqual(s.Data, []byte("scs_test_data_new")) == false {
		t.Fatalf("got %v: expected %v", s.Data, []byte("scs_test_data_new"))
	}
}

func TestSaveExpiry(t *testing.T) {
	dsn := os.Getenv("SCS_XORM_TEST_DSN")
	db, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	_, err = db.Truncate(&session{})
	if err != nil {
		t.Fatal(err)
	}

	xs, err := NewWithCleanupInterval(db, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := xs.Commit("scs_test_token", []byte("scs_test_data"), time.Now().Add(300*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	_, found, _ := xs.Find("scs_test_token")
	if !found {
		t.Fatalf("got %v: expected %v", found, true)
	}

	time.Sleep(300 * time.Millisecond)
	_, found, _ = xs.Find("scs_test_token")
	if !found {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestDelete(t *testing.T) {
	dsn := os.Getenv("SCS_XORM_TEST_DSN")
	db, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	_, err = db.Truncate(&session{})
	if err != nil {
		t.Fatal(err)
	}

	ts := &session{
		Token:  "scs_test_token",
		Data:   []byte("scs_test_data"),
		Expiry: time.Now().Add(time.Minute),
	}
	_, err = db.Insert(ts)
	if err != nil {
		t.Fatal(err)
	}

	xs, err := NewWithCleanupInterval(db, 0)
	if err != nil {
		t.Fatal(err)
	}

	if err := xs.Delete("scs_test_token"); err != nil {
		t.Fatal(err)
	}

	total, err := db.Count(&session{})
	if err != nil {
		t.Fatal(err)
	}
	if total != 0 {
		t.Fatalf("got %v: expected %v", total, 0)
	}
}

func TestCleanup(t *testing.T) {
	dsn := os.Getenv("SCS_XORM_TEST_DSN")
	db, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	_, err = db.Truncate(&session{})
	if err != nil {
		t.Fatal(err)
	}

	xs, err := NewWithCleanupInterval(db, 200*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer xs.StopCleanup()

	if err := xs.Commit("scs_test_token", []byte("scs_test_data"), time.Now().Add(100*time.Millisecond)); err != nil {
		t.Fatal(err)
	}

	total, err := db.Count(&session{})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Fatalf("got %v: expected %v", total, 1)
	}

	time.Sleep(500 * time.Millisecond)
	total, err = db.Count(&session{})
	if err != nil {
		t.Fatal(err)
	}
	if total != 0 {
		t.Fatalf("got %v: expected %v", total, 0)
	}
}

func TestStopNilCleanup(t *testing.T) {
	dsn := os.Getenv("SCS_XORM_TEST_DSN")
	db, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	xs, err := NewWithCleanupInterval(db, 0)
	if err != nil {
		t.Fatal(err)
	}

	xs.StopCleanup()
}
