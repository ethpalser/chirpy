package auth

import (
	"golang.org/x/crypto/bcrypt"
)

func CreatePasswordHash(password string) (string, error) {
	hashpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashpass), nil
}

func VerifyPasswordHash(hashpass string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
}
