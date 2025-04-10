package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/xelarion/go-layout/pkg/errs"
)

// GenRSAKey generate rsa public and private key use rsa
func GenRSAKey(bitSize int) (privateKey, publicKey []byte, err error) {
	// Generate RSA key.
	key, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, nil, errs.WrapInternal(err, "failed to generate private key")
	}

	// Extract public component.
	pub := key.Public()
	// Encode private key to PKCS#1 ASN.1 PEM.
	privateKey = pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	// Encode public key to PKCS#1 ASN.1 PEM.
	publicKey = pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: x509.MarshalPKCS1PublicKey(pub.(*rsa.PublicKey)),
		},
	)

	return privateKey, publicKey, nil
}

// RSAEncrypt encrypt data use rsa public key
func RSAEncrypt(data, publicKey []byte) ([]byte, error) {
	// Decode public key.
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errs.NewInternal("failed to parse PEM block containing the public key")
	}

	// Parse public key.
	pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, errs.WrapInternal(err, "failed to parse public key")
	}

	// Encrypt data.
	return rsa.EncryptPKCS1v15(rand.Reader, pub, data)
}

// RSADecrypt decrypt data use rsa private key
func RSADecrypt(data, privateKey []byte) ([]byte, error) {
	// Decode private key.
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errs.NewInternal("failed to parse PEM block containing the private key")
	}

	// Parse private key.
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errs.WrapInternal(err, "failed to parse private key")
	}

	// Decrypt data.
	return rsa.DecryptPKCS1v15(rand.Reader, key, data)
}
