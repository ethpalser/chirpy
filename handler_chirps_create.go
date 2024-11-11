package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethpalser/chirpy/internal/auth"
	database2 "github.com/ethpalser/chirpy/internal/database/v2"
	"github.com/google/uuid"
)

type ChirpView struct {
	ID       int    `json:"id_old"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`

	UUID	uuid.UUID	`json:"id"`
	CreatedAt time.Time	`json:"created_at"`
	UpdatedAt time.Time	`json:"updated_at"`
	UserID	uuid.UUID	`json:"user_id"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type ChirpRequest struct {
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
	params := ChirpRequest{}
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
		ID:       dbChirp.ID,
		Body:     dbChirp.Message,
		AuthorID: userID,
	})
}

func (cfg *apiConfig) handlerChirpsCreateV2(w http.ResponseWriter, r *http.Request) {
	 type ChirpRequest struct {
               	Body string `json:"body"`
		UserID  string `json:"user_id"` // temp., as any user could author anyone's chirp
        }

	decoder := json.NewDecoder(r.Body)
	params := ChirpRequest{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		responseWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	cleaned, err := validateChirp(params.Body)
	if err != nil {
		responseWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := uuid.Parse(params.UserID)
	if err != nil {
		log.Printf("Failed to parse uuid %s string: %s\n", params.UserID, err.Error())
		responseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	args := database2.CreateChirpParams{
		Body: cleaned,
		UserID: userID,
	}
	dbChirp, err := cfg.dbQueries.CreateChirp(r.Context(), args)
	if err != nil {
		responseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responseWithJSON(w, http.StatusCreated, ChirpView{
		UUID: dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body: dbChirp.Body,
		UserID: dbChirp.UserID,
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
