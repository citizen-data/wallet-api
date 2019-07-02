package tests

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	timestampLayout = "2006-01-02T15:04:05.000Z"
)

var (
	apiKey                 = os.Getenv("DATA_WALLET_API_KEY")
	testUrl                = os.Getenv("DATA_WALLET_TEST_URL")
	superTopSecretPassword = "password"

	walletID = "zIfL2CPMg7pZ3pxxQKrmjLGZgqN6t9k1pU7lAHaRPEE="

	// Only for Tests
	publicKey = `-----BEGIN PUBLIC KEY-----
MIIBCgKCAQEA7myZ+NkPUSoWT1Fgd3sVGaXJWm2WJOc2n/+j6M7Jf1mgjUcStPCB
ZgIA5081ZPfdAvfofCbnns+rd60Dd3SFAt5ru/xIYdRQ75hkd94ElVzmRUQMq/Im
Zg6tgxt7Z00m5kuX8AIlRGkZsmmFWxMtL3BwlFqEhBQWoBdSo7JqLrQ+zgO4WEPs
cG7Uk/ViZBgMLsnT4YwlAzIuQ/uyKe20yHfv0ikWbpCkWPFV75d+1hnSJR/V4FLv
tacd07vynEKpYndfjdEO4IwaJlAvUOufNuGKvhNShSEcrLN0qMHGik228BSCSlCm
exp5wUd7mVsOWkNbsthctGWL0qkQaSOeOQIDAQAB
-----END PUBLIC KEY-----`

	// Only for tests
	privateKey = `-----BEGIN PRIVATE KEY-----
MIIEowIBAAKCAQEA7myZ+NkPUSoWT1Fgd3sVGaXJWm2WJOc2n/+j6M7Jf1mgjUcS
tPCBZgIA5081ZPfdAvfofCbnns+rd60Dd3SFAt5ru/xIYdRQ75hkd94ElVzmRUQM
q/ImZg6tgxt7Z00m5kuX8AIlRGkZsmmFWxMtL3BwlFqEhBQWoBdSo7JqLrQ+zgO4
WEPscG7Uk/ViZBgMLsnT4YwlAzIuQ/uyKe20yHfv0ikWbpCkWPFV75d+1hnSJR/V
4FLvtacd07vynEKpYndfjdEO4IwaJlAvUOufNuGKvhNShSEcrLN0qMHGik228BSC
SlCmexp5wUd7mVsOWkNbsthctGWL0qkQaSOeOQIDAQABAoIBAH76ObpNJ5orVvxh
M4YOd/tTzvuo91iqBu6JQGshjjCTtCzpmC3jFJcWJBMMnTwrwXnuO9O7CIoMwZ4X
94ur85bGjAvu3UG0yHIB8CfihwBhHIXxKup8KTUbRg9YGI24iioGQmHhUqWvb68U
AaCygtMMB/kKiB6rcC1Mnodt4p0B/HTP2C6G449jDEuPo/HHy6jx+EKtOljHDoR2
bjs9MD8/LLWltWYk3XG5WNuccCJZZAeNR6LuF/1VeknRXaea8aHwUiR3OjeaB9uk
Icv97obk88oEP/Ds6x6vO1u3kDqEFqNMpDlaEwMYg8ZLUMZLoTw82yRIZ3jJnCOS
7milVX0CgYEA9czkHXoqIn6FcQK2LCN3l2YhGVNFbd2GdKzDbZGF/pK6eFFDvSL5
39tpBcVam+M48ZsoYzVaca43NxrzHdqpdB6/+YdHJ7EWReicLGQCQOYdo0nhUgF+
ccgKtpca3G78+ZEhhpjTpMzeKwDhfZFDFNkaNmE5kqq5A8thALhMwisCgYEA+FFa
xfhqtCMbmMDmfv56Uf9maFv+bsEln8MAH6zhWk3UyowbgdtUq8h246CwjiPibAOs
xdkhHos+d/8PWZQ2/vAaJrM2+vIzI69D2soeQSeDCgPKxmS4lyL8cdr5GWZS+6Vw
qG5LMc0/8lCYoya6NH2PCvMrH8kMTqm//zwpgysCgYEAp0/6jt4TRDufFZfk7RKP
Wy0Xpqd6ARjjZxQaSsDd1rWF3FRkqZ/fOrOdP2JhFO+MWVlmGnG8yNjvmMDtcArh
gbtUrcOZebkfEiMN+2Fv70E0N2wYxbtimIy0Til5DUc3R6G0kmwA1JLnP5pv4ws4
AD7visiPafhvy9dqhhTtmtUCgYAJrVH2SRoPbxbSOyJAbLZjn6pkAsHFmy1WLolA
ssINfN8ADbm8s8l28FcBw+9derSGNRZ0l2OdBxwmHQCCIy6JfN3oCC/qU6n+iAQC
8MGBFIMczs0GMkKnUSu5XCk8/inZuLbNOY8gn7kQPmfUY9v507LRYGybzn/2SNM8
pSGRBQKBgE+BDq+6WrwTFNK6eRvhVgbBwy7Cf4jgvaXZW1mr6uQwF8khLhaR7oVf
Ko0Hn77Z22xaV5YFfTFHokIDMPFGQBsYvQyq2T4WVe+CS8kCE9uarGWSWaCUFwvA
f5MhFBDcLMZBTNwoby0omasPtB0Ne9HNfto7juRlUFCICo1IIpAY
-----END PRIVATE KEY-----`
)

func getPrivateKey() *rsa.PrivateKey {
	data, _ := pem.Decode([]byte(privateKey))

	privKey, err := x509.ParsePKCS1PrivateKey(data.Bytes)
	if err != nil {
		panic(err)
	}

	return privKey
}

func getPublicKey() *rsa.PublicKey {
	data, _ := pem.Decode([]byte(publicKey))

	privKey, err := x509.ParsePKCS1PublicKey(data.Bytes)
	if err != nil {
		panic(err)
	}

	return privKey
}

func signRequest(req *http.Request, body string) {
	pk := getPrivateKey()

	timestamp := time.Now().UTC().Format(timestampLayout)
	payload := []byte(fmt.Sprintf("%s|%s|%s", strings.Replace(req.URL.Path, "/dev","", 1), body, timestamp))

	hashed := sha256.Sum256(payload)

	sig, err := rsa.SignPKCS1v15(rand.Reader, pk, crypto.SHA256, hashed[:])
	if err != nil {
		panic(err)
	}

	signature := base64.StdEncoding.EncodeToString(sig)

	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("x-api-timestamp", timestamp)
	req.Header.Set("x-api-signature", signature)
	req.Header.Set("Content-Type", "application/json")
}

func urlEncode(str string) string {
	return url.PathEscape(str)
}

func encrypt(plaintext string) string {
	// The key argument should be the AES key, either 16 or 32 bytes
	// to select AES-128 or AES-256.
	key := []byte("AES256Key-32Characters1234567890")
	plaintextBytes := []byte(plaintext)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintextBytes, nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}
