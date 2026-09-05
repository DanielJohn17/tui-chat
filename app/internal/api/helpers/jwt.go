package helpers

import (
	"fmt"
	"time"

	"github.com/DanielJohn17/tui-chat/app/internal/api/config"
	"github.com/golang-jwt/jwt/v5"
)

type UserToken struct {
	jwt.RegisteredClaims
	ID       int64
	Username string
}

var (
	secretKey []byte
	jwtExp    int32
)

func init() {
	secretKey = []byte(config.ENV.JWTSecretKey)
	jwtExp = config.ENV.JWTExpirationInSeconds
}

func CreateToken(user UserToken) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"email": user.Username,
		"exp":   time.Now().Add(time.Second * time.Duration(jwtExp)).Unix(),
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) (*UserToken, error) {
	userToken := new(UserToken)

	token, err := jwt.ParseWithClaims(
		tokenString,
		&UserToken{},
		func(token *jwt.Token) (any, error) {
			return secretKey, nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("error parsing token")
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if claims, ok := token.Claims.(*UserToken); ok {
		userToken.ID = claims.ID
		userToken.Username = claims.Username
	}

	return userToken, nil
}
