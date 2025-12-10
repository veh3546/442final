package data_access

var players = []string{"Alice", "Bob", "Charlie"}
var currentIndex = 0

// Board represents the 8x8 game board
var board [8][8]string

// Initialize the board with starting positions
func init() {
	// Initialize empty board
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			board[i][j] = ""
		}
	}
	// Set starting pieces
	board[3][3] = "white"
	board[3][4] = "black"
	board[4][3] = "black"
	board[4][4] = "white"
}

// GetBoard returns the current board state
func GetBoard() [8][8]string {
	return board
}

// SetBoard updates the board state
func SetBoard(newBoard [8][8]string) {
	board = newBoard
}

// Query the database to figure out whose turn it is
func GetTurn() string {
	return players[currentIndex]
}

// Update the database to advance to the next player's turn
func NextTurn() string {
	currentIndex = (currentIndex + 1) % len(players)
	return players[currentIndex]
}
