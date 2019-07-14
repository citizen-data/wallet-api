package tests

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/citizendata/datawallet/wallet-api/store/wallets"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)


func TestCreateWallet(t *testing.T) {

	url := fmt.Sprintf("%s/wallet", testUrl)

	pubKeyBase64 := base64.StdEncoding.EncodeToString([]byte(publicKey))
	encryptedPrivateKey := encrypt(privateKey)

	wallet := &wallets.Wallet{
		PublicKeyBase64:     pubKeyBase64,
		PrivateKeyEncrypted: encryptedPrivateKey,
	}

	var jsonStr = []byte(wallet.Json())
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	assert.NoError(t, err)

	signRequest(req, string(jsonStr))

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)

	defer resp.Body.Close()

	if b, err := ioutil.ReadAll(resp.Body); err == nil {
		t.Log(string(b))
	}

	assert.Equal(t, 200, resp.StatusCode)
}

func TestAddData(t *testing.T) {
	url := fmt.Sprintf("%s/wallet/%s/data", testUrl, urlEncode(walletID))


	dataItem := &wallets.WalletDataItem{
		ReferenceID: "test123",
		DataSignature: "signature",
		EncryptedChunks: []string{
			uuid.New().String(),
			uuid.New().String(),
			uuid.New().String(),
		},
	}

	var jsonStr = []byte(dataItem.Json())
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	assert.NoError(t, err)

	signRequest(req, string(jsonStr))

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)

	defer resp.Body.Close()

	if b, err := ioutil.ReadAll(resp.Body); err == nil {
		t.Log(string(b))
	}

	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetList(t *testing.T) {
	url := fmt.Sprintf("%s/wallet/%s", testUrl, walletID)
	get(t, url)
}

func get(t *testing.T, url string) []byte {
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)

	signRequest(req, "")
	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	t.Log(string(b))
	assert.Equal(t, 200, resp.StatusCode)
	return b
}

func TestGetItems(t *testing.T) {
	// get wallet list
	get(t, fmt.Sprintf("%s/wallet/%s", testUrl, urlEncode(walletID)))

	// get item history
	get(t, fmt.Sprintf("%s/wallet/%s/data/%s", testUrl, urlEncode(walletID), "test123"))

	// get latest
	b := get(t, fmt.Sprintf("%s/wallet/%s/data/%s/latest", testUrl, urlEncode(walletID), "test123"))

	var data wallets.WalletDataItem
	err := json.Unmarshal(b, &data)
	assert.NoError(t, err)

	// get specific hash
	get(t, fmt.Sprintf("%s/wallet/%s/data/%s/%s", testUrl, urlEncode(walletID), "test123", urlEncode(data.VersionHash)))

}

func TestShareData(t *testing.T) {
	url := fmt.Sprintf("%s/wallet/%s/share/%s/data", testUrl, urlEncode(walletID), urlEncode(walletID))


	dataItem := &wallets.WalletDataItem{
		ReferenceID: "test123",
		DataSignature: "signature",
		EncryptedChunks: []string{
			uuid.New().String(),
			uuid.New().String(),
			uuid.New().String(),
		},
	}

	var jsonStr = []byte(dataItem.Json())
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	assert.NoError(t, err)

	signRequest(req, string(jsonStr))

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)

	defer resp.Body.Close()

	if b, err := ioutil.ReadAll(resp.Body); err == nil {
		t.Log(string(b))
	}

	assert.Equal(t, 200, resp.StatusCode)
}

func TestGetISharetems(t *testing.T) {
	// get wallet list
	get(t, fmt.Sprintf("%s/wallet/%s/shares", testUrl, urlEncode(walletID)))
}
