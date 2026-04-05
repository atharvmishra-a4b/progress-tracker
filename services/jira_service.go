package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type JiraGraphQLResponse struct {
	Data struct {
		User struct {
			Email  string  `json:"email"`
			Issues []Issue `json:"issues"`
		} `json:"user"`
	} `json:"data"`
}

type Issue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Fields struct {
		Summary string `json:"summary"`
		Created string `json:"created"`
		Updated string `json:"updated"`
		Status  struct {
			Name string `json:"name"`
		} `json:"status"`
		IssueType struct {
			Name string `json:"name"`
		} `json:"issuetype"`
		Priority *struct {
			Name string `json:"name"`
		} `json:"priority"`
		ResolutionDate *string `json:"resolutiondate"`
		Project        struct {
			Key  string `json:"key"`
			Name string `json:"name"`
		} `json:"project"`
	} `json:"fields"`
}


func FetchJiraIssues(email string, apiToken string, baseURL string) (*JiraGraphQLResponse, error) {
	jql := "assignee=currentUser() order by created desc"

	fullURL, err := buildJiraSearchURL(baseURL, jql)
	if err != nil {
		return nil, err
	}

	body, err := makeJiraRequest(fullURL, email, apiToken)
	if err != nil {
		return nil, err
	}

	var jiraResp struct {
		Issues []Issue `json:"issues"`
	}
	if err := json.Unmarshal(body, &jiraResp); err != nil {
		return nil, fmt.Errorf("failed to parse jira response: %w", err)
	}

	// Convert to our GraphQL-style response format
	response := &JiraGraphQLResponse{}
	response.Data.User.Email = maskEmail(email)
	response.Data.User.Issues = jiraResp.Issues

	return response, nil
}


func makeJiraRequest(url string, email string, apiToken string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(email, apiToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jira API error (%d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}

func buildJiraSearchURL(baseURL string, jql string) (string, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return "", fmt.Errorf("jira base URL is empty")
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid Jira base URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid Jira base URL: missing scheme or host")
	}

	hostname := parsed.Hostname()
	if parsed.Scheme != "https" && !(parsed.Scheme == "http" && (hostname == "localhost" || hostname == "127.0.0.1")) {
		return "", fmt.Errorf("insecure Jira base URL: only https is allowed")
	}

	pathURL, err := url.Parse("/rest/api/3/search/jql")
	if err != nil {
		return "", fmt.Errorf("failed to build Jira search path: %w", err)
	}

	searchURL := parsed.ResolveReference(pathURL)
	q := searchURL.Query()
	q.Set("jql", jql)
	q.Set("fields", "summary,status,issuetype,created,updated,resolutiondate,priority,project")
	q.Set("maxResults", "100")
	searchURL.RawQuery = q.Encode()

	return searchURL.String(), nil
}

func maskEmail(email string) string {
	email = strings.TrimSpace(email)
	if email == "" {
		return "jira-user"
	}

	parts := strings.Split(email, "@")
	local := parts[0]
	if local == "" {
		local = "jira-user"
	}

	return local + "@redacted.local"
}
