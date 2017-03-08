# dynamostore
[![godoc](https://godoc.org/github.com/alexedwards/scs/engine/dynamostore?status.png)](https://godoc.org/github.com/alexedwards/scs/engine/dynamostore)

Package dynamostore is a Redis-based storage engine for the [SCS session package](https://godoc.org/github.com/alexedwards/scs/session).

## Usage

### Installation

Either:

```
$ go get github.com/alexedwards/scs/engine/dynamostore
```

Or (recommended) use use [gvt](https://github.com/FiloSottile/gvt) to vendor the `engine/dynamostore` and `session` sub-packages:

```
$ gvt fetch github.com/alexedwards/scs/engine/dynamostore
$ gvt fetch github.com/alexedwards/scs/session
```

### Example

The dynamostore package uses the [aws-sdk-go](https://godoc.org/github.com/aws/aws-sdk-go/service/dynamodb) DynamoDB client.


```go
package main

import (
    "io"
    "net/http"

    "github.com/alexedwards/scs/session"
    "github.com/alexedwards/scs/engine/dynamostore"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/endpoints"
    awsSession "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/dynamodb"
)

func main() {
    // Create a DynamoDB client.
    conf := &aws.Config{Region: aws.String(endpoints.UsEast1RegionID)}
    dynamo := dynamodb.New(awsSession.New(), conf)

    // Create a new dynamostore instance using the DynamoDB client.
    engine := dynamostore.New(dynamo)

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

DynamoDB has [Time To Live](http://docs.aws.amazon.com/ja_jp/amazondynamodb/latest/developerguide/TTL.html) option.

Set TTL to DynamoDB session table, then expired session keys automatically removed.

## Notes

Full godoc documentation: [https://godoc.org/github.com/alexedwards/scs/engine/dynamostore](https://godoc.org/github.com/alexedwards/scs/engine/dynamostore).
