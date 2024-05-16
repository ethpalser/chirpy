package main

import (
	"net/http"
	"sort"
	"strconv"
)

func (cfg *apiConfig) handlerChirpsGetOne(w http.ResponseWriter, r *http.Request) {
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
