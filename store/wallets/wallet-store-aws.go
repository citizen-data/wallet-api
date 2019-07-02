package wallets

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go/service/s3"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
)

const (
	walletTable  = "wallets"
	dataTable    = "wallet-data"
	dataRefIndex = "referenceId-createdAt-index"
	bucket       = "data-wallet-storage"
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
	WalletID    string                 `json:"walletId"`
	ObjectKey   string                 `json:"objectKey"`
	Summary     *WalletDataItemSummary `json:"summary"`
	ReferenceID string                 `json:"referenceId"`
	CreatedAt   string                 `json:"createdAt"`
	VersionHash    string                 `json:"versionHash"`
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
	objectKey := fmt.Sprintf("%s/%s/%s/%s", tenantID, walletID, data.ReferenceID, data.VersionHash)

	item, err := dynamodbattribute.MarshalMap(&DynamoWalletData{
		WalletID:  fmt.Sprintf("%s/%s", tenantID, walletID),
		ObjectKey: objectKey,
		Summary: &WalletDataItemSummary{
			DataSignature: data.DataSignature,
			ReferenceID:   data.ReferenceID,
			CreatedAt:     data.CreatedAt,
			VersionHash:      data.VersionHash,
		},
		VersionHash:      data.VersionHash,
		CreatedAt:   data.CreatedAt,
		ReferenceID: fmt.Sprintf("%s/%s/%s", tenantID, walletID, data.ReferenceID),
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

type walletS3Obj struct {
	objectKey string
	data      *WalletDataItem
}

func (s *AWSWalletStore) getObject(objectKey string) (*WalletDataItem, error) {
	obj, err := s.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
	})

	if err != nil {
		return nil, err
	}
	defer obj.Body.Close()

	var data WalletDataItem

	b, err := ioutil.ReadAll(obj.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (s *AWSWalletStore) getObjects(objectKeys []string) ([]*WalletDataItem, error) {
	var g errgroup.Group
	resultsMap := make(map[string]*WalletDataItem)
	ch := make(chan *walletS3Obj, len(objectKeys))

	for _, objKey := range objectKeys {
		objectKey := objKey
		g.Go(func() error {
			data, err := s.getObject(objectKey)
			if err != nil {
				return err
			}
			ch <- &walletS3Obj{
				objectKey: objectKey,
				data:      data,
			}
			return nil
		})
	}

	var err error
	go func() {
		err = g.Wait()
		close(ch)
	}()

	for obj := range ch {
		resultsMap[obj.objectKey] = obj.data
	}

	if err != nil {
		return nil, err
	}

	results := make([]*WalletDataItem, len(objectKeys), 0)
	for _, objKey := range objectKeys {
		if obj, ok := resultsMap[objKey]; ok {
			results = append(results, obj)
		}
	}

	return results, nil
}

func (s *AWSWalletStore) GetLatestDataItem(ctx context.Context, tenantID, walletID, referenceId string) (*WalletDataItem, error) {
	refID := fmt.Sprintf("%s/%s/%s", tenantID, walletID, referenceId)
	key := expression.Key("referenceId").Equal(expression.Value(refID))
	proj := expression.NamesList(expression.Name("objectKey"))
	expr, err := expression.NewBuilder().WithKeyCondition(key).WithProjection(proj).Build()
	if err != nil {
		return nil, err
	}

	// get last entry for this reference ID
	res, err := s.db.Query(&dynamodb.QueryInput{
		TableName:                 aws.String(dataTable),
		IndexName:                 aws.String(dataRefIndex),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ProjectionExpression:      expr.Projection(),
		Limit:                     aws.Int64(1),
		ScanIndexForward:          aws.Bool(false),
	})

	if err != nil {
		return nil, err
	}

	for _, item := range res.Items {
		var dwd DynamoWalletData
		err = dynamodbattribute.UnmarshalMap(item, &dwd)
		if err != nil {
			return nil, err
		}

		return s.getObject(dwd.ObjectKey)
	}

	return nil, errors.New("cannot find " + refID)
}

func (s *AWSWalletStore) GetDataItem(ctx context.Context, tenantID, walletID, referenceId, hash string) (*WalletDataItem, error) {
	objectKey := fmt.Sprintf("%s/%s/%s/%s", tenantID, walletID, referenceId, hash)
	return s.getObject(objectKey)
}

func (s *AWSWalletStore) GetDataItemHistory(ctx context.Context, tenantID, walletID, referenceId string) (*WalletDataItemList, error) {
	refID := fmt.Sprintf("%s/%s/%s", tenantID, walletID, referenceId)
	key := expression.Key("referenceId").Equal(expression.Value(refID))
	proj := expression.NamesList(expression.Name("objectKey"))

	expr, err := expression.NewBuilder().WithKeyCondition(key).WithProjection(proj).Build()
	if err != nil {
		return nil, err
	}

	var objectKeys []string
	err = s.db.QueryPagesWithContext(ctx, &dynamodb.QueryInput{
		IndexName:                 aws.String(dataRefIndex),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ProjectionExpression:      expr.Projection(),
	},
		func(page *dynamodb.QueryOutput, lastPage bool) bool {
			for _, item := range page.Items {
				var dwd DynamoWalletData
				err = dynamodbattribute.UnmarshalMap(item, &dwd)
				if err != nil {
					//TODO: handle this?
					continue
				}

				objectKeys = append(objectKeys, dwd.ObjectKey)
			}
			return lastPage
		})
	if err != nil {
		return nil, err
	}

	objects, err := s.getObjects(objectKeys)
	if err != nil {
		return nil, err
	}

	return &WalletDataItemList{
		Items: objects,
	}, nil
}

func (s *AWSWalletStore) ListData(ctx context.Context, tenantID, walletID string) (*WalletList, error) {
	itemMap := make(map[string][]*WalletDataItemSummary)

	key := expression.Key("walletId").Equal(expression.Value(calcWalletID(tenantID, walletID)))

	expr, err := expression.NewBuilder().WithKeyCondition(key).Build()
	if err != nil {
		return nil, err
	}

	err = s.db.QueryPagesWithContext(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(dataTable),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	},
		func(page *dynamodb.QueryOutput, lastPage bool) bool {
			for _, item := range page.Items {
				var dwd DynamoWalletData
				err = dynamodbattribute.UnmarshalMap(item, &dwd)
				if err != nil {
					//TODO: handle this?
					continue
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
