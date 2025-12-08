package service

import (
	"net/http"

	"checkers/business_logic"
)

func GetTurnHandler(w http.ResponseWriter, r *http.Request) {
	turn := business_logic.GetCurrentTurn()
	jsonResponse(w, http.StatusOK, map[string]string{"currentTurn": turn})
}

func NextTurnHandler(w http.ResponseWriter, r *http.Request) {
	next := business_logic.AdvanceTurn()
	jsonResponse(w, http.StatusOK, map[string]string{"nextTurn": next})
}
