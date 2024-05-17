package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (cfg *apiConfig) webhookPolka(w http.ResponseWriter, r *http.Request) {
	type WebhookRequest struct {
		Event string                 `json:"event"`
		Data  map[string]interface{} `json:"data"`
	}

	accessToken := r.Header.Get("Authorization")
	if accessToken == "" {
		responseWithError(w, http.StatusUnauthorized, "invalid auth token")
		return
	}
	tokenVal := strings.TrimPrefix(accessToken, "ApiKey ")
	if tokenVal != cfg.polkaApiKey {
		responseWithError(w, http.StatusUnauthorized, "unauthorized access")
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := WebhookRequest{}
	err := decoder.Decode(&params)
	if err != nil {
		responseWithError(w, http.StatusInternalServerError, "something went wrong: failed reading request")
		return
	}

	if params.Event != "user.upgraded" {
		responseWithJSON(w, http.StatusNoContent, nil)
		return
	}

	dataUserID, ok := params.Data["user_id"]
	if !ok {
		responseWithError(w, http.StatusBadRequest, "invalid request: missing user_id")
		return
	}

	// Numbers are decoded into float64
	floatUserID, ok := dataUserID.(float64)
	if !ok {
		responseWithError(w, http.StatusBadRequest, "invalid request: user_id is not a number")
		return
	}

	updErr := cfg.database.UpdateUserPremiumRed(int(floatUserID), true)
	if updErr != nil {
		responseWithError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	responseWithJSON(w, http.StatusNoContent, nil)
}
