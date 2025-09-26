package utils

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateRefreshToken 生成 32 字节(256 bit) 高熵刷新令牌，使用 URL 安全 Base64 无填充编码
// 返回长度通常为 43 字符（因为 32 raw bytes -> base64 = 44 去掉=变 43），可直接用于存储与传输
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32) // 256 bit
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
