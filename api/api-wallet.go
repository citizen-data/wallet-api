package api

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"github.com/citizendata/datawallet/wallet-api/store/wallets"
	"log"
	"strings"
)

type WalletAPI struct {
	walletStore wallets.WalletStore
}

func NewWalletAPI(store wallets.WalletStore) *WalletAPI {
	return &WalletAPI{
		walletStore: store,
	}
}

func (c *WalletAPI) authorizeWallet(ctx context.Context, request *ApiRequest) (string, *ApiResponse) {
	walletID, ok := request.PathParams["wallet"]
	if !ok {
		return "", NewApiError("invalid wallet ID in path", ErrorValidation)
	}

	wallet, err := c.walletStore.GetWallet(ctx, request.TenantID, walletID)
	if err != nil {
		return "", NewApiError("error getting wallet "+walletID+": "+err.Error(), ErrorValidation)
	}

	err = request.ValidateSignature(wallet.PublicKeyBase64)
	if err != nil {
		return "", NewApiError("invalid signature: "+err.Error()+", pub="+wallet.PublicKeyBase64, ErrorUnauthorized)
	}
	return walletID, nil
}

func (c *WalletAPI) AddData(ctx context.Context, request *ApiRequest) *ApiResponse {
	walletID, authErr := c.authorizeWallet(ctx, request)
	if authErr != nil {
		return authErr
	}

	var dataItem wallets.WalletDataItem
	err := json.Unmarshal([]byte(request.Body), &dataItem)
	if err != nil {
		return NewApiError("could not unmarshal payload", ErrorValidation)
	}

	if dataItem.ReferenceID == "" {
		return NewApiError("ReferenceID is required", ErrorValidation)
	}

	dataItem.CreatedAt = request.RequestTimeUTC

	encrypted := strings.Join(dataItem.EncryptedChunks,"")
	hash := sha256.Sum256([]byte(encrypted))
	dataItem.VersionHash = base64.URLEncoding.EncodeToString(hash[:])

	err = c.walletStore.AddDataItem(ctx, request.TenantID, walletID, &dataItem)

	if err != nil {
		return NewApiError("error saving data: "+err.Error(), ErrorValidation)
	}

	return ApiSuccessMessage("data saved successfully")
}

func (c *WalletAPI) GetData(ctx context.Context, request *ApiRequest) *ApiResponse {
	walletID, authErr := c.authorizeWallet(ctx, request)
	if authErr != nil {
		return authErr
	}

	refID, ok := request.PathParams["referenceId"]
	if !ok {
		return NewApiError("invalid reference ID in path", ErrorValidation)
	}

	version, ok := request.PathParams["version"]
	if !ok {
		return NewApiError("invalid version in path", ErrorValidation)
	}


	if version == "latest" {
		res, err := c.walletStore.GetLatestDataItem(ctx, request.TenantID, walletID, refID)
		if err != nil {
			return NewApiError("error getting data: "+err.Error(), ErrorInternalError)
		}
		return ApiResponseObject(res)
	} else {
		res, err := c.walletStore.GetDataItem(ctx, request.TenantID, walletID, refID, version)
		if err != nil {
			return NewApiError("error getting data: "+err.Error(), ErrorInternalError)
		}
		return ApiResponseObject(res)
	}
}

func (c *WalletAPI) GetDataHistory(ctx context.Context, request *ApiRequest) *ApiResponse {
	walletID, authErr := c.authorizeWallet(ctx, request)
	if authErr != nil {
		return authErr
	}

	refID, ok := request.PathParams["referenceId"]
	if !ok {
		return NewApiError("invalid reference ID in path", ErrorValidation)
	}

	res, err := c.walletStore.GetDataItemHistory(ctx, request.TenantID, walletID, refID)
	if err != nil {
		return NewApiError("error getting data: "+err.Error(), ErrorInternalError)
	}
	return ApiResponseObject(res)
}

func (c *WalletAPI) ListData(ctx context.Context, request *ApiRequest) *ApiResponse {
	walletID, authErr := c.authorizeWallet(ctx, request)
	if authErr != nil {
		return authErr
	}

	walletList, err := c.walletStore.ListData(ctx, request.TenantID, walletID)
	if err != nil {
		return NewApiError("error getting list: "+err.Error(), ErrorValidation)
	}

	return ApiResponseObject(walletList)
}

func (c *WalletAPI) CreateWallet(ctx context.Context, request *ApiRequest) *ApiResponse {
	var wallet wallets.Wallet
	err := json.Unmarshal([]byte(request.Body), &wallet)
	if err != nil {
		return NewApiError("could not unmarshal payload", ErrorValidation)
	}

	err = request.ValidateSignature(wallet.PublicKeyBase64)
	if err != nil {
		return NewApiError("invalid signature:  expected Base64(sha256(PKCS1v15('body|path|timestamp'))), "+request.Path, ErrorUnauthorized)
	}

	wallet.TenantID = request.TenantID

	walletID, err := wallet.CalculateWalletId()
	if err != nil {
		log.Print(err.Error())
		return NewApiError("could not calculate ID from public key (format)", ErrorValidation)
	}
	wallet.WalletID = walletID

	err = c.walletStore.CreateWallet(ctx, &wallet)
	if err != nil {
		log.Print(err.Error())
		return NewApiError("could not store wallet", ErrorInternalError)
	}

	return ApiResponseObject(wallet)
}

func (c *WalletAPI) ListMySharedItems(ctx context.Context, request *ApiRequest) *ApiResponse {
	walletID, authErr := c.authorizeWallet(ctx, request)
	if authErr != nil {
		return authErr
	}

	walletList, err := c.walletStore.ListSharedItems(ctx, request.TenantID, walletID)
	if err != nil {
		return NewApiError("error getting list: "+err.Error(), ErrorValidation)
	}

	return ApiResponseObject(walletList)
}

func (c *WalletAPI) GetSharedDataItem(ctx context.Context, request *ApiRequest) *ApiResponse {
	walletID, authErr := c.authorizeWallet(ctx, request)
	if authErr != nil {
		return authErr
	}

	fromWalletID, ok := request.PathParams["fromWallet"]
	if !ok {
		return NewApiError("invalid fromWalletID in path", ErrorValidation)
	}

	refID, ok := request.PathParams["referenceId"]
	if !ok {
		return NewApiError("invalid reference ID in path", ErrorValidation)
	}

	version, ok := request.PathParams["version"]
	if !ok {
		return NewApiError("invalid version in path", ErrorValidation)
	}

	res, err := c.walletStore.GetSharedDataItem(ctx, request.TenantID, fromWalletID, walletID, refID, version)
	if err != nil {
		return NewApiError("error getting data: "+err.Error(), ErrorInternalError)
	}
	return ApiResponseObject(res)
}

func (c *WalletAPI) ShareDataItem(ctx context.Context, request *ApiRequest) *ApiResponse {
	walletID, authErr := c.authorizeWallet(ctx, request)
	if authErr != nil {
		return authErr
	}

	toWalletID, ok := request.PathParams["toWallet"]
	if !ok {
		return NewApiError("invalid toWalletID in path", ErrorValidation)
	}

	var dataItem wallets.WalletDataItem
	err := json.Unmarshal([]byte(request.Body), &dataItem)
	if err != nil {
		return NewApiError("could not unmarshal payload", ErrorValidation)
	}

	if dataItem.ReferenceID == "" {
		return NewApiError("ReferenceID is required", ErrorValidation)
	}

	dataItem.CreatedAt = request.RequestTimeUTC

	encrypted := strings.Join(dataItem.EncryptedChunks,"")
	hash := sha256.Sum256([]byte(encrypted))
	dataItem.VersionHash = base64.URLEncoding.EncodeToString(hash[:])

	err = c.walletStore.ShareDataItem(ctx, request.TenantID, walletID, toWalletID, &dataItem)

	if err != nil {
		return NewApiError("error saving data: "+err.Error(), ErrorValidation)
	}

	return ApiSuccessMessage("data saved successfully")
}

func (c *WalletAPI) GetPublicKey(ctx context.Context, request *ApiRequest) *ApiResponse {
	walletID, ok := request.PathParams["wallet"]
	if !ok {
		return NewApiError("invalid wallet ID in path", ErrorValidation)
	}

	wallet, err := c.walletStore.GetWallet(ctx, request.TenantID, walletID)
	if err != nil {
		return NewApiError("error getting wallet "+walletID+": "+err.Error(), ErrorValidation)
	}

	return ApiResponseObject(wallet.PublicKeyBase64)
}