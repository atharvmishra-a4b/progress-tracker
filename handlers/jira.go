package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"internal-tracker/services"
)

func GetJiraStats(w http.ResponseWriter, r *http.Request) {
	email := os.Getenv("JIRA_EMAIL")
	apiToken := os.Getenv("JIRA_API_TOKEN")
	baseURL := os.Getenv("JIRA_BASE_URL")

	if email == "" || apiToken == "" || baseURL == "" {
		writeJSONError(w, http.StatusInternalServerError, "Jira integration is not configured")
		return
	}

	data, err := services.FetchJiraIssues(email, apiToken, baseURL)
	if err != nil {
		log.Printf("jira stats fetch failed: %v", err)
		writeJSONError(w, http.StatusBadGateway, "Failed to fetch Jira data")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}