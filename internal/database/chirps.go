package database

import "errors"

type Chirp struct {
	Id       int    `json:"id"`
	Message  string `json:"body"`
	AuthorID int    `json:"user_id"`
}

func (db *DB) CreateChirp(body string, authorID int) (Chirp, error) {
	data, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(data.Chirps) + 1
	chirp := Chirp{
		Id:       id,
		Message:  body,
		AuthorID: authorID,
	}
	data.Chirps[id] = chirp
	wErr := db.writeDB(data)
	if wErr != nil {
		return Chirp{}, wErr
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

func (db *DB) DeleteChirp(id int) error {
	data, err := db.loadDB()
	if err != nil {
		return err
	}

	_, exists := data.Chirps[id]
	if !exists {
		return ErrNotExist
	}

	// hard delete
	delete(data.Chirps, id)
	wErr := db.writeDB(data)
	if wErr != nil {
		return wErr
	}
	return nil
}
