package service

import (
	"fmt"
	"net/http"
	"sync"

	"checkers/business_logic"
	"checkers/data_access"
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

		// Require non-empty username and password
		if username == "" || password == "" {
			http.Error(w, "username and password required", http.StatusBadRequest)
			return
		}

		// Verify credentials using business logic
		valid, err := business_logic.VerifyCredentials(username, password)
		if err != nil || !valid {
			http.Error(w, "invalid username or password", http.StatusUnauthorized)
			return
		}

		// Create a secure session token and store mapping (base64url, fits VARCHAR(50))
		token := business_logic.GenerateSessionToken()
		storeSession(token, username)

		// Persist token to account table (best-effort)
		if err := data_access.SetAccountToken(username, token); err != nil {
			fmt.Printf("warning: failed to persist account token for %s: %v\n", username, err)
		}

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
		// Lookup username for this session, then remove from session store
		if username, ok := lookupSession(cookie.Value); ok {
			// Best-effort: clear persisted account token
			if err := data_access.ClearAccountToken(username); err != nil {
				fmt.Printf("warning: failed to clear account token for %s: %v\n", username, err)
			}
		}

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
