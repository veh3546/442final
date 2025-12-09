package business_logic

import (
	"checkers/data_access"

	"golang.org/x/crypto/bcrypt"
)

// VerifyCredentials checks if the provided username and password match the stored hash
func VerifyCredentials(username, password string) (bool, error) {
	storedHash, ok := data_access.GetUser(username)
	if !ok {
		// User does not exist
		return false, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
		// Password mismatch
		return false, nil
	}

	return true, nil
}

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// GenerateSessionToken creates a new session token
func GenerateSessionToken() string {
	return GenRandomBase64URL(32) // 32 bytes => 44 base64url chars
}

// GenerateRegistrationToken creates a registration token
func GenerateRegistrationToken() string {
	return GenRandomHex(16)
}
