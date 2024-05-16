package main

import (
	"encoding/json"
	"net/http"
)

type UserView struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
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

	dbUser, getErr := cfg.database.CreateUser(params.Email, params.Password)
	if getErr != nil {
		responseWithError(w, http.StatusConflict, getErr.Error())
		return
	}

	responseWithJSON(w, 201, UserView{
		ID:    dbUser.Id,
		Email: dbUser.Email,
	})
}
