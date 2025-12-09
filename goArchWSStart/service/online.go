package service

import (
	"sync"

	"github.com/gorilla/websocket"
)

type User struct {
	Account_Token string `json:"account_token,omitempty"`
	Username      string `json:"username,omitempty"`
}

// ChatHub coordinates all chat activity.
//
// Concurrency model:
// - clients: set of active WebSocket connections (guarded by mu)
// - register/unregister: channels to add/remove clients (serialized by Run loop)
// - broadcast: channel to fan messages out to all connected clients
// - messages: in-memory history; appended to on each broadcast (you'd replace with DB table)
// - mu: protects both clients and messages across goroutines
type UserHub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan User
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	users      []User
	mu         sync.RWMutex
}
