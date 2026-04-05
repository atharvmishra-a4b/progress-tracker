package main

import (
	"crypto/subtle"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"internal-tracker/config"
	"internal-tracker/database"
	"internal-tracker/handlers"

	"github.com/gorilla/mux"
)

func main() {
	config.LoadEnv()

	database.InitDB()

	r := mux.NewRouter()
	r.Use(securityHeadersMiddleware)
	r.Use(apiKeyMiddleware(os.Getenv("INTERNAL_TRACKER_API_KEY")))

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("OK"))
	}).Methods("GET")

	r.HandleFunc("/github/stats", handlers.GetGitHubStats).Methods("GET")

	r.HandleFunc("/jira/stats", handlers.GetJiraStats).Methods("GET")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	listenAddr := "127.0.0.1:" + config.AppConfig.Port
	if bindAddr := strings.TrimSpace(os.Getenv("BIND_ADDR")); bindAddr != "" {
		listenAddr = bindAddr + ":" + config.AppConfig.Port
	}

	server := &http.Server{
		Addr:              listenAddr,
		Handler:           r,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Server running on %s", listenAddr)
	log.Fatal(server.ListenAndServe())
}

func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		next.ServeHTTP(w, r)
	})
}

func apiKeyMiddleware(apiKey string) mux.MiddlewareFunc {
	apiKey = strings.TrimSpace(apiKey)

	return func(next http.Handler) http.Handler {
		if apiKey == "" {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/github/stats" && r.URL.Path != "/jira/stats" {
				next.ServeHTTP(w, r)
				return
			}

			provided := strings.TrimSpace(r.Header.Get("X-Internal-API-Key"))
			if provided == "" {
				authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
				if strings.HasPrefix(authHeader, "Bearer ") {
					provided = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
				}
			}

			if subtle.ConstantTimeCompare([]byte(provided), []byte(apiKey)) != 1 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}