package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ethpalser/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
	type UserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := UserRequest{}
	err := decoder.Decode(&params)
	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	tokenHeader := r.Header.Get("Authorization")
	// Removing 'Bearer ' from 'Autorization: Bearer <token>'
	token := strings.TrimPrefix(tokenHeader, "Bearer ")
	jwt, parseErr := auth.ParseJWT(cfg.jwtSecret, token)
	if parseErr != nil {
		responseWithError(w, http.StatusUnauthorized, parseErr.Error())
		return
	}

	userIdStr, claimErr := jwt.Claims.GetSubject()
	if claimErr != nil {
		responseWithError(w, http.StatusUnauthorized, claimErr.Error())
		return
	}

	userId, convErr := strconv.Atoi(userIdStr)
	if convErr != nil {
		responseWithError(w, http.StatusUnauthorized, convErr.Error())
		return
	}

	upErr := cfg.database.UpdateUser(userId, params.Email, params.Password)
	if upErr != nil {
		responseWithError(w, http.StatusInternalServerError, upErr.Error())
		return
	}

	responseWithJSON(w, http.StatusOK, UserView{
		ID:    userId,
		Email: params.Email,
	})
}
