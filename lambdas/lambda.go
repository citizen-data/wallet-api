package lambdas

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/citizendata/datawallet/wallet-api/api"
	"github.com/citizendata/datawallet/wallet-api/store/tenants"
	"github.com/citizendata/datawallet/wallet-api/store/wallets"
)

func InitWalletAPI(ctx context.Context, request events.APIGatewayProxyRequest) (*api.WalletAPI, *api.ApiRequest, error) {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)
	s3Svc := s3.New(sess)

	walletStore := wallets.NewAWSWalletStore(svc, s3Svc)
	tenantStore := tenants.NewDynamoTenantStore(svc)

	tenantID, err := tenantStore.GetTenantId(ctx, request.RequestContext.Identity.APIKey)
	req := api.ApiRequestFromLambda(&request, tenantID)

	if err != nil {
		return nil, nil, err
	}

	return api.NewWalletAPI(walletStore), req, err
}