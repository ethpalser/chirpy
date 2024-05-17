package main

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/ethpalser/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerChirpsGetOne(w http.ResponseWriter, r *http.Request) {
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

	dbChirp, getErr := cfg.database.GetChirp(idVal)
	if getErr != nil {
		responseWithError(w, http.StatusNotFound, getErr.Error())
		return
	}

	if dbChirp.AuthorID != userID {
		responseWithError(w, http.StatusUnauthorized, "unauthorized access")
	}

	responseWithJSON(w, http.StatusOK, ChirpView{
		ID:   dbChirp.Id,
		Body: dbChirp.Message,
	})
}

func (cfg *apiConfig) handlerChirpsGetAll(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.database.GetChirps()
	if err != nil {
		responseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	chirps := []ChirpView{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, ChirpView{
			ID:   dbChirp.Id,
			Body: dbChirp.Message,
		})
	}

	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})

	responseWithJSON(w, http.StatusOK, chirps)
}
