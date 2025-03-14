package utils

import (
	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("adgihioasxbfjkcbAEWIOFGHBIOHasegfWEAWEgWEARx")

type Claims struct {
	UserId int64
	jwt.StandardClaims
}

func GenerateToken(userId int64) (string, error) {
	claims := &Claims{
		UserId: userId,
		StandardClaims: jwt.StandardClaims{},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ParseToken(token string) (*Claims, error) {
	// 把字符串token解析成token对象
	tokenObj, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	// 从token对象中获取claims
	claims, ok := tokenObj.Claims.(*Claims)
	if !ok {
		return nil, err
	}
	return claims, nil
}