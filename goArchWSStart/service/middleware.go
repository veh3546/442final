package service

import (
	"net/http"
	"strings"
)

// SessionMiddleware ensures that requests have a non-empty 'session' cookie
// except for the /login endpoint and static asset/root page loads so that
// the browser can first obtain a session cookie.
func SessionMiddleware(next http.Handler) http.Handler {
	// asterisk before the param type means we are returning a pointer to that type, not a copy
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// Allow unauthenticated access for login, logout, root (serves login), and static assets
		if path == "/login" || path == "/logout" || path == "/" || strings.HasPrefix(path, "/assets/") {
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

		// here there should probably be a call to a biz layer function to validate the session...

		// Continue to the underlying handler
		next.ServeHTTP(w, r)
	})
}
