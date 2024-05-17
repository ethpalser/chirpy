package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethpalser/chirpy/internal/auth"
)

type ChirpView struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type ChripRequest struct {
		Body string `json:"body"`
	}

	accessToken := r.Header.Get("Authorization")
	if accessToken == "" {
		responseWithError(w, http.StatusUnauthorized, "invalid auth token")
		return
	}
	tokenVal := strings.TrimPrefix(accessToken, "Bearer ")
	jwtToken, jwtErr := auth.ParseJWT(cfg.jwtSecret, tokenVal)
	if jwtErr != nil {
		responseWithError(w, http.StatusInternalServerError, jwtErr.Error())
		return
	}
	user, claimErr := jwtToken.Claims.GetSubject()
	if claimErr != nil {
		responseWithError(w, http.StatusInternalServerError, claimErr.Error())
		return
	}
	userID, atoiErr := strconv.Atoi(user)
	if atoiErr != nil {
		responseWithError(w, http.StatusInternalServerError, atoiErr.Error())
		return
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

	dbChirp, getErr := cfg.database.CreateChirp(cleaned, userID)
	if getErr != nil {
		responseWithError(w, http.StatusBadRequest, getErr.Error())
		return
	}
	responseWithJSON(w, http.StatusCreated, ChirpView{
		ID:       dbChirp.Id,
		Body:     dbChirp.Message,
		AuthorID: userID,
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
