package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	// Parse
	type parameters struct {
		Email         string `json:"email"`
		Password      string `json:"password"`
		ExpireSeconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		responseWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	dbUser, getErr := cfg.database.Login(params.Email, params.Password, params.ExpireSeconds)
	if getErr != nil {
		responseWithError(w, http.StatusUnauthorized, getErr.Error())
		return
	}

	responseWithJSON(w, http.StatusOK, UserView{
		ID:    dbUser.Id,
		Email: dbUser.Email,
	})
}
