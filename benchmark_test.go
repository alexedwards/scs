package scs

import (
	"database/sql"
	"encoding/gob"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/SlinSo/scs/engine/buntstore"
	"github.com/alexedwards/scs/engine/boltstore"
	"github.com/alexedwards/scs/engine/cookiestore"
	"github.com/alexedwards/scs/engine/memstore"
	"github.com/alexedwards/scs/engine/mysqlstore"
	"github.com/alexedwards/scs/engine/pgstore"
	"github.com/alexedwards/scs/engine/redisstore"
	"github.com/alexedwards/scs/session"
	"github.com/boltdb/bolt"
	"github.com/garyburd/redigo/redis"
	"github.com/tidwall/buntdb"
)

func benchSCS(b *testing.B, engine session.Engine) {
	manager := session.Manage(engine)

	setupHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := session.PutInt(r, "counter", 1)
		if err != nil {
			b.Fatal(err)
		}
	})

	benchHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i, err := session.GetInt(r, "counter")
		if err != nil {
			b.Fatal(err)
		}
		err = session.PutInt(r, "counter", i+1)
		if err != nil {
			b.Fatal(err)
		}
	})

	w := httptest.NewRecorder()
	r := new(http.Request)

	manager(setupHandler).ServeHTTP(w, r)
	sessionCookie := w.Header().Get("Set-Cookie")

	r = new(http.Request)
	r.Header = make(http.Header)
	r.Header.Add("Cookie", sessionCookie)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var nr http.Request
		nr = *r
		manager(benchHandler).ServeHTTP(w, &nr)
	}
}

type User struct {
	Name string
	Age  int
}

func benchSCSObject(b *testing.B, engine session.Engine) {
	gob.Register(new(User))

	manager := session.Manage(engine)

	setupHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := session.PutObject(r, "user", &User{"Alex", 33})
		if err != nil {
			b.Fatal(err)
		}
	})

	benchHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := new(User)
		err := session.GetObject(r, "user", user)
		if err != nil {
			b.Fatal(err)
		}
		user.Age = user.Age + 1
		err = session.PutObject(r, "user", user)
		if err != nil {
			b.Fatal(err)
		}
	})

	w := httptest.NewRecorder()
	r := new(http.Request)

	manager(setupHandler).ServeHTTP(w, r)
	sessionCookie := w.Header().Get("Set-Cookie")

	r = new(http.Request)
	r.Header = make(http.Header)
	r.Header.Add("Cookie", sessionCookie)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var nr http.Request
		nr = *r
		manager(benchHandler).ServeHTTP(w, &nr)
	}
}

func BenchmarkSCSMemstore(b *testing.B) {
	benchSCS(b, memstore.New(0))
}

func BenchmarkSCSCookies(b *testing.B) {
	keyset, err := cookiestore.NewKeyset(
		[]byte("f71dc7e58abab014ddad2652475056f185164d262869c8931b239de52711ba87"),
		[]byte("911182cec2f206986c8c82440adb7d17"),
	)
	if err != nil {
		b.Fatal(err)
	}

	benchSCS(b, cookiestore.New(keyset))
}

func BenchmarkSCSRedis(b *testing.B) {
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		conn, err := redis.Dial("tcp", os.Getenv("SESSION_REDIS_TEST_ADDR"))
		if err != nil {
			return nil, err
		}
		return conn, err
	}, 50)
	defer redisPool.Close()

	benchSCS(b, redisstore.New(redisPool))
}

func BenchmarkSCSPostgres(b *testing.B) {
	db, err := sql.Open("postgres", os.Getenv("SESSION_PG_TEST_DSN"))
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	benchSCS(b, pgstore.New(db, 0))
}

func BenchmarkSCSMySQL(b *testing.B) {
	db, err := sql.Open("mysql", os.Getenv("SESSION_MYSQL_TEST_DSN"))
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	benchSCS(b, mysqlstore.New(db, 0))
}

func BenchmarkSCSBoltstore(b *testing.B) {
	db, err := bolt.Open("/tmp/testing.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	benchSCS(b, boltstore.New(db, 0))
}

func BenchmarkSCSBuntstore(b *testing.B) {
	db, err := buntdb.Open("/tmp/bunttesting.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	benchSCS(b, buntstore.New(db))

	b.StopTimer()
	err = db.Update(func(tx *buntdb.Tx) error {
		return tx.DeleteAll()
	})
	if err != nil {
		log.Fatal(err)
	}

	err = db.Shrink()
	if err != nil {
		log.Fatal(err)
	}
}

func BenchmarkSCSObjectMemstore(b *testing.B) {
	benchSCSObject(b, memstore.New(0))
}

func BenchmarkSCSObjectCookies(b *testing.B) {
	keyset, err := cookiestore.NewKeyset(
		[]byte("f71dc7e58abab014ddad2652475056f185164d262869c8931b239de52711ba87"),
		[]byte("911182cec2f206986c8c82440adb7d17"),
	)
	if err != nil {
		b.Fatal(err)
	}

	benchSCSObject(b, cookiestore.New(keyset))
}

func BenchmarkSCSObjectRedis(b *testing.B) {
	redisPool := redis.NewPool(func() (redis.Conn, error) {
		conn, err := redis.Dial("tcp", os.Getenv("SESSION_REDIS_TEST_ADDR"))
		if err != nil {
			return nil, err
		}
		return conn, err
	}, 50)
	defer redisPool.Close()

	benchSCSObject(b, redisstore.New(redisPool))
}

func BenchmarkSCSObjectPostgres(b *testing.B) {
	db, err := sql.Open("postgres", os.Getenv("SESSION_PG_TEST_DSN"))
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	benchSCSObject(b, pgstore.New(db, 0))
}

func BenchmarkSCSObjectMySQL(b *testing.B) {
	db, err := sql.Open("mysql", os.Getenv("SESSION_MYSQL_TEST_DSN"))
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	benchSCSObject(b, mysqlstore.New(db, 0))
}

func BenchmarkSCSObjectBoltstore(b *testing.B) {
	db, err := bolt.Open("/tmp/testing.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	benchSCSObject(b, boltstore.New(db, 0))
}

func BenchmarkSCSObjectBuntstore(b *testing.B) {
	db, err := buntdb.Open("/tmp/bunttesting.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	benchSCSObject(b, buntstore.New(db))

	b.StopTimer()
	err = db.Update(func(tx *buntdb.Tx) error {
		return tx.DeleteAll()
	})
	if err != nil {
		log.Fatal(err)
	}

	err = db.Shrink()
	if err != nil {
		log.Fatal(err)
	}
}
