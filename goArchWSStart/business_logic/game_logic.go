package business_logic

// ValidateTurnTransition validates if a turn transition is allowed
// This is where you add core business rules like:
// - whether the user is in the game
// - whether it's actually their turn
// - game state checks, etc.
func ValidateTurnTransition() error {
	// Add business logic validation here
	// For now, we'll allow any transition
	// In reality, you'd check game state, player status, etc.
	return nil
}

// Note: GetCurrentTurn and AdvanceTurn are now handled through the service layer
// which orchestrates the business logic validation with data access calls.
