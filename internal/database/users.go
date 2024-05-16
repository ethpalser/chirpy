package database

import (
	"errors"
)

type User struct {
	Id       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
}

func (db *DB) CreateUser(email string, password string) (User, error) {
	data, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	if data.Users == nil {
		data.Users = make(map[int]User)
	}

	var existing *User
	for _, dbUser := range data.Users {
		if dbUser.Email == email {
			existing = &dbUser
			break
		}
	}

	if existing != nil {
		return User{}, errors.New("email already used by existing user")
	}

	id := len(data.Users) + 1
	user := User{
		Id:       id,
		Email:    email,
		Password: password,
	}
	data.Users[id] = user
	wErr := db.writeDB(data)
	if wErr != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) Login(email string, password string, expireSeconds int) (User, error) {
	data, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	var user *User
	for _, dbUser := range data.Users {
		if dbUser.Email == email {
			user = &dbUser
			break
		}
	}

	if user == nil {
		return User{}, errors.New("user does not exist")
	}

	return *user, nil
}
