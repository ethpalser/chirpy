package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ethpalser/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
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

	pathVal := r.PathValue("chirpID")
	idVal, atoiErr := strconv.Atoi(pathVal)
	if atoiErr != nil {
		responseWithError(w, http.StatusBadRequest, atoiErr.Error())
		return
	}

	// Fetch and Authorize
	dbChirp, getErr := cfg.database.GetChirp(idVal)
	if getErr != nil || dbChirp.AuthorID != userID {
		responseWithError(w, http.StatusForbidden, "forbidden")
		return
	}

	// Delete
	delErr := cfg.database.DeleteChirp(idVal)
	if delErr != nil {
		responseWithError(w, http.StatusNotFound, delErr.Error())
		return
	}

	responseWithJSON(w, http.StatusNoContent, nil)
}
