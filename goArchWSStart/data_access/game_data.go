package data_access

var players = []string{"Alice", "Bob", "Charlie"}
var currentIndex = 0

// Query the database to figure out whose turn it is
func GetTurn() string {
	return players[currentIndex]
}

// Update the database to advance to the next player's turn
func NextTurn() string {
	currentIndex = (currentIndex + 1) % len(players)
	return players[currentIndex]
}
