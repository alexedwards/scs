# DynamoDB Store

An AWS DynamoDB based session store for [SCS](https://github.com/alexedwards/scs).

## Setup

You must have an existing DynamoDB table with the following attribute definition, key schema and time to live specification:

*Example CloudFormation resource definition:*
```yaml
SessionStoreTable:
  Type: AWS::DynamoDB::Table
  Properties:
    TableName: my-session-store
    BillingMode: PAY_PER_REQUEST
    AttributeDefinitions:
      - AttributeName: session_id
        AttributeType: S
    KeySchema:
      - AttributeName: session_id
        KeyType: HASH
    TimeToLiveSpecification:
      AttributeName: expires_at
      Enabled: true
```

## Example

```go
package main

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/alexedwards/scs/dynamodbstore"
	"github.com/alexedwards/scs/v2"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var sessionManager *scs.SessionManager

func main() {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}

	dynamodbClient := dynamodb.NewFromConfig(cfg)

	sessionManager = scs.New()
	sessionManager.Store = dynamodbstore.New("dynamodb-table-name", dynamodbClient)

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

AWS DynamoDB provides automatic deletion of items once the configured TTL has been reached. See [docs](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/TTL.html). If you do not wish to enable the automatic TTL management, the DynamoDBStore will also delete expired items when fetched. However, this does not cater for the automatic deletion of data that is not fetched again. Using the TTL mechanism provided by AWS is recommended to ensure expired data is cleaned up automatically.
