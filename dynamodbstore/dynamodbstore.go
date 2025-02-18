package dynamodbstore

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBClient interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

// DynamoDBStore represents a DynamoDB backed session store.
type DynamoDBStore struct {
	table  string
	client DynamoDBClient
}

// New returns a new DynamoDBStore instance with the given table name and DynamoDB client.
func New(table string, client DynamoDBClient) *DynamoDBStore {
	return &DynamoDBStore{
		table:  table,
		client: client,
	}
}

// Commit behaves the same as CommitCtx but with a background context.
func (d *DynamoDBStore) Commit(token string, b []byte, expiry time.Time) error {
	return d.CommitCtx(context.Background(), token, b, expiry)
}

// CommitCtx adds a session token and data to the DynamoDB store with the
// given expiry time. The session will be created if it does not exist, and updated if it already exists.
func (d *DynamoDBStore) CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) error {
	item := map[string]types.AttributeValue{
		"session_id": &types.AttributeValueMemberS{
			Value: token,
		},
		"data": &types.AttributeValueMemberB{
			Value: b,
		},
		"updated_at": &types.AttributeValueMemberN{
			Value: strconv.FormatInt(time.Now().Unix(), 10),
		},
		"expires_at": &types.AttributeValueMemberN{
			Value: strconv.FormatInt(expiry.Unix(), 10),
		},
	}

	_, err := d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &d.table,
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("unable to put item in dynamodb: %w", err)
	}

	return nil
}

// Delete behaves the same as DeleteCtx but with a background context.
func (d *DynamoDBStore) Delete(token string) error {
	return d.DeleteCtx(context.Background(), token)
}

// DeleteCtx removes a session token and corresponding data from the DynamoDBStore instance.
func (d *DynamoDBStore) DeleteCtx(ctx context.Context, token string) error {
	_, err := d.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &d.table,
		Key: map[string]types.AttributeValue{
			"session_id": &types.AttributeValueMemberS{
				Value: token,
			},
		},
	})

	if err != nil {
		return fmt.Errorf("unable to delete item in dynamodb: %w", err)
	}

	return nil
}

// Find behaves the same as FindCtx but with a background context.
func (d *DynamoDBStore) Find(token string) ([]byte, bool, error) {
	return d.FindCtx(context.Background(), token)
}

// FindCtx returns the data for a given session token from the DynamoDBStore instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (d *DynamoDBStore) FindCtx(ctx context.Context, token string) ([]byte, bool, error) {
	res, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &d.table,
		Key: map[string]types.AttributeValue{
			"session_id": &types.AttributeValueMemberS{
				Value: token,
			},
		},
	})

	if err != nil {
		return nil, false, fmt.Errorf("unable to get item in dynamodb: %w", err)
	}

	if res.Item == nil {
		return nil, false, nil
	}

	// Validate we haven't passed the expiry time (TTL). Given the automatic deletion of items by DynamoDB is asynchronous,
	// and does not happen immediately at the point of expiry, we may end up retrieving items that have passed their expiry time.
	var expiry int64
	if err = attributevalue.Unmarshal(res.Item["expires_at"], &expiry); err != nil {
		return nil, false, fmt.Errorf("unable to unmarshal expires_at attribute value: %w", err)
	}

	if time.Now().Unix() > expiry {
		// Delete the item from the store to avoid retrieving it again.
		if err = d.DeleteCtx(ctx, token); err != nil {
			return nil, false, fmt.Errorf("unable to delete expired item: %w", err)
		}

		return nil, false, nil
	}

	var data []byte
	if err = attributevalue.Unmarshal(res.Item["data"], &data); err != nil {
		return nil, false, fmt.Errorf("unable to unmarshal data attribute value: %w", err)
	}

	return data, true, nil
}
