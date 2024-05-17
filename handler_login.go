package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ethpalser/chirpy/internal/auth"
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

	dbUser, getErr := cfg.database.Login(params.Email, params.Password)
	if getErr != nil {
		responseWithError(w, http.StatusUnauthorized, getErr.Error())
		return
	}

	token, jwtErr := auth.IssueJWT(cfg.jwtSecret, fmt.Sprint(dbUser.Id), params.ExpireSeconds)
	if jwtErr != nil {
		responseWithError(w, http.StatusInternalServerError, jwtErr.Error())
		return
	}

	dbToken, refErr := cfg.database.CreateRefreshToken(dbUser.Id)
	if refErr != nil {
		responseWithError(w, http.StatusInternalServerError, refErr.Error())
		return
	}

	responseWithJSON(w, http.StatusOK, UserView{
		ID:           dbUser.Id,
		Email:        dbUser.Email,
		Token:        token,
		RefreshToken: dbToken.Val,
	})
}
