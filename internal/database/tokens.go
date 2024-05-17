package database

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Token struct {
	UserID int       `json:"user_id"`
	Val    string    `json:"val"`
	Iss    time.Time `json:"iss"`
	Exp    time.Time `json:"exp"`
}

func (db *DB) CreateRefreshToken(userID int) (Token, error) {
	c := 32
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		return Token{}, err
	}

	database, err := db.loadDB()
	if err != nil {
		return Token{}, err
	}

	key := hex.EncodeToString(b)

	token := Token{
		UserID: userID,
		Val:    key,
		Iss:    time.Now(),
		Exp:    time.Now().Add(time.Hour * 1440),
	}

	database.Tokens[key] = token
	wErr := db.writeDB(database)
	if wErr != nil {
		return Token{}, wErr
	}
	return token, nil
}

func (db *DB) FindRefreshToken(token string) (Token, error) {
	database, err := db.loadDB()
	if err != nil {
		return Token{}, err
	}

	existing, ok := database.Tokens[token]
	if !ok {
		return Token{}, ErrNotExist
	}
	return existing, nil
}

func (db *DB) RevokeRefreshToken(token string) error {
	c := 32
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		return err
	}

	database, err := db.loadDB()
	if err != nil {
		return err
	}

	existing, ok := database.Tokens[token]
	if !ok {
		return ErrNotExist
	}
	// Revoke by expiring token
	existing.Exp = time.Now()
	database.Tokens[token] = existing
	return db.writeDB(database)
}
