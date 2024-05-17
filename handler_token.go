package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethpalser/chirpy/internal/auth"
	"github.com/ethpalser/chirpy/internal/database"
)

type TokenView struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) handlerTokenRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.Header.Get("Authorization")
	if refreshToken == "" {
		responseWithError(w, http.StatusUnauthorized, "invalid auth token")
		return
	}

	tokenVal := strings.TrimPrefix(refreshToken, "Bearer ")
	dbToken, err := cfg.database.FindRefreshToken(tokenVal)

	if errors.Is(err, database.ErrNotExist) || time.Since(dbToken.Exp) > 0 {
		responseWithError(w, http.StatusUnauthorized, "unauthorized access")
		return
	}

	accessToken, jwtErr := auth.IssueJWT(cfg.jwtSecret, fmt.Sprint(dbToken.UserID), 3600)
	if jwtErr != nil {
		responseWithError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	responseWithJSON(w, http.StatusOK, TokenView{
		Token: accessToken,
	})
}

func (cfg *apiConfig) handlerTokenRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.Header.Get("Authorization")
	if refreshToken == "" {
		responseWithError(w, http.StatusUnauthorized, "invalid auth token")
		return
	}

	tokenVal := strings.TrimPrefix(refreshToken, "Bearer ")
	err := cfg.database.RevokeRefreshToken(tokenVal)
	if err != nil {
		responseWithError(w, http.StatusUnauthorized, "unauthorized access")
		return
	}
	responseWithJSON(w, http.StatusNoContent, nil)
}
