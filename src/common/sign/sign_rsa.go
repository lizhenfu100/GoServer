package sign

import (
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

// ------------------------------------------------------------
// 与客户端约定的签名规则
func CalcSign(s string) string {
	const k_pt_key = "yqqs(#(%$(%!$"
	key := fmt.Sprintf("%x", md5.Sum([]byte(k_pt_key)))
	sign := fmt.Sprintf("%s&%s", s, strings.ToLower(key))
	//fmt.Println("-----------CalcSign: ", key)
	return fmt.Sprintf("%x", md5.Sum([]byte(sign)))
}

// ------------------------------------------------------------
type RSA struct {
	PrivateKey *rsa.PrivateKey //私钥，自己保留：给数据签名，产生sign发给用户
	PublicKey  *rsa.PublicKey  //公钥，发给用户：验证数据是否来自公钥发放者
}

func (this *RSA) Encrypt(plaintext []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, this.PublicKey, plaintext)
}
func (this *RSA) Decrypt(ciphertext []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, this.PrivateKey, ciphertext)
}
func (this *RSA) Sign(src []byte, hash crypto.Hash) ([]byte, error) {
	h := hash.New()
	h.Write(src)
	hashed := h.Sum(nil)
	return rsa.SignPKCS1v15(rand.Reader, this.PrivateKey, hash, hashed)
}
func (this *RSA) Verify(src []byte, sign []byte, hash crypto.Hash) error {
	h := hash.New()
	h.Write(src)
	hashed := h.Sum(nil)
	return rsa.VerifyPKCS1v15(this.PublicKey, hash, hashed, sign)
}

// ------------------------------------------------------------
func NewRSA(privateKey, publicKey []byte) (ret *RSA, err error) {
	ret = &RSA{}
	if err == nil && privateKey != nil { //生成私钥
		if block, _ := pem.Decode(privateKey); block != nil {
			ret.PrivateKey, err = genPrivate(block.Bytes, block.Type)
		} else {
			err = errors.New("private key error")
		}
	}
	if err == nil && publicKey != nil { //生成公钥
		if block, _ := pem.Decode(publicKey); block != nil {
			ret.PublicKey, err = genPublic(block.Bytes)
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
