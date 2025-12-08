package business_logic

// LogInUser returns a hard-coded session token.
// In a real application this would validate user credentials against the database
// and generate a cryptographically secure random token, persisted server-side.
func LogInUser() string {
	return "demo-session-token-12345"
}
