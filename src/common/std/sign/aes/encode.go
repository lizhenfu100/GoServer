package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

var (
	_key       = []byte("zf73.2/vb9w*Gh1e") //密钥长度为16的倍数
	_cipher, _ = aes.NewCipher(_key)
)

func Encode(orig string) []byte {
	blockSize := _cipher.BlockSize()
	blockMode := cipher.NewCBCEncrypter(_cipher, _key[:blockSize])
	buf := pkcs7Padding([]byte(orig), blockSize) //【补全，可能更改原buf】
	ret := make([]byte, len(buf))
	blockMode.CryptBlocks(ret, buf)
	return ret
}
func Decode(buf []byte) []byte {
	blockSize := _cipher.BlockSize()
	blockMode := cipher.NewCBCDecrypter(_cipher, _key[:blockSize])
	ret := make([]byte, len(buf))
	blockMode.CryptBlocks(ret, buf)
	ret = pkcs7UnPadding(ret)
	return ret
}
func pkcs7Padding(buf []byte, blockSize int) []byte {
	padding := blockSize - len(buf)%blockSize
	padbuf := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(buf, padbuf...)
}
func pkcs7UnPadding(buf []byte) []byte {
	kLen := len(buf)
	unpadding := int(buf[kLen-1])
	return buf[:kLen-unpadding]
}
