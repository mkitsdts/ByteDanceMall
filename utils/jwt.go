package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 基础密钥（从环境变量或配置文件读取）
var baseSecret = []byte("secret_base")

type Claims struct {
	UserId uint64 `json:"user_id"`
	Salt   string `json:"salt"`
	jwt.RegisteredClaims
}

// 生成随机盐值
func generateSalt() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", bytes), nil
}

// 根据用户ID和盐值生成动态密钥
func generateDynamicKey(userId uint64, salt string) []byte {
	h := sha256.New()
	h.Write(baseSecret)
	fmt.Fprintf(h, "%d", userId)
	h.Write([]byte(salt))
	return h.Sum(nil)
}

func GenerateToken(userId uint64) (string, error) {
	// 生成随机盐值
	salt, err := generateSalt()
	if err != nil {
		return "", err
	}

	// 生成token
	claims := Claims{
		UserId: userId,
		Salt:   salt,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * 24 * time.Hour)),
		},
	}

	// 使用动态密钥
	dynamicKey := generateDynamicKey(userId, salt)

	// 使用HS256签名算法
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(dynamicKey)

	return s, err
}

func ParseToken(tokenstring string) (*Claims, error) {
	// 先解析获取盐值（不验证签名）
	token, _, err := new(jwt.Parser).ParseUnverified(tokenstring, &Claims{})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		slog.Error("invalid token")
		return nil, fmt.Errorf("无法解析claims")
	}

	// 使用解析出的用户ID和盐值重新生成密钥进行验证
	dynamicKey := generateDynamicKey(claims.UserId, claims.Salt)

	t, err := jwt.ParseWithClaims(tokenstring, &Claims{}, func(token *jwt.Token) (any, error) {
		return dynamicKey, nil
	})

	// 这里也要类型安全处理
	parsedClaims, ok := t.Claims.(*Claims)
	if ok && t.Valid {
		return parsedClaims, nil
	} else if !t.Valid {
		slog.Error("invalid token")
		return nil, fmt.Errorf("invalid token")
	} else {
		slog.Error("failed to parse token")
		return nil, err
	}
}
