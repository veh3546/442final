package main

import (
	"log"
	"net/http"
	"os"
	"othello/data_access"
	"othello/service"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file for local development (optional). If not present, fall back to environment variables.
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, using existing environment variables")
	}

	// Initialize DB from environment variables (DB_USER, DB_PASS, DB_HOST, DB_PORT, DB_NAME)
	db, err := data_access.NewDB(os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}
	defer db.Close()

	// start with this, to show serving up static files:
	/*
		fs := http.FileServer(http.Dir("./static"))
		http.Handle("/", fs)
		http.ListenAndServe("localhost:8080", nil)
	*/

	// Start the chat hub as a background goroutine
	go service.Hub.Run()

	// a mux (multiplexer) routes incoming requests to their respective handlers
	mux := http.NewServeMux()

	// Public endpoints (login, root)
	mux.HandleFunc("/login", service.LoginHandler)
	mux.HandleFunc("/register", service.RegisterHandler)

	// Protected endpoints (require session)
	mux.HandleFunc("/lobby", service.LobbyHandler)
	mux.HandleFunc("/me", service.MeHandler)
	mux.HandleFunc("/logout", service.LogoutHandler)

	// Protected API endpoints
	mux.HandleFunc("/turn", service.GetTurnHandler)
	// mux.HandleFunc("/next", service.NextTurnHandler)
	mux.HandleFunc("/ws/chat", service.ChatHandler)
	mux.HandleFunc("/board", service.BoardHandler)

	// Root (/) serves login page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			service.LoginHandler(w, r)
			return
		}
		http.NotFound(w, r)
	})

	// Static file server for assets (CSS, JS, images)
	// Must be registered after "/" to take precedence
	fs := http.FileServer(http.Dir("./static/assets"))
	mux.Handle("/assets/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set MIME types for common files
		if len(r.URL.Path) > 4 && r.URL.Path[len(r.URL.Path)-4:] == ".css" {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		} else if len(r.URL.Path) > 3 && r.URL.Path[len(r.URL.Path)-3:] == ".js" {
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		} else if len(r.URL.Path) > 5 && r.URL.Path[len(r.URL.Path)-5:] == ".json" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		}
		http.StripPrefix("/assets/", fs).ServeHTTP(w, r)
	}))

	// Wrap with session middleware
	protected := service.SessionMiddleware(mux)

	// If we hadn't created a custom mux to enable middleware,
	// the second param would be nil, which uses http.DefaultServeMux.
	//http.ListenAndServe("localhost:8080", protected)
	addr := ":8080"
	log.Printf("main: listening on %s", addr)
	log.Printf("main: address (quoted) = %q", addr)

	if err := http.ListenAndServe(addr, protected); err != nil {
		log.Fatalf("failed %v", err)
	}
}
