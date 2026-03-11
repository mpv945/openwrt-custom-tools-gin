package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT 密钥（生产建议使用 env 配置或 Vault）
//var secret = []byte("my-secret-key")

var secret []byte

func InitJWT(s string) {
	//s := viper.GetString("jwt.secret")
	if s == "" {
		panic("jwt.secret not configured")
	}
	secret = []byte(s)
}

// Claims 自定义 JWT Claim
type Claims struct {
	UserID string   `json:"uid"`
	Roles  []string `json:"roles"` // 用户角色
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT
func GenerateToken(userID string, roles []string, expireHours int64) (string, error) {

	claims := Claims{
		UserID: userID,
		Roles:  roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "gin-jwt",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secret)
}

// ParseToken 解析并验证 JWT
func ParseToken(tokenStr string) (*Claims, error) {

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
