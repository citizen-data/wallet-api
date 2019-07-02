package tenants

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
)

type DynamoTenantStore struct {
	db *dynamodb.DynamoDB
}

type Tenant struct {
	Key string `json:"key"`
	TenantId  string `json:"tenantId"`
	Name string `json:"name"`
}

func NewDynamoTenantStore(db *dynamodb.DynamoDB) *DynamoTenantStore {
	return &DynamoTenantStore{
		db: db,
	}
}

func (t *DynamoTenantStore) GetTenantId(ctx context.Context, apikey string) (string, error) {
	res, err := t.db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("apiKeys"),
		Key: map[string]*dynamodb.AttributeValue{
			"key": {
				S: aws.String(apikey),
			},
		},
	})

	if err != nil {
		log.Fatal(err)
		return "", err
	}

	var tenant Tenant

	err = dynamodbattribute.UnmarshalMap(res.Item, &tenant)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	return tenant.TenantId, nil
}