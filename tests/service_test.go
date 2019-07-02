package tests

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/citizendata/datawallet/wallet-api/store/wallets"
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
	url := fmt.Sprintf("%s/wallet/%s", testUrl, walletID)

	dataItem := &wallets.WalletDataItem{
		ReferenceID: "test123",
		DataSignature: "signature",
		EncryptedChunks: []string{
			"chunk1",
			"chunk2",
			"chunk3",
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

func TestListData(t *testing.T) {
	url := fmt.Sprintf("%s/wallet/%s/list", testUrl, walletID)
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)

	signRequest(req, "")

	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)

	defer resp.Body.Close()

	if b, err := ioutil.ReadAll(resp.Body); err == nil {
		t.Log(string(b))
	}

	assert.Equal(t, 200, resp.StatusCode)
}