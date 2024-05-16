package database

import "errors"

type Chirp struct {
	Id      int    `json:"id"`
	Message string `json:"body"`
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
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
