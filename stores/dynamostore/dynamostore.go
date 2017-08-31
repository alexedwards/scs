// Package dynamostore is a DynamoDB-based session store for the SCS session package.
//
// The dynamostore package relis on the aws-sdk-go client.
// (https://godoc.org/github.com/aws/aws-sdk-go/service/dynamodb)
package dynamostore

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// DynamoStore represents the currently configured session session store. It is essentially
// a wrapper around a DynamoDB client. And table is a table name session stored. token, data,
// expiry are key names.
type DynamoStore struct {
	DB     *dynamodb.DynamoDB
	table  string
	token  string
	data   string
	expiry string
	ttl    string
}

const (
	defaultTable  = "scs_session"
	defaultToken  = "token"
	defaultData   = "data"
	defaultExpiry = "expiry"
	defaultTTL    = "ttl"
)

// New returns a new DynamoStore instance. The client parameter shoud be a pointer to a
// aws-sdk-go DynamoDB client. See https://godoc.org/github.com/aws/aws-sdk-go/service/dynamodb#DynamoDB.
func New(dynamo *dynamodb.DynamoDB) *DynamoStore {
	return NewWithOption(dynamo, defaultTable, defaultToken, defaultData, defaultExpiry, defaultTTL)
}

// NewWithOption returns a new DynamoStore instance. The client parameter shoud be a pointer to a
// aws-sdk-go DynamoDB client. See https://godoc.org/github.com/aws/aws-sdk-go/service/dynamodb#DynamoDB.
// The parameter table is DynamoDB tabel name, and token/data/expiry are key names.
func NewWithOption(dynamo *dynamodb.DynamoDB, table string, token string, data string, expiry string, ttl string) *DynamoStore {
	return &DynamoStore{
		DB:     dynamo,
		table:  table,
		token:  token,
		data:   data,
		expiry: expiry,
		ttl:    ttl,
	}
}

// Find returns the data for a given session token from the DynamoStore instance. If the session
// token is not found or is expired, the returned exists flag will be set to false.
func (d *DynamoStore) Find(token string) (b []byte, found bool, err error) {
	params := &dynamodb.GetItemInput{
		TableName: aws.String(d.TableName()),
		Key: map[string]*dynamodb.AttributeValue{
			d.TokenName(): {
				S: aws.String(token),
			},
		},
		ConsistentRead: aws.Bool(true),
	}

	resp, err := d.DB.GetItem(params)
	if err != nil {
		return nil, false, err
	}
	if resp.Item == nil {
		return nil, false, nil
	}

	expiry, err := strconv.ParseInt(aws.StringValue(resp.Item[d.ExpiryName()].N), 10, 64)
	if err != nil {
		return nil, false, err
	}

	if expiry < time.Now().UnixNano() {
		return nil, false, d.Delete(token)
	}

	return resp.Item[d.DataName()].B, true, nil
}

// Save adds a session token and data to the RedisStore instance with the given expiry time.
// If the session token already exists then the data and expiry time are updated.
func (d *DynamoStore) Save(token string, b []byte, expiry time.Time) error {
	params := &dynamodb.PutItemInput{
		TableName: aws.String(d.TableName()),
		Item: map[string]*dynamodb.AttributeValue{
			d.TokenName(): {
				S: aws.String(token),
			},
			d.DataName(): {
				B: b,
			},
			d.ExpiryName(): {
				N: aws.String(strconv.FormatInt(expiry.UnixNano(), 10)),
			},
			d.TTLName(): {
				// TTL is used by DynamoDB Time To Live. It must be Unix Epoch format.
				// TTL cannot handle under second like milliseocnd and nanosecond, but
				// Expiry can.
				N: aws.String(strconv.FormatInt(expiry.Add(1*time.Second).Unix(), 10)),
			},
		},
	}
	_, err := d.DB.PutItem(params)
	return err
}

// Delete removes a session token and corresponding data from the ResisStore instance.
func (d *DynamoStore) Delete(token string) error {
	params := &dynamodb.DeleteItemInput{
		TableName: aws.String(d.TableName()),
		Key: map[string]*dynamodb.AttributeValue{
			d.TokenName(): {
				S: aws.String(token),
			},
		},
	}

	_, err := d.DB.DeleteItem(params)
	return err
}

// Ping checks to exisit session table in DynamoDB.
func (d *DynamoStore) Ping() error {
	params := &dynamodb.DescribeTableInput{
		TableName: aws.String(d.TableName()),
	}

	_, err := d.DB.DescribeTable(params)
	return err
}

// TableName returns session table name.
func (d *DynamoStore) TableName() string {
	return d.table
}

// TokenName returns session token key name.
func (d *DynamoStore) TokenName() string {
	return d.token
}

// DataName returns session data key name.
func (d *DynamoStore) DataName() string {
	return d.data
}

// ExpiryName returns session expiry key name.
func (d *DynamoStore) ExpiryName() string {
	return d.expiry
}

// TTLName returns session expiry key name.
func (d *DynamoStore) TTLName() string {
	return d.ttl
}
