package gozen

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

func UtilCryptoMD5Lower(str string) string {
	str = strings.ToLower(strings.TrimSpace(str))
	hash := md5.New()
	hash.Write([]byte(str))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func UtilCryptoMD5(str string) string {
	str = strings.TrimSpace(str)
	hash := md5.New()
	hash.Write([]byte(str))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func UtilCryptoSha1(s string) string {
	t := sha1.New()
	io.WriteString(t, s)
	return fmt.Sprintf("%x", t.Sum(nil))
}

func UtilCryptoMd5(s string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(s))
	cipherStr := md5Ctx.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func UtilCryptoMd5Lower(s string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(s))
	cipherStr := md5Ctx.Sum(nil)
	return strings.ToLower(hex.EncodeToString(cipherStr))
}

func UtilCryptoGenerateRandomToken16() (string, error) {
	return UtilCryptoGenerateRandomToken(8)
}

func UtilCryptoGenerateRandomToken32() (string, error) {
	return UtilCryptoGenerateRandomToken(16)
}

func UtilCryptoGenerateRandomToken(n int) (string, error) {
	token := make([]byte, n)
	_, err := rand.Read(token)
	return fmt.Sprintf("%x", token), err
}

//  ======= AES/CBC/PKCS5Padding ========

// 加密数据
func AesCbcEncrypt(key string, iv string, data []byte) ([]byte, error) {
	aesBlockEncrypter, err := aes.NewCipher([]byte(key))
	content := PKCS5Padding(data, aesBlockEncrypter.BlockSize())
	encrypted := make([]byte, len(content))
	if err != nil {
		return nil, err
	}
	aesEncrypter := cipher.NewCBCEncrypter(aesBlockEncrypter, []byte(iv))
	aesEncrypter.CryptBlocks(encrypted, content)
	return encrypted, nil
}

// 解密数据
func AesCbcDecrypt(key string, iv string, src []byte) (data []byte, err error) {
	decrypted := make([]byte, len(src))
	var aesBlockDecrypter cipher.Block
	aesBlockDecrypter, err = aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	aesDecrypter := cipher.NewCBCDecrypter(aesBlockDecrypter, []byte(iv))
	aesDecrypter.CryptBlocks(decrypted, src)
	return PKCS5Trimming(decrypted), nil
}

func AesCbcDecryptBase64(key string, iv string, str string) (data []byte, err error) {
	var src []byte
	src, err = base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	decrypted := make([]byte, len(src))
	var aesBlockDecrypter cipher.Block
	aesBlockDecrypter, err = aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	aesDecrypter := cipher.NewCBCDecrypter(aesBlockDecrypter, []byte(iv))
	aesDecrypter.CryptBlocks(decrypted, src)
	return PKCS5Trimming(decrypted), nil
}

/*
*
PKCS5包装
*/
func PKCS5Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, padText...)
}

/*
解包装
*/
func PKCS5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}
