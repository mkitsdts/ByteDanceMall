package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func GenerateSalt() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func EncryptPassword(password, salt string) string {
	h := sha256.New()
	h.Write([]byte(salt))
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

func VerifyPassword(password, salt, encrypted string) bool {
	if salt == "" {
		return encryptLegacyPassword(password) == encrypted
	}
	return EncryptPassword(password, salt) == encrypted
}

func encryptLegacyPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}
