package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/citizendata/datawallet/wallet-api/api"
	"github.com/citizendata/datawallet/wallet-api/lambdas"
)

type Response events.APIGatewayProxyResponse


// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error){
	ctx := context.Background()
	walletAPI, req, err := lambdas.InitWalletAPI(ctx, request)
	if err != nil {
		resp := api.NewApiError(err.Error(), api.ErrorValidation)
		return *api.LambdaResponseFromApiResponse(resp), nil
	}

	apiResp := walletAPI.AddData(ctx, req)
	return *api.LambdaResponseFromApiResponse(apiResp), nil
}

func main() {
	lambda.Start(Handler)
}
