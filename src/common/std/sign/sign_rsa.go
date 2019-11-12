package sign

import (
	"common"
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

// ------------------------------------------------------------
// 与客户端约定的签名规则
func CalcSign(s string) string { return GetSign(s, "yqqs(#(%$(%!$") }
func GetSign(s, key string) string {
	k := fmt.Sprintf("%x", md5.Sum([]byte(key)))
	sign := fmt.Sprintf("%s&%s", s, strings.ToLower(k))
	return fmt.Sprintf("%x", md5.Sum(common.S2B(sign)))
}

// ------------------------------------------------------------
var ( //我们生成的私钥公钥
	KPrivateKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCm5KmHwtDFdLTva5ntFqB1OrW7HnO8WR7m+ufFPrF/K0O2Hss
Y9DnL/IR4/zcNutvVDm6EtqfmSi8N9WjPZcQxpgceP2vD/DVftcjqagxk6n9XAX
BXIa/VkoDQn/+VtAqWn+n8SMfiypbgKPhnuEEjhac8zmiyUeK/RpVuCLG/NwIDA
QABAoGAOQHObtM64Ne2njmRAI1EDgcZ4GrMeb+vcJKv7I43rwqmPGVUVpfFzknZ
LkNvGL+FK8ecNLT3sRAR9qCnRoCqfQWgZbtjAMu8A7iI+Cv/2BOq4llTRg0hxC1
dFtNmrUbEnkfgYjKp0zhmjddTGfe8SIsCJAq4HJgtU9JCjc5jkYECQQDaNvywcv
ZDQSAXpQjogA49ZkuNkxuKS3nimshqzv7/UrD31tx8z8pSk7PNnM4BopyPvG9qT
JMLRHSJcDzVzppXAkEAw8qx7OtaZXDzJvjaBBkiX/EFAB28w73oYReL6zTVxNIv
7/O4iwlu90tmBGPL6BZY84259p9dx7p+ESuv2Zq2IQJBAMwnWa2zQJaXXXEBpB3
xgGENTW48zS1Lg9LvwMW8t3EkahDVYh8bQEyVh0i8hTeebR9EynAHCCMofmb/LM
tTqa0CQEg/7Bh5YQo9+/xNqGYKwFyXHDlGv/mbgr0Ra1iITroqtfXeAiOMf55R/
HtyODSUyo5VpmITvQ+PCiZb8LBkHwECQBWqACittloSz2U67ntJIl+1Cx2rjX5c
WHSxN2wrm6eniwVY4jGRQPIxAmX7VN2+xkcoKsPnWxhjsAtkxalus+s=
-----END RSA PRIVATE KEY-----`)
	KPublicKey = []byte(`-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCm5KmHwtDFdLTva5ntFqB1OrW7
HnO8WR7m+ufFPrF/K0O2HssY9DnL/IR4/zcNutvVDm6EtqfmSi8N9WjPZcQxpgce
P2vD/DVftcjqagxk6n9XAXBXIa/VkoDQn/+VtAqWn+n8SMfiypbgKPhnuEEjhac8
zmiyUeK/RpVuCLG/NwIDAQAB
-----END PUBLIC KEY-----`)

	g_rsa RSA
)

func init() {
	if e := g_rsa.Init(KPrivateKey, KPublicKey); e != nil {
		panic(e)
	}
}
func Decode(s ...*string) {
	for _, v := range s {
		if b, e := base64.StdEncoding.DecodeString(*v); e == nil {
			if b, e = g_rsa.Decrypt(b); e == nil {
				*v = common.B2S(b)
			}
		}
	}
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
func (this *RSA) Init(privateKey, publicKey []byte) (err error) {
	if err == nil && privateKey != nil { //生成私钥
		if block, _ := pem.Decode(privateKey); block != nil {
			this.PrivateKey, err = genPrivate(block.Bytes, block.Type)
		} else {
			err = errors.New("private key error")
		}
	}
	if err == nil && publicKey != nil { //生成公钥
		if block, _ := pem.Decode(publicKey); block != nil {
			this.PublicKey, err = genPublic(block.Bytes)
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
