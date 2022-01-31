# etcdstore

A [etcd](https://github.com/etcd-io/etcd) based session store for [SCS](https://github.com/alexedwards/scs).

## Setup

You should follow the instructions to [setup a connection](https://github.com/etcd-io/etcd/tree/main/client/v3#install), and pass the connection to `etcdstore.New()` to establish the session store.

## Example

```go
package main

import (
	"io"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/etcdstore"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var sessionManager *scs.SessionManager

func main() {
	// Establish connection to etcd.
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"host:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()
	
	// Initialize a new session manager and configure it to use etcdstore as the session store.
	sessionManager = scs.New()
	sessionManager.Store = etcdstore.New(cli)

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

Etcd will [automatically remove](https://etcd.io/docs/v3.5/tutorials/how-to-create-lease/) expired session keys.

## Key Collisions

By default keys are in the form `scs:session:<token>`. For example:

```
"scs:session:ZnirGwi2FiLwXeVlP5nD77IpfJZMVr6un9oZu2qtJrg"
```

Because the token is highly unique, key collisions are not a concern. But if you're configuring *multiple session managers*, both of which use `etcdstore`, then you may want the keys to have a different prefix depending on which session manager wrote them. You can do this by using the `NewWithPrefix()` method like so:

```go
cli, err := clientv3.New(clientv3.Config{
	Endpoints:   []string{"host:2379"},
	DialTimeout: 5 * time.Second,
})
if err != nil {
	log.Fatal(err)
}
defer cli.Close()

sessionManagerOne = scs.New()
sessionManagerOne.Store = etcdstore.NewWithPrefix(cli, "scs:session:1:")

sessionManagerTwo = scs.New()
sessionManagerTwo.Store = etcdstore.NewWithPrefix(cli, "scs:session:2:")
```
