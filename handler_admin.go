package main

import(
	"net/http"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Fatal error: .env failed to load")	
	}

	platform := os.Getenv("PLATFORM")
	if platform != "dev" {
		responseWithError(w, http.StatusForbidden, "Forbidden 403")
		return
	}

	cfg.dbQueries.DeleteAllUsers(r.Context())
	responseWithJSON(w, http.StatusOK, nil)
}
