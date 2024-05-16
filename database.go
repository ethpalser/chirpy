package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
)

var connectionString string = "database.json"

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

func NewDB(path string) (*DB, error) {
	database := DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	err := database.ensureDB()
	if err != nil {
		return nil, err
	}

	return &database, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()
	data, err := db.loadDB()
	if err != nil {
		return Chirp{Message: ""}, err
	}

	if data.Chirps == nil {
		data.Chirps = make(map[int]Chirp)
	}

	id := len(data.Chirps) + 1
	chirp := Chirp{
		Id:      id,
		Message: body,
	}
	data.Chirps[id] = chirp
	wErr := db.writeDB(data)
	if wErr != nil {
		return Chirp{Message: ""}, err
	}

	return chirp, nil
}

func (db *DB) GetChirp(id int) (Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	data, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	chirp, exists := data.Chirps[id]
	if !exists {
		return Chirp{}, errors.New("does not exist")
	}
	return chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()
	data, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := []Chirp{}
	for _, chirp := range data.Chirps {
		chirps = append(chirps, chirp)
	}
	return chirps, err
}

func (db *DB) CreateUser(email string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	data, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	if data.Users == nil {
		data.Users = make(map[int]User)
	}

	id := len(data.Users) + 1
	user := User{
		Id:    id,
		Email: email,
	}
	data.Users[id] = user
	wErr := db.writeDB(data)
	if wErr != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if err != nil {
		os.WriteFile(db.path, []byte("{}"), 0644)
	}
	return nil
}

func (db *DB) loadDB() (DBStructure, error) {
	file, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}
	dbStruct := *new(DBStructure)
	mErr := json.Unmarshal(file, &dbStruct)
	if mErr != nil {
		log.Printf("Error decoding database JSON: %s", err)
		return DBStructure{}, err
	}
	return dbStruct, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	dbJSON, err := json.Marshal(dbStructure)
	if err != nil {
		log.Printf("Error marshaling JSON: %s", err)
		return err
	}
	os.WriteFile(db.path, []byte(dbJSON), 0644)
	return nil
}
