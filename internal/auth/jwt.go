package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func IssueJWT(secret string, subject string, expiresInSeconds int) (string, error) {
	var expireTime int
	if 0 < expiresInSeconds && expiresInSeconds < 86400 {
		expireTime = expiresInSeconds
	} else {
		expireTime = 86400
	}
	signingMethod := jwt.SigningMethodHS256
	registeredClaims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(expireTime))),
		Subject:   subject,
	}

	token := jwt.NewWithClaims(signingMethod, registeredClaims)
	return token.SignedString([]byte(secret))
}

func ParseJWT(secret string, token string) (*jwt.Token, error) {
	keyFunc := func(jwtToken *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}

	jwt, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, keyFunc)
	if err != nil {
		return nil, err
	}
	return jwt, nil
}
