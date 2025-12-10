package service

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"checkers/data_access"

	"github.com/gorilla/websocket"
)

// upgrader converts an incoming HTTP request to a WebSocket connection.
// CheckOrigin returns true to allow all origins during local development.
// In production, tighten this to validate r.Origin against allowed hosts.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for local development
	},
}

// ChatMessage is the payload exchanged over WebSockets.
// It currently carries the raw message text and an ISO timestamp string.
type ChatMessage struct {
	Account_Token string `json:"account_token,omitempty"`
	Username      string `json:"username,omitempty"`
	Message       string `json:"message"`
	Time          string `json:"time,omitempty"`
}

// ChatHub coordinates all chat activity.
//
// Concurrency model:
// - clients: set of active WebSocket connections (guarded by mu)
// - register/unregister: channels to add/remove clients (serialized by Run loop)
// - broadcast: channel to fan messages out to all connected clients
// - messages: in-memory history; appended to on each broadcast (you'd replace with DB table)
// - mu: protects both clients and messages across goroutines
type ChatHub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan ChatMessage
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	messages   []ChatMessage
	users      []User
	mu         sync.RWMutex
}

// Hub is the single global instance used by the server.
var Hub = &ChatHub{
	clients:    make(map[*websocket.Conn]bool),
	broadcast:  make(chan ChatMessage),
	register:   make(chan *websocket.Conn),
	unregister: make(chan *websocket.Conn),
	messages:   make([]ChatMessage, 0),
	users:      make([]User, 0),
}

// Run is the event loop for the hub. It should be started once (e.g., in main)
// and runs forever, handling client registration, unregistration, and message
// broadcasts. All mutations of hub state happen in this loop (with appropriate
// locking) to avoid data races.
//
// Syntax note: (h *ChatHub) is a "receiver" – it makes Run a method on the ChatHub type.
// This means you call it as Hub.Run() rather than Run(Hub). The 'h' is like 'self' or 'this'
// in other languages, giving the method access to the ChatHub instance's fields.
func (h *ChatHub) Run() {
	for {
		select {

		// A new client has connected.
		//
		// h.register is a channel of *websocket.Conn values; this of it as a message queue.
		// The left arrow means to dequeue a value from that channel when one is available.
		// In this case, it means that a new client has connected.
		case client := <-h.register:
			// The Go Mutex lock ensures that the clients map is safely modified.
			h.mu.Lock()
			h.clients[client] = true
			// Release the Mutex lock after modification.
			h.mu.Unlock()

			// Send chat history to the new client. Try DB first, fall back to in-memory.
			ctx := context.Background()
			msgs, err := data_access.GetMessages(ctx, 100)
			if err == nil {
				// DB returns messages newest-first; send them oldest-first to clients
				for i := len(msgs) - 1; i >= 0; i-- {
					dm := msgs[i]
					sm := ChatMessage{
						Account_Token: dm.Account_Token,
						Username:      dm.Username,
						Message:       dm.Message,
						Time:          dm.Chat_Date.Format(time.RFC3339),
					}
					if err := client.WriteJSON(sm); err != nil {
						log.Printf("Error sending history: %v", err)
					}
				}
			} else {
				h.mu.RLock()
				for _, msg := range h.messages {
					if err := client.WriteJSON(msg); err != nil {
						log.Printf("Error sending history: %v", err)
					}
				}
				h.mu.RUnlock()
			}

		// A client disconnected or errored.
		//
		// Remove the client from the hub and close the connection.
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()

		// A client sent a message to broadcast to all other clients.
		//
		// Dequeue the message from the broadcast channel, store it in the
		// in-memory history, and fan it out to all connected clients.
		case message := <-h.broadcast:
			// Sanitize `message` here to prevent XSS

			// Store message in memory and persist to DB
			h.mu.Lock()
			h.messages = append(h.messages, message)
			h.mu.Unlock()

			// Persist to DB (best-effort; log errors)
			go func(m ChatMessage) {
				ctx := context.Background()
				// convert time if provided, otherwise DB will set timestamp
				if _, err := data_access.InsertMessage(ctx, m.Account_Token, m.Username, m.Message, m.Time); err != nil {
					log.Printf("Error inserting chat message: %v", err)
				}
			}(message)

			// Broadcast to all connected clients. If a client write fails,
			// close and drop that client to avoid leaking dead connections.
			h.mu.RLock()
			for client := range h.clients {
				if err := client.WriteJSON(message); err != nil {
					log.Printf("Error broadcasting: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *UserHub) Run() {
	for {
		select {

		// A new client has connected.
		//
		// h.register is a channel of *websocket.Conn values; this of it as a message queue.
		// The left arrow means to dequeue a value from that channel when one is available.
		// In this case, it means that a new client has connected.
		case client := <-h.register:
			// The Go Mutex lock ensures that the clients map is safely modified.
			h.mu.Lock()
			h.clients[client] = true
			// Release the Mutex lock after modification.
			h.mu.Unlock()

			// Send chat history to the new client. Try DB first, fall back to in-memory.
			usrs, err := data_access.GetOnlineUsers()
			if err == nil {
				// DB returns messages newest-first; send them oldest-first to clients
				for i := len(usrs) - 1; i >= 0; i-- {
					dm := usrs[i]
					sm := User{
						Account_Token: dm.Account_Token,
						Username:      dm.Username,
					}
					if err := client.WriteJSON(sm); err != nil {
						log.Printf("Error sending history: %v", err)
					}
				}
			} else {
				h.mu.RLock()
				for _, msg := range h.users {
					if err := client.WriteJSON(msg); err != nil {
						log.Printf("Error sending history: %v", err)
					}
				}
				h.mu.RUnlock()
			}

		// A client disconnected or errored.
		//
		// Remove the client from the hub and close the connection.
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
		}
	}
}

// ChatHandler upgrades the HTTP request to a WebSocket and then pumps
// incoming messages from that client into the hub's broadcast channel.
// Lifecycle:
// 1) Upgrade to WebSocket
// 2) Register client with hub (triggers history replay)
// 3) Loop reading JSON ChatMessage values and forward to hub
// 4) On error/close, unregister client
func ChatHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Capture session info from the initial HTTP request so we can attach
	// username/account token to messages sent over this WebSocket.
	var sessToken string
	var sessUser string
	if c, err := r.Cookie("session"); err == nil {
		sessToken = c.Value
		if u, ok := lookupSession(sessToken); ok {
			sessUser = u
		}
	}

	Hub.register <- conn

	// The defer keyword delays execution of the function until the surrounding
	// function (ChatHandler) returns. Here, it ensures that the client is
	// unregistered from the hub when this function exits (e.g., on error or close)
	// and that resources are cleaned up.
	defer func() {
		Hub.unregister <- conn
	}()

	for {
		var msg ChatMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Enrich message with server-side session info if available.
		if sessUser != "" {
			msg.Username = sessUser
		}
		if sessToken != "" {
			msg.Account_Token = sessToken
		}
		if msg.Time == "" {
			msg.Time = time.Now().UTC().Format(time.RFC3339)
		}

		Hub.broadcast <- msg
	}
}
func UserListHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Capture session info from the initial HTTP request so we can attach
	// username/account token to messages sent over this WebSocket.
	var sessToken string
	var sessUser string
	if c, err := r.Cookie("session"); err == nil {
		sessToken = c.Value
		if u, ok := lookupSession(sessToken); ok {
			sessUser = u
		}
	}

	Hub.register <- conn

	// The defer keyword delays execution of the function until the surrounding
	// function (ChatHandler) returns. Here, it ensures that the client is
	// unregistered from the hub when this function exits (e.g., on error or close)
	// and that resources are cleaned up.
	defer func() {
		Hub.unregister <- conn
	}()

	for {
		var msg ChatMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Enrich message with server-side session info if available.
		if sessUser != "" {
			msg.Username = sessUser
		}
		if sessToken != "" {
			msg.Account_Token = sessToken
		}
		if msg.Time == "" {
			msg.Time = time.Now().UTC().Format(time.RFC3339)
		}

		Hub.broadcast <- msg
	}
}

// GetChatHistoryHandler returns the current in-memory message history as JSON.
// This can be useful for non-WebSocket clients or debugging.
func GetChatHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// Optional `limit` query parameter, default 100
	limit := 100
	if q := r.URL.Query().Get("limit"); q != "" {
		if v, err := strconv.Atoi(q); err == nil && v > 0 {
			limit = v
		}
	}

	ctx := r.Context()
	msgs, err := data_access.GetMessages(ctx, limit)
	if err != nil {
		// Fall back to in-memory history if DB read fails
		log.Printf("DB chat history read failed: %v; returning in-memory messages", err)
		Hub.mu.RLock()
		defer Hub.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Hub.messages)
		return
	}

	// Map DB messages to wire ChatMessage
	out := make([]ChatMessage, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, ChatMessage{
			Account_Token: m.Account_Token,
			Username:      m.Username,
			Message:       m.Message,
			Time:          m.Chat_Date.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}
