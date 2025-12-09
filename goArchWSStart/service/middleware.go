package service

import (
	"fmt"
	"net/http"
	"strings"

	"checkers/data_access"
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

		// Validate the session token against persisted Account_Token in DB
		username, ok := data_access.GetUsernameByToken(cookie.Value)
		if !ok {
			fmt.Printf("SessionMiddleware: token not found or invalid for session %s\n", cookie.Value)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"invalid or expired session"}`))
			return
		}

		// Token is valid; also verify it matches the in-memory session for consistency
		if inMemUser, ok := lookupSession(cookie.Value); ok && inMemUser != username {
			fmt.Printf("SessionMiddleware: token mismatch for %s: DB=%s, Memory=%s\n", cookie.Value, username, inMemUser)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":"session mismatch"}`))
			return
		}

		// Continue to the underlying handler
		next.ServeHTTP(w, r)
	})
}
