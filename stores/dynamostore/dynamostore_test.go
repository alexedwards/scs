package dynamostore

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const (
	defaultRegion = endpoints.ApNortheast1RegionID
	token         = "session_token"
	data          = "encoded_data"
	dataUpdated   = "encoded_data_updated"
)

func getTestDynamoDB(t *testing.T) *dynamodb.DynamoDB {
	conf := &aws.Config{Region: aws.String(defaultRegion)}
	sess, err := awsSession.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	dy := dynamodb.New(sess, conf)
	if dy == nil {
		t.Fatal("failed to create dynamodb client")
	}

	d := New(dy)
	err = d.Ping()
	if err != nil {
		t.Fatal(err)
	}

	return dy
}

func clearTestDynamoDB(t *testing.T, dy *dynamodb.DynamoDB) {
	d := New(dy)
	_, found, err := d.Find(token)
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		return
	}
	err = d.Delete(token)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFind(t *testing.T) {
	dy := getTestDynamoDB(t)
	clearTestDynamoDB(t, dy)

	d := New(dy)

	err := d.Save(token, []byte(data), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	b, found, err := d.Find(token)
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if bytes.Equal(b, []byte(data)) == false {
		t.Fatalf("got %v: expected %v", b, []byte(data))
	}
}

func TestFindMissing(t *testing.T) {
	dy := getTestDynamoDB(t)
	clearTestDynamoDB(t, dy)

	d := New(dy)
	_, found, err := d.Find(token)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatalf("got %v: expected %v", err, nil)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}

func TestSaveNew(t *testing.T) {
	dy := getTestDynamoDB(t)
	clearTestDynamoDB(t, dy)

	d := New(dy)
	err := d.Save(token, []byte(data), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	b, found, err := d.Find(token)
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if reflect.DeepEqual(b, []byte(data)) != true {
		t.Fatalf("got %v: expected %v", b, []byte(data))
	}
}

func TestSaveUpdated(t *testing.T) {
	dy := getTestDynamoDB(t)
	clearTestDynamoDB(t, dy)

	d := New(dy)
	err := d.Save(token, []byte(data), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = d.Find(token)
	if err != nil {
		t.Fatal(err)
	}

	err = d.Save(token, []byte(dataUpdated), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}

	b, found, err := d.Find(token)
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, true)
	}
	if reflect.DeepEqual(b, []byte(dataUpdated)) != true {
		t.Fatalf("got %v: expected %v", b, []byte(dataUpdated))
	}
}

func TestExpiry(t *testing.T) {
	dy := getTestDynamoDB(t)
	clearTestDynamoDB(t, dy)

	d := New(dy)

	err := d.Save(token, []byte(dataUpdated), time.Now().Add(100*time.Millisecond))
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)

	_, found, _ := d.Find(token)
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}
