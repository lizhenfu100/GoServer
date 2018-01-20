package api

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// ------------------------------------------------------------
//
type RSA struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func (this *RSA) Encrypt(plaintext []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, this.publicKey, plaintext)
}
func (this *RSA) Decrypt(ciphertext []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, this.privateKey, ciphertext)
}
func (this *RSA) Sign(src []byte, hash crypto.Hash) ([]byte, error) {
	h := hash.New()
	h.Write(src)
	hashed := h.Sum(nil)
	return rsa.SignPKCS1v15(rand.Reader, this.privateKey, hash, hashed)
}
func (this *RSA) Verify(src []byte, sign []byte, hash crypto.Hash) error {
	h := hash.New()
	h.Write(src)
	hashed := h.Sum(nil)
	return rsa.VerifyPKCS1v15(this.publicKey, hash, hashed, sign)
}

// ------------------------------------------------------------
//
func NewRSA(privateKey, publicKey string) (ret *RSA, err error) {
	ret = &RSA{}
	if err == nil { //生成私钥
		if block, _ := pem.Decode([]byte(privateKey)); block != nil {
			ret.privateKey, err = genPrivate(block.Bytes, block.Type)
		} else {
			err = errors.New("private key error")
		}
	}
	if err == nil { //生成公钥
		if block, _ := pem.Decode([]byte(publicKey)); block != nil {
			ret.publicKey, err = genPublic(block.Bytes)
		} else {
			err = errors.New("public key error")
		}
	}
	return
}
func genPrivate(privateKey []byte, typ string) (ret *rsa.PrivateKey, err error) {
	switch typ {
	case "RSA PRIVATE KEY":
		ret, err = x509.ParsePKCS1PrivateKey(privateKey)
	case "PRIVATE KEY":
		var key interface{}
		if key, err = x509.ParsePKCS8PrivateKey(privateKey); err == nil {
			ret = key.(*rsa.PrivateKey)
		}
	default:
		err = errors.New("unsupport private key type")
	}
	return
}
func genPublic(publicKey []byte) (*rsa.PublicKey, error) {
	if pub, err := x509.ParsePKIXPublicKey(publicKey); err == nil {
		return pub.(*rsa.PublicKey), nil
	} else {
		return nil, err
	}
}
