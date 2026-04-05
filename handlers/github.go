package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"internal-tracker/services"
)

func GetGitHubStats(w http.ResponseWriter, r *http.Request) {

	token := os.Getenv("GITHUB_TOKEN")
	username := os.Getenv("GITHUB_USERNAME")

	if token == "" || username == "" {
		writeJSONError(w, http.StatusInternalServerError, "GitHub integration is not configured")
		return
	}

	data, err := services.FetchPRs(username, token)
	if err != nil {
		log.Printf("github stats fetch failed: %v", err)
		writeJSONError(w, http.StatusBadGateway, "Failed to fetch GitHub data")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
