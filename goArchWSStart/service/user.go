package service

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// In-memory session store: maps session token -> username
var (
	sessionsMu sync.RWMutex
	sessions   = make(map[string]string)
)

func storeSession(token, username string) {
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	sessions[token] = username
}

func lookupSession(token string) (string, bool) {
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	u, ok := sessions[token]
	return u, ok
}

// LoginHandler serves the login form (GET) and processes login (POST)
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Simple validation: require non-empty username (in production, check against DB)
		if username == "" && password == "" {
			http.Error(w, "Username and password required", http.StatusBadRequest)
			return
		}

		// Create a session token and store mapping
		token := fmt.Sprintf("session-%d", time.Now().UnixNano())
		storeSession(token, username)

		// Set session cookie (HttpOnly for security)
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   86400 * 7, // 7 days
		})

		// Redirect to lobby
		http.Redirect(w, r, "/lobby", http.StatusSeeOther)
		return
	}

	// GET: Serve login form from static assets
	http.ServeFile(w, r, "./static/login.html")
}

// MeHandler returns the current user's info based on the session cookie
func MeHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"missing or invalid session"}`))
		return
	}

	if username, ok := lookupSession(cookie.Value); ok {
		jsonResponse(w, http.StatusOK, map[string]string{"username": username})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"invalid session"}`))
}

// LogoutHandler clears the session server-side and deletes the cookie
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil && cookie.Value != "" {
		// Remove from server-side session store
		sessionsMu.Lock()
		delete(sessions, cookie.Value)
		sessionsMu.Unlock()
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	jsonResponse(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// LobbyHandler serves the lobby page (protected by session middleware)
func LobbyHandler(w http.ResponseWriter, r *http.Request) {
	// Serve the lobby HTML file from static directory
	http.ServeFile(w, r, "./static/index.html")
}
