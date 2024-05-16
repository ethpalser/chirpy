package database

import (
	"github.com/ethpalser/chirpy/internal/auth"
)

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (db *DB) CreateUser(email string, password string) (User, error) {
	data, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	existing := findUserByEmail(email, data.Users)
	if existing != nil {
		return User{}, ErrConflict
	}

	hashPassword, err := auth.CreatePasswordHash(password)
	if err != nil {
		return User{}, err
	}

	id := len(data.Users) + 1
	user := User{
		Id:       id,
		Email:    email,
		Password: hashPassword,
	}
	data.Users[id] = user
	wErr := db.writeDB(data)
	if wErr != nil {
		return User{}, err
	}

	return user, nil
}

func findUserByEmail(email string, users map[int]User) *User {
	var existing *User
	for _, dbUser := range users {
		if dbUser.Email == email {
			existing = &dbUser
			break
		}
	}
	return existing
}

func (db *DB) UpdateUser(id int, email string, password string) error {
	data, err := db.loadDB()
	if err != nil {
		return err
	}

	_, ok := data.Users[id]
	if !ok {
		return ErrNotExist
	}

	hashPassword, hashErr := auth.CreatePasswordHash(password)
	if hashErr != nil {
		return hashErr
	}

	data.Users[id] = User{
		Id:       id,
		Email:    email,
		Password: hashPassword,
	}

	return db.writeDB(data)
}

func (db *DB) Login(email string, password string, expireSeconds int) (User, error) {
	data, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	existing := findUserByEmail(email, data.Users)
	if existing == nil {
		return User{}, ErrConflict
	}

	verified := auth.VerifyPasswordHash(existing.Password, password)
	if verified != nil {
		return User{}, ErrUnauthorized
	}

	return *existing, nil
}
