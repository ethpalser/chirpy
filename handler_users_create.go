package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type UserView struct {
	ID           int    `json:"id"`
	UUID		uuid.UUID	`json:"uuid"`
	Email        string `json:"email"`
	Token        string `json:"token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	PremiumRed   bool   `json:"is_chirpy_red"`
	CreatedAt	time.Time	`json:"created_at"`
	UpdatedAt	time.Time	`json:"updated_at"`
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
		responseWithError(w, http.StatusInternalServerError, getErr.Error())
		return
	}

	responseWithJSON(w, http.StatusCreated, UserView{
		ID:         dbUser.Id,
		Email:      dbUser.Email,
		PremiumRed: dbUser.PremiumRed,
	})
}

func (cfg *apiConfig) handlerUsersCreateV2(w http.ResponseWriter, r *http.Request) {
	type UserRequest struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := UserRequest{}
	err := decoder.Decode(&params)
	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	dbUser, createErr := cfg.dbQueries.CreateUser(r.Context(), params.Email)
	if createErr != nil {
		responseWithError(w, http.StatusInternalServerError, createErr.Error())
		return
	}

	responseWithJSON(w, http.StatusCreated, UserView{
		UUID:		dbUser.ID,
		CreatedAt:	dbUser.CreatedAt,
		UpdatedAt:	dbUser.UpdatedAt,
		Email:		dbUser.Email,
	})
}
