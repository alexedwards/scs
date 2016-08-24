# redisstore 
[![godoc](https://godoc.org/github.com/alexedwards/scs/engine/redisstore?status.png)](https://godoc.org/github.com/alexedwards/scs/engine/redisstore)

Package redisstore is a Redis-based storage engine for the [SCS session package](https://godoc.org/github.com/alexedwards/scs/session).

Warning: The redisstore API is not finalized and may change, possibly significantly. The package is fine to use as-is, but it is strongly recommended that you vendor the package to avoid compatibility problems in the future.

## Usage

### Installation

Either:

```
$ go get github.com/alexedwards/scs/engine/redisstore
```

Or (recommended) use use [gvt](https://github.com/FiloSottile/gvt) to vendor the `engine/redisstore` and `session` sub-packages:

```
$ gvt fetch github.com/alexedwards/scs/engine/redisstore
$ gvt fetch github.com/alexedwards/scs/session
```

### Example

The redisstore package uses the popular [Redigo](https://github.com/garyburd/redigo) Redis client.

You should follow the Redigo instructions to [setup a connection pool](https://godoc.org/github.com/garyburd/redigo/redis#Pool), and pass the pool to `redisstore.New()` to establish the session storage engine.

```go
package main

import (
    "io"
    "net/http"

    "github.com/alexedwards/scs/engine/redisstore"
    "github.com/alexedwards/scs/session"
    "github.com/garyburd/redigo/redis"
)

func main() {
    // Establish a Redigo connection pool.
    pool := &redis.Pool{
        MaxIdle: 10,
        Dial: func() (redis.Conn, error) {
            return redis.Dial("tcp", "localhost:6379")
        },
    }

    // Create a new RedisStore instance using the connection pool.
    engine := redisstore.New(pool)

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

Redis will [automatically remove](http://redis.io/commands/expire#how-redis-expires-keys) expired session keys. 

## Notes

Full godoc documentation: [https://godoc.org/github.com/alexedwards/scs/engine/redisstore](https://godoc.org/github.com/alexedwards/scs/engine/redisstore).