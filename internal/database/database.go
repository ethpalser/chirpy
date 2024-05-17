package database

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

var ErrConflict = errors.New("conflict with existing resource")
var ErrNotExist = errors.New("resource does not exist")
var ErrUnauthorized = errors.New("unauthorized access")

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp    `json:"chirps"`
	Users  map[int]User     `json:"users"`
	Tokens map[string]Token `json:"tokens"`
}

func NewDB(path string) (*DB, error) {
	database := &DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	err := database.ensureDB()
	return database, err
}

func (db *DB) createDB() error {
	dbStructure := DBStructure{
		Chirps: map[int]Chirp{},
		Users:  map[int]User{},
		Tokens: map[string]Token{},
	}
	return db.writeDB(dbStructure)
}

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return db.createDB()
	}
	return err
}

func (db *DB) ResetDB() error {
	err := os.Remove(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return db.ensureDB()
}

func (db *DB) loadDB() (DBStructure, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	dbStructure := DBStructure{}
	file, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return dbStructure, err
	}

	jsonErr := json.Unmarshal(file, &dbStructure)
	if jsonErr != nil {
		return dbStructure, jsonErr
	}

	return dbStructure, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbJSON, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	return os.WriteFile(db.path, []byte(dbJSON), 0644)
}
