package auth

import (
	"golang.org/x/crypto/bcrypt"
)

func createPasswordHash(password string) (string, error) {
	hashpass, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(hashpass), nil
}

func verifyPasswordHash(hashpass string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashpass), []byte(password))
}
