package business_logic

import "fmt"

// ValidateUsername checks if a username is valid (non-empty)
func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	return nil
}

// ValidatePassword checks if a password meets minimum requirements
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}
	return nil
}

// ValidateCredentials checks both username and password
func ValidateCredentials(username, password string) error {
	if err := ValidateUsername(username); err != nil {
		return err
	}
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}
	return nil
}
