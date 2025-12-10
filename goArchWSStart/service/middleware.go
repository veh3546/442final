package service

import (
	"fmt"
	"net/http"
	"strings"

	"othello/data_access"
)

// SessionMiddleware ensures that requests have a non-empty 'session' cookie
// except for the /login, /register endpoint and static asset/root page loads so that
// the browser can first obtain a session cookie.
func SessionMiddleware(next http.Handler) http.Handler {
	// asterisk before the param type means we are returning a pointer to that type, not a copy
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// Allow unauthenticated access for login, register, logout, root (serves login), and static assets
		if path == "/login" || path == "/register" || path == "/logout" || path == "/" || strings.HasPrefix(path, "/assets/") {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"missing or invalid session"}`))
			return
		}

		// Validate the session token - try database first, fall back to in-memory
		fmt.Printf("SessionMiddleware: checking token %s\n", cookie.Value)
		username, dbOk := data_access.GetUsernameByToken(cookie.Value)
		inMemUser, memOk := lookupSession(cookie.Value)
		fmt.Printf("SessionMiddleware: DB result - username: %s, found: %v\n", username, dbOk)
		fmt.Printf("SessionMiddleware: Memory result - username: %s, found: %v\n", inMemUser, memOk)

		// If neither database nor in-memory has the token, reject
		if !dbOk && !memOk {
			fmt.Printf("SessionMiddleware: token not found in DB or memory for session %s\n", cookie.Value)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"invalid or expired session"}`))
			return
		}

		// If both are available, ensure they match for consistency
		if dbOk && memOk && username != inMemUser {
			fmt.Printf("SessionMiddleware: token mismatch for %s: DB=%s, Memory=%s\n", cookie.Value, username, inMemUser)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"session mismatch"}`))
			return
		}

		// Use whichever source has the username
		if !dbOk && memOk {
			username = inMemUser
		}

		// Continue to the underlying handler
		next.ServeHTTP(w, r)
	})
}
