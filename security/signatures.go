package security

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
)

func VerifySignature(paylod []byte, signatureBase64 string, pubKey *rsa.PublicKey) error {
	hashed := sha256.Sum256(paylod)
	sigBytes, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return errors.New("signature not base64: "+err.Error())
	}
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], sigBytes)
}

func PemBase64ToPublicKey(publicKeyBase64 string) (*rsa.PublicKey, error) {
	pemStr, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return nil, err
	}
	data, _ := pem.Decode([]byte(pemStr))

	if data == nil {
		return nil, errors.New("cannot pem decode publicKey")
	}

	pubKey, err := x509.ParsePKCS1PublicKey(data.Bytes)
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}