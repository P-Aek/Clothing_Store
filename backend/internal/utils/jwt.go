package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const TokenTTL = 24 * time.Hour

type JWTClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID, role, secret string, now time.Time) (string, error) {
	claims := JWTClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now.UTC()),
			ExpiresAt: jwt.NewNumericDate(now.UTC().Add(TokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseJWT(tokenString, secret string) (JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return JWTClaims{}, errors.New("invalid token")
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || claims.Subject == "" {
		return JWTClaims{}, errors.New("invalid token claims")
	}
	return *claims, nil
}
