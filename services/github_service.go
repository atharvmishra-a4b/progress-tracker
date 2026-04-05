package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GraphQLResponse struct {
	Data struct {
		User struct {
			Login        string `json:"login"`
			PullRequests struct {
				Nodes []struct {
					Title       string  `json:"title"`
					State       string  `json:"state"`
					CreatedAt   string  `json:"createdAt"`
					BaseRefName string  `json:"baseRefName"`
					MergedAt    *string `json:"mergedAt"`
					Repository  struct {
						NameWithOwner string `json:"nameWithOwner"`
					} `json:"repository"`
				} `json:"nodes"`
			} `json:"pullRequests"`
		} `json:"user"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func FetchPRs(username string, token string) (*GraphQLResponse, error) {
	query := `
	query($login: String!) {
	  user(login: $login) {
	    login
	    pullRequests(first: 100, orderBy: {field: CREATED_AT, direction: DESC}) {
	      nodes {
	        title
	        state
	        createdAt
	        baseRefName
	        mergedAt
	        repository {
	          nameWithOwner
	        }
	      }
	    }
	  }
	}`

	bodyMap := map[string]interface{}{
		"query": query,
		"variables": map[string]string{
			"login": username,
		},
	}

	jsonBody, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode GitHub GraphQL body: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (%d): %s", resp.StatusCode, string(body))
	}

	var gqlResp GraphQLResponse
	err = json.NewDecoder(resp.Body).Decode(&gqlResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode GitHub response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GitHub GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	return &gqlResp, nil
}