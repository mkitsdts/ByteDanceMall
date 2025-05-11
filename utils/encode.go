package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

const (
	// 信用卡加密私钥
	creditCardEncryptKey = "u_gu13i*bfeqw5@4IFh^O4KBC5%Vo7i5Wge563!ra124zfdAate115"
)

// CreditCard加密
func EncryptCreditCard(creditCard string) (string, error) {
	// 使用AES加密算法加密信用卡信息
	encryptedCard, err := AesEncrypt([]byte(creditCard), []byte(creditCardEncryptKey))
	if err != nil {
		return "", err
	}
	// 使用Base64编码以便于存储和传输
	return base64.StdEncoding.EncodeToString(encryptedCard), nil
}

// CreditCard解密
func DecryptCreditCard(encryptedCard string) (string, error) {
	// Base64解码
	cipherText, err := base64.StdEncoding.DecodeString(encryptedCard)
	if err != nil {
		return "", err
	}

	// 使用AES解密算法解密信用卡信息
	decryptedCard, err := AesDecrypt(cipherText, []byte(creditCardEncryptKey))
	if err != nil {
		return "", err
	}
	return string(decryptedCard), nil
}

// AES加密算法
func AesEncrypt(plainText, key []byte) ([]byte, error) {
	// 创建cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建gcm
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 创建nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// 加密
	cipherText := gcm.Seal(nonce, nonce, plainText, nil)
	return cipherText, nil
}

// AES解密算法
func AesDecrypt(cipherText, key []byte) ([]byte, error) {
	// 创建cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 创建gcm
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 确保数据长度合法
	if len(cipherText) < gcm.NonceSize() {
		return nil, errors.New("密文长度太短")
	}

	// 提取nonce
	nonce, cipherTextWithoutNonce := cipherText[:gcm.NonceSize()], cipherText[gcm.NonceSize():]

	// 解密
	plainText, err := gcm.Open(nil, nonce, cipherTextWithoutNonce, nil)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}
