package service

import (
	"net/http"

	"othello/business_logic"
	"othello/data_access"
)

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
