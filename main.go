package main

import (
	"net/http"
)

func main() {
	// Create a multiplexer that can handle HTTP requests for a server at its endpoints
	mux := http.NewServeMux()
	// Update the multiplexer to accept CORS data
	corsMux := middlewareCors(mux)
	// Setup a server that uses the new multiplexer
	server := http.Server{Addr: "localhost:8080", Handler: corsMux}
	server.ListenAndServe()
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
