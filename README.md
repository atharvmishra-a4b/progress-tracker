# Internal Tracker

Internal Tracker is a Go web application that provides a unified dashboard for:

- GitHub pull request analytics
- Jira issue analytics

It serves a static frontend and exposes backend API endpoints that fetch data from GitHub and Jira.

## Features

- GitHub analytics dashboard (`/index.html`)
- Jira analytics dashboard (`/jira.html`)
- Health endpoint (`/health`)
- Optional API-key protection for analytics endpoints
- Security headers and HTTP server timeouts
- Safer upstream API handling (timeouts, response checks)

## Tech Stack

- Go (net/http)
- Gorilla Mux router
- Vanilla HTML/CSS/JS frontend
- Chart.js for charts
- Tailwind CSS (CDN)

## Project Structure

```text
.
├── main.go
├── config/
│   └── config.go
├── database/
│   └── db.go
├── handlers/
│   ├── github.go
│   └── jira.go
├── services/
│   ├── github_service.go
│   └── jira_service.go
├── static/
│   ├── index.html
│   └── jira.html
├── .env.example
├── .gitignore
├── go.mod
└── go.sum
```

## Getting Started

### Prerequisites

- Go 1.25+
- A GitHub personal access token with permission to read the target user/org data
- A Jira API token and Jira base URL

### Setup

1. Clone the repository.
2. Copy environment template:

   ```bash
   cp .env.example .env
   ```

3. Fill in required values in `.env`.
4. Start the server:

   ```bash
   go run main.go
   ```

5. Open:
   - `http://127.0.0.1:8080/index.html`
   - `http://127.0.0.1:8080/jira.html`

## Configuration

Environment variables (see `.env.example`):

- `PORT` (default: `8080`)
- `BIND_ADDR` (default: `127.0.0.1`)
- `BASE_URL`
- `INTERNAL_TRACKER_API_KEY` (optional)

GitHub:

- `GITHUB_TOKEN`
- `GITHUB_USERNAME`

Jira:

- `JIRA_EMAIL`
- `JIRA_API_TOKEN`
- `JIRA_BASE_URL` (must be HTTPS, except localhost for development)

Other:

- `DATABASE_URL` (reserved for future DB usage)

## API Endpoints

- `GET /health` → plain text `OK`
- `GET /github/stats` → GitHub analytics payload
- `GET /jira/stats` → Jira analytics payload

If `INTERNAL_TRACKER_API_KEY` is set, both analytics endpoints require one of:

- `X-Internal-API-Key: <value>`
- `Authorization: Bearer <value>`

## Optional Frontend API Key Support

The frontend can send an API key from browser storage (`localStorage` key: `internalTrackerApiKey`).

Example in browser console:

```js
localStorage.setItem('internalTrackerApiKey', 'your-key')
```

## Security Notes

This project includes:

- Security-related HTTP headers (CSP, frame protections, referrer policy, etc.)
- Server read/write/header/idle timeouts
- Upstream request timeouts for GitHub and Jira
- Generic error responses to clients (avoids leaking upstream/internal details)
- Basic XSS hardening in dynamic UI rendering

Before publishing or deploying:

- Do not commit `.env`
- Rotate all credentials if they were ever exposed
- Restrict network exposure (bind to localhost or private network)
- Use HTTPS and a reverse proxy in production

## Development

Run formatting and build checks:

```bash
gofmt -w main.go handlers/*.go services/*.go config/*.go database/*.go
go build ./...
```

## Troubleshooting

- `401 Unauthorized` on analytics endpoints:
  - Set `INTERNAL_TRACKER_API_KEY` correctly and provide matching request header.
- `Failed to fetch GitHub data`:
  - Verify `GITHUB_TOKEN` and `GITHUB_USERNAME`.
- `Failed to fetch Jira data`:
  - Verify `JIRA_EMAIL`, `JIRA_API_TOKEN`, and `JIRA_BASE_URL`.
  - Ensure Jira URL is valid and reachable.


