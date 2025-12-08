package business_logic

import "checkers/data_access"

func GetCurrentTurn() string {
	// add validation here, like whether the user is in the game
	return data_access.GetTurn()
}

func AdvanceTurn() string {
	// add validation here, like whether it's the user's turn and can advance the turn
	return data_access.NextTurn()
}
