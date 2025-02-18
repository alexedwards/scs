package dynamodbstore

import (
	"bytes"
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// mockDynamoDBClient is a mock implementation of the DynamoDBClient interface.
type mockDynamoDBClient struct {
	items map[string]map[string]types.AttributeValue
}

func (m *mockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	return &dynamodb.GetItemOutput{
		Item: m.items[params.Key["session_id"].(*types.AttributeValueMemberS).Value],
	}, nil
}

func (m *mockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	if m.items == nil {
		m.items = make(map[string]map[string]types.AttributeValue)
	}

	m.items[params.Item["session_id"].(*types.AttributeValueMemberS).Value] = params.Item
	return &dynamodb.PutItemOutput{}, nil
}

func (m *mockDynamoDBClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	delete(m.items, params.Key["session_id"].(*types.AttributeValueMemberS).Value)
	return &dynamodb.DeleteItemOutput{}, nil
}

func TestNew(t *testing.T) {
	client := &mockDynamoDBClient{}
	table := "test-table"
	store := New(table, client)

	if store.table != table {
		t.Fatalf("got %s: expected %s", store.table, table)
	}
	if store.client != client {
		t.Fatalf("got %v: expected %v", store.client, client)
	}
}

func TestCommit(t *testing.T) {
	client := &mockDynamoDBClient{}
	expiryTime := time.Now().Add(time.Minute)

	store := New("test-table", client)
	store.Commit("key1", []byte("value1"), expiryTime)

	if len(client.items) != 1 {
		t.Fatalf("got %d: expected 1", len(client.items))
	}

	item := client.items["key1"]
	key := item["session_id"].(*types.AttributeValueMemberS).Value
	data := item["data"].(*types.AttributeValueMemberB).Value
	expiry := item["expires_at"].(*types.AttributeValueMemberN).Value

	if key != "key1" {
		t.Fatalf("got %s: expected %s", key, "key1")
	}
	if !bytes.Equal(data, []byte("value1")) {
		t.Fatalf("got %v: expected %v", data, []byte("value1"))
	}
	if expiry != strconv.FormatInt(expiryTime.Unix(), 10) {
		t.Fatalf("got %s: expected %s", expiry, strconv.FormatInt(expiryTime.Unix(), 10))
	}
}

func TestFind(t *testing.T) {
	store := New("test-table", &mockDynamoDBClient{})
	store.Commit("key1", []byte("value1"), time.Now().Add(time.Minute))
	store.Commit("key2", []byte("value2"), time.Now().Add(time.Minute))
	store.Commit("key3", []byte("value3"), time.Now().Add(time.Minute))

	v, found, err := store.Find("key1")
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, false)
	}
	if !bytes.Equal(v, []byte("value1")) {
		t.Fatalf("got %v: expected %v", v, []byte("value1"))
	}

	v, found, err = store.Find("key2")
	if err != nil {
		t.Fatal(err)
	}
	if found != true {
		t.Fatalf("got %v: expected %v", found, false)
	}
	if !bytes.Equal(v, []byte("value2")) {
		t.Fatalf("got %v: expected %v", v, []byte("value2"))
	}
}

func TestDelete(t *testing.T) {
	store := New("test-table", &mockDynamoDBClient{})
	store.Commit("key1", []byte("value1"), time.Now().Add(time.Minute))

	err := store.Delete("key1")
	if err != nil {
		t.Fatal(err)
	}

	_, found, err := store.Find("key1")
	if err != nil {
		t.Fatal(err)
	}
	if found != false {
		t.Fatalf("got %v: expected %v", found, false)
	}
}
