package service

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"checkers/data_access"

	"golang.org/x/crypto/bcrypt"
)

type regInfo struct {
	IP      string
	UA      string
	Expires time.Time
}

var (
	regMu     sync.RWMutex
	regTokens = make(map[string]regInfo)
)

func clientIP(r *http.Request) string {
	if x := r.Header.Get("X-Forwarded-For"); x != "" {
		parts := strings.Split(x, ",")
		return strings.TrimSpace(parts[0])
	}
	ip := r.RemoteAddr
	if i := strings.LastIndex(ip, ":"); i != -1 {
		return ip[:i]
	}
	return ip
}

// RegisterHandler serves registration page (GET) with a nonce and handles POST registrations
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// generate token and store associated info
		token := genRandomHex(16)
		info := regInfo{
			IP:      clientIP(r),
			UA:      r.Header.Get("User-Agent"),
			Expires: time.Now().Add(5 * time.Minute),
		}
		regMu.Lock()
		regTokens[token] = info
		regMu.Unlock()

		// read template and inject token
		b, err := ioutil.ReadFile("./static/register.html")
		if err != nil {
			http.Error(w, "could not load register page", http.StatusInternalServerError)
			return
		}
		page := strings.ReplaceAll(string(b), "{{TOKEN}}", token)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(page))
		return

	case http.MethodPost:
		// Validate token and create user
		username := r.FormValue("username")
		password := r.FormValue("password")
		token := r.FormValue("reg_token")
		if username == "" || token == "" {
			http.Error(w, "missing registration fields", http.StatusBadRequest)
			return
		}

		regMu.RLock()
		info, ok := regTokens[token]
		regMu.RUnlock()
		if !ok {
			http.Error(w, "invalid or expired token", http.StatusBadRequest)
			return
		}
		if time.Now().After(info.Expires) {
			// remove expired
			regMu.Lock()
			delete(regTokens, token)
			regMu.Unlock()
			http.Error(w, "token expired", http.StatusBadRequest)
			return
		}

		// Check client information matches
		if info.IP != clientIP(r) || info.UA != r.Header.Get("User-Agent") {
			http.Error(w, "client information mismatch", http.StatusBadRequest)
			return
		}

		// Validate password and hash it before storing
		if len(password) < 6 {
			http.Error(w, "password must be at least 6 characters", http.StatusBadRequest)
			return
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "could not process password", http.StatusInternalServerError)
			return
		}

		// Create user (DB or in-memory fallback). Pass the hashed password to storage.
		if err := data_access.CreateUser(username, string(hashed)); err != nil {
			http.Error(w, fmt.Sprintf("could not create user: %v", err), http.StatusInternalServerError)
			return
		}

		// Remove used token
		regMu.Lock()
		delete(regTokens, token)
		regMu.Unlock()

		// Create session and sign in user (base64url to fit VARCHAR(50))
		sess := genRandomBase64URL(32) // 32 bytes => 44 base64url chars
		storeSession(sess, username)

		// Persist token to account table (best-effort)
		if err := data_access.SetAccountToken(username, sess); err != nil {
			fmt.Printf("warning: failed to persist account token for %s: %v\n", username, err)
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sess,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   86400 * 7,
		})

		http.Redirect(w, r, "/lobby", http.StatusSeeOther)
		return

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
