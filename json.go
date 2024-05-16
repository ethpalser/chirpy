package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func responseWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with 5xx error: %s", msg)
	}
	type returnErr struct {
		Error string `json:"error"`
	}
	responseWithJSON(w, code, returnErr{
		Error: msg,
	})
}

func responseWithJSON(w http.ResponseWriter, code int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json, err := json.Marshal(body)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}
	w.WriteHeader(code)
	w.Write(json)
}
