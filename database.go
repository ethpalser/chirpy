package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func NewDB(path string) (*DB, error) {
	database := DB{
		path: "database.json",
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
		return Chirp{message: ""}, err
	}

	chirp := Chirp{message: body}
	data.Chirps[len(data.Chirps)+1] = chirp
	wErr := db.writeDB(data)
	if wErr != nil {
		return Chirp{message: ""}, err
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

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if err != nil {
		os.WriteFile(db.path, nil, 0644)
		return nil
	}
	return os.ErrExist
}

func (db *DB) loadDB() (DBStructure, error) {
	file, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{nil}, err
	}
	dbStruct := *new(DBStructure)
	mErr := json.Unmarshal(file, &dbStruct)
	if mErr != nil {
		log.Printf("Error decoding database JSON: %s", err)
		return DBStructure{nil}, err
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
