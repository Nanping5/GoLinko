package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte("GoLinkoSecretKey")

// Claims 定义 JWT 载荷结构
type Claims struct {
	UserId string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateJwtToken 生成JWT令牌，有效期为24小时
func GenerateJwtToken(userId string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "GoLinko",
			Subject:   "user token",
			ID:        GenerateUUID(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseJwtToken 解析JWT令牌并返回Claims
func ParseJwtToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if token != nil {
		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			return claims, nil
		}
	}
	return nil, err
}
