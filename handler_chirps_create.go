package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

type ChirpView struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type ChripRequest struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := ChripRequest{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		responseWithError(w, 500, "Something went wrong")
		return
	}

	cleaned, err := validateChirp(params.Body)
	if err != nil {
		responseWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	dbChirp, getErr := cfg.database.CreateChirp(cleaned)
	if getErr != nil {
		responseWithError(w, http.StatusBadRequest, getErr.Error())
		return
	}
	responseWithJSON(w, http.StatusCreated, ChirpView{
		ID:   dbChirp.Id,
		Body: dbChirp.Message,
	})
}

func validateChirp(msg string) (string, error) {
	if len(msg) > 140 {
		return "", errors.New("chirp is too long")
	}

	profanity := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	cleaned := filterProfanity(msg, profanity)
	return cleaned, nil
}

func filterProfanity(msg string, badWords map[string]struct{}) string {
	msg_split := strings.Split(msg, " ")
	lower := strings.ToLower(msg)
	lower_split := strings.Split(lower, " ")
	for index, word := range lower_split {
		_, ok := badWords[word]
		if ok {
			msg_split[index] = "****"
		}
	}
	return strings.Join(msg_split, " ")
}
