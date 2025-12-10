package service

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"othello/business_logic"
	"othello/data_access"

	"github.com/gorilla/websocket"
)

// BoardUpdate represents a board state update sent over WebSockets
type BoardUpdate struct {
	Board [8][8]string `json:"board"`
	Turn  string       `json:"turn"`
}

// GameHub coordinates all game activity for real-time board updates
type GameHub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan BoardUpdate
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.RWMutex
}

// GameHubInstance is the single global instance used by the server
var GameHubInstance = &GameHub{
	clients:    make(map[*websocket.Conn]bool),
	broadcast:  make(chan BoardUpdate),
	register:   make(chan *websocket.Conn),
	unregister: make(chan *websocket.Conn),
}

// Run is the event loop for the game hub
func (h *GameHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

			// Send current board state to new client
			board := data_access.GetBoard()
			turn := data_access.GetTurn()
			update := BoardUpdate{
				Board: board,
				Turn:  turn,
			}
			if err := client.WriteJSON(update); err != nil {
				log.Printf("Error sending initial board: %v", err)
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()

		case update := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				if err := client.WriteJSON(update); err != nil {
					log.Printf("Error broadcasting board update: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// GameWSHandler upgrades the HTTP request to a WebSocket for game updates
func GameWSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	GameHubInstance.register <- conn

	defer func() {
		GameHubInstance.unregister <- conn
	}()

	// Keep connection alive, but we don't expect incoming messages from clients
	// Clients will send moves via HTTP POST to /move endpoint
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

// MoveHandler handles player moves and broadcasts board updates
func MoveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse move data (assuming JSON with row, col, player)
	var move struct {
		Row    int    `json:"row"`
		Col    int    `json:"col"`
		Player string `json:"player"`
	}
	if err := json.NewDecoder(r.Body).Decode(&move); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate and apply move (you'll need to implement this in business_logic)
	// For now, just update the board directly
	board := data_access.GetBoard()
	if move.Row >= 0 && move.Row < 8 && move.Col >= 0 && move.Col < 8 {
		board[move.Row][move.Col] = move.Player
		data_access.SetBoard(board)
		data_access.NextTurn()
	}

	// Broadcast updated board
	updatedBoard := data_access.GetBoard()
	turn := data_access.GetTurn()
	update := BoardUpdate{
		Board: updatedBoard,
		Turn:  turn,
	}
	GameHubInstance.broadcast <- update

	jsonResponse(w, http.StatusOK, map[string]string{"status": "move accepted"})
}

func GetTurnHandler(w http.ResponseWriter, r *http.Request) {
	// Service orchestrates: fetch current turn from data access
	turn := data_access.GetTurn()
	jsonResponse(w, http.StatusOK, map[string]string{"currentTurn": turn})
}

func NextTurnHandler(w http.ResponseWriter, r *http.Request) {
	// Service orchestrates: validate business rules, then update data
	if err := business_logic.ValidateTurnTransition(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	next := data_access.NextTurn()
	jsonResponse(w, http.StatusOK, map[string]string{"nextTurn": next})
}

func BoardHandler(w http.ResponseWriter, r *http.Request) {
	// Serve the board.html file from the root directory
	http.ServeFile(w, r, "./static/board.html")
}
