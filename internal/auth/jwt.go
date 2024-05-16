package auth

import (
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func IssueJWT(subject string, secret string, expiresInSeconds int) (string, error) {
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
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireTime))),
		Subject:   subject,
	}

	token := jwt.NewWithClaims(signingMethod, registeredClaims)
	return token.SignedString([]byte(secret))
}

func ParseJWT(token string) (*jwt.Token, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return []byte("test"), nil
	}
	jwt, err := jwt.ParseWithClaims(token, jwt.MapClaims{}, keyFunc)
	if err != nil {
		log.Fatal("Issue parsing jwt token")
		return nil, err
	}
	return jwt, nil
}
