package wallets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	walletTable = "wallets"
	dataTable = "wallet-data"
	bucket      = "data-wallet-storage"
)

type AWSWalletStore struct {
	db *dynamodb.DynamoDB
	s3 *s3.S3
}

type DynamoWallet struct {
	WalletID string  `json:"walletId"`
	Wallet   *Wallet `json:"wallet"`
	TenantID string  `json:"tenantId"`
}

type DynamoWalletData struct {
	WalletID  string                `json:"walletId"`
	ObjectKey string                `json:"objectKey"`
	Summary   *WalletDataItemSummary `json:"summary"`
}

func NewAWSWalletStore(db *dynamodb.DynamoDB, s3 *s3.S3) *AWSWalletStore {
	return &AWSWalletStore{
		db: db,
		s3: s3,
	}
}

func (s *AWSWalletStore) CreateWallet(ctx context.Context, wallet *Wallet) error {
	item, err := dynamodbattribute.MarshalMap(&DynamoWallet{
		WalletID: fmt.Sprintf("%s/%s", wallet.TenantID, wallet.WalletID),
		Wallet:   wallet,
		TenantID: wallet.TenantID,
	})

	if err != nil {
		return err
	}

	_, err = s.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(walletTable),
		Item:      item,
	})

	return err
}

func (s *AWSWalletStore) GetWallet(ctx context.Context, tenantID, walletID string) (*Wallet, error) {
	res, err := s.db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(walletTable),
		Key: map[string]*dynamodb.AttributeValue{
			"walletId": {
				S: aws.String(fmt.Sprintf("%s/%s", tenantID, walletID)),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var wallet DynamoWallet

	err = dynamodbattribute.UnmarshalMap(res.Item, &wallet)
	if err != nil {
		return nil, err
	}

	return wallet.Wallet, nil
}

func (s *AWSWalletStore) AddDataItem(ctx context.Context, tenantID, walletID string, data *WalletDataItem) error {
	objectKey := fmt.Sprintf("%s/%s/%s/%s", tenantID, walletID, data.ReferenceID, data.CreatedAt)

	item, err := dynamodbattribute.MarshalMap(&DynamoWalletData{
		WalletID: fmt.Sprintf("%s/%s", tenantID, walletID),
		ObjectKey: objectKey,
		Summary: &WalletDataItemSummary{
			DataSignature: data.DataSignature,
			ReferenceID: data.ReferenceID,
			CreatedAt: data.CreatedAt,
		},
	})

	if err != nil {
		return err
	}

	_, err = s.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(dataTable),
		Item:      item,
	})

	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = s.s3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(b),
	})

	if err != nil {
		return err
	}

	return err
}

func calcWalletID(tenantID, walletID string) string {
	return fmt.Sprintf("%s/%s", tenantID, walletID)
}

func (s *AWSWalletStore) ListData(ctx context.Context, tenantID, walletID string) (*WalletList, error) {
	itemMap := make(map[string][]*WalletDataItemSummary)

	key := expression.Key("walletId").Equal(expression.Value(calcWalletID(tenantID, walletID)))

	expr, err := expression.NewBuilder().WithKeyCondition(key).Build()
	if err != nil {
		return nil, err
	}

	err = s.db.QueryPagesWithContext(ctx, &dynamodb.QueryInput{
		TableName: aws.String(dataTable),
		KeyConditionExpression: expr.KeyCondition(),
		ExpressionAttributeNames: expr.Names(),
		ExpressionAttributeValues:expr.Values(),
	},
		func(page *dynamodb.QueryOutput, lastPage bool) bool {
			for _, item := range page.Items {
				var dwd DynamoWalletData
				err = dynamodbattribute.UnmarshalMap(item, &dwd)
				if err != nil {
					//TODO: handle this
					continue
				}
				if _, ok := itemMap[dwd.Summary.ReferenceID]; !ok {
					itemMap[dwd.Summary.ReferenceID] = make([]*WalletDataItemSummary, 0)
				}
				itemMap[dwd.Summary.ReferenceID] = append(itemMap[dwd.Summary.ReferenceID], dwd.Summary)
			}
			return lastPage
		})

	if err != nil {
		return nil, err
	}

	return &WalletList{
		Items: itemMap,
	}, nil
}
