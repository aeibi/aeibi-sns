package util

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	jwt.RegisteredClaims
}

func GenerateJWT(subject, secret, issuer string, ttl time.Duration) (string, error) {
	if subject == "" {
		return "", errors.New("subject is required")
	}
	if secret == "" {
		return "", errors.New("secret is required")
	}
	if ttl <= 0 {
		return "", errors.New("ttl must be positive")
	}

	now := time.Now()
	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseJWT(tokenString, secret string) (*JWTClaims, error) {
	if tokenString == "" {
		return nil, errors.New("token is empty")
	}
	if secret == "" {
		return nil, errors.New("secret is required")
	}

	claims := &JWTClaims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

	if _, err := parser.ParseWithClaims(tokenString, claims, func(_ *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}); err != nil {
		return nil, err
	}

	if claims.Subject == "" {
		return nil, errors.New("token subject (uid) is empty")
	}

	return claims, nil
}
