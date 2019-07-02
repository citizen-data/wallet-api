package wallets

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
)

type Wallet struct {

	// TenantID is set by API via API-KEY
	TenantID string `json:"tenantId"`

	// WalletID is set by API (sha256 of public key)
	WalletID string `json:"walletId"`

	//PublicKeyBase64 Base64 of public RSA pem
	PublicKeyBase64 string `json:"publicKeyBase64"`

	//PrivateKeyEncrypted is opaque/encrypted private key for recovery
	PrivateKeyEncrypted string `json:"privateKeyEncrypted"`
}

func (w *Wallet) Json() string {
	res, err := json.Marshal(w)
	if err != nil {
		panic(err)
	}
	return string(res)
}

func (w *Wallet) CalculateWalletId() (string, error) {
	if w.PublicKeyBase64 == "" {
		return "", errors.New("cannot calculate ID from blank public key")
	}
	publicKey, err := base64.StdEncoding.DecodeString(w.PublicKeyBase64)
	if err != nil {
		return "", err
	}
	b := sha256.Sum256([]byte(publicKey))
	return base64.URLEncoding.EncodeToString(b[:]), nil
}

type WalletList struct {
	//map of reference IDs to versions, ordered from oldest to newest
	Items map[string][]*WalletDataItemSummary `json:"items"`
}

type WalletDataItem struct {
	ReferenceID     string   `json:"referenceId"`
	EncryptedChunks []string `json:"encryptedChunks"`
	VersionHash     string   `json:"versionHash"`
	DataSignature   string   `json:"dataSignature"`
	CreatedAt       string   `json:"createdAt"`
}

type WalletDataItemList struct {
	Items []*WalletDataItem
}

// WalletDataItem minus EncryptedChunks
type WalletDataItemSummary struct {
	ReferenceID   string `json:"referenceId"`
	DataSignature string `json:"dataSignature"`
	CreatedAt     string `json:"createdAt"`
	VersionHash   string `json:"versionHash"`
}

func (w *WalletDataItem) Json() string {
	res, err := json.Marshal(w)
	if err != nil {
		panic(err)
	}
	return string(res)
}

type WalletStore interface {
	CreateWallet(ctx context.Context, wallet *Wallet) error
	GetWallet(ctx context.Context, tenantID, walletId string) (*Wallet, error)
	ListData(ctx context.Context, tenantID, walletID string) (*WalletList, error)
	GetLatestDataItem(ctx context.Context, tenantID, walletID, referenceID string) (*WalletDataItem, error)
	GetDataItem(ctx context.Context, tenantID, walletID, referenceID, hash string) (*WalletDataItem, error)
	GetDataItemHistory(ctx context.Context, tenantID, walletID, referenceID string) (*WalletDataItemList, error)
	AddDataItem(ctx context.Context, tenantID, walletID string, data *WalletDataItem) error
}
