package database

import (
	"errors"
	"sort"

	"github.com/ethpalser/chirpy/internal/util"
)

type Chirp struct {
	ID       int    `json:"id"`
	Message  string `json:"body"`
	AuthorID int    `json:"user_id"`
}

type ChirpOptions struct {
	AuthorID int
	SortAsc  bool
}

func (db *DB) CreateChirp(body string, authorID int) (Chirp, error) {
	data, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(data.Chirps) + 1
	chirp := Chirp{
		ID:       id,
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

func (db *DB) GetChirps(opts ChirpOptions) ([]Chirp, error) {
	data, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := []Chirp{}
	for _, chirp := range data.Chirps {
		chirps = append(chirps, chirp)
	}

	// Filter by AuthorID, ignore Zero value
	if opts.AuthorID != 0 {
		chirps = util.Filter(chirps, func(chirp Chirp) bool {
			return opts.AuthorID == chirp.AuthorID
		})
	}

	sort.Slice(chirps, func(i, j int) bool {
		if opts.SortAsc {
			return chirps[i].ID < chirps[j].ID
		} else {
			return chirps[i].ID > chirps[j].ID
		}
	})

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
