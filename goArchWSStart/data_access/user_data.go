package data_access

import (
	"database/sql"
	"fmt"
	"sync"
)

var (
	usersMu     sync.RWMutex
	inMemUsers  = make(map[string]string)
	tokensMu    sync.RWMutex
	inMemTokens = make(map[string]string) // token -> username
	userTokens  = make(map[string]string) // username -> token
)

// CreateUser attempts to insert into DB if available, otherwise falls back to in-memory map.
func CreateUser(username, password string) error {
	// Store the provided password value as-is. Caller is responsible for hashing.
	if DB != nil {
		result, err := DB.Exec("INSERT INTO `442Account` (Username, Password_Hashed) VALUES (?, ?)", username, password)
		if err == nil {
			rows, _ := result.RowsAffected()
			fmt.Printf("CreateUser: inserted %d rows for %s into DB\n", rows, username)
			return nil
		}
		// Log DB error but don't fall back to in-memory; fail the request
		fmt.Printf("CreateUser: DB error for %s: %v\n", username, err)
		return fmt.Errorf("database insert failed: %v", err)
	}

	// DB not configured; use in-memory fallback
	fmt.Printf("CreateUser: DB not available, using in-memory fallback for %s\n", username)
	usersMu.Lock()
	defer usersMu.Unlock()
	if _, ok := inMemUsers[username]; ok {
		return fmt.Errorf("user already exists")
	}
	inMemUsers[username] = password
	return nil
}

// GetUser returns password and presence flag.
func GetUser(username string) (string, bool) {
	if DB != nil {
		var pass string
		row := DB.QueryRow("SELECT Password_Hashed FROM `442Account` WHERE Username = ?", username)
		if err := row.Scan(&pass); err == nil {
			return pass, true
		} else if err != sql.ErrNoRows {
			// If some DB error occurred, fall back to in-memory
		}
	}

	usersMu.RLock()
	defer usersMu.RUnlock()
	p, ok := inMemUsers[username]
	return p, ok
}

// SetAccountToken saves the session token for a username. Caller provides the token string.
// It attempts to update the DB first; on DB errors it falls back to an in-memory map.
func SetAccountToken(username, token string) error {
	if DB != nil {
		result, err := DB.Exec("UPDATE `442Account` SET Account_Token = ? WHERE Username = ?", token, username)
		if err == nil {
			rows, _ := result.RowsAffected()
			fmt.Printf("SetAccountToken: updated %d rows for %s with token\n", rows, username)
			if rows == 0 {
				fmt.Printf("SetAccountToken: warning - no rows updated for user %s (user may not exist in DB)\n", username)
			}
			return nil
		}
		// Log DB error but fail the request instead of silently falling back
		fmt.Printf("SetAccountToken: DB error for %s: %v\n", username, err)
		return fmt.Errorf("database update failed: %v", err)
	}

	// DB not configured; use in-memory fallback
	fmt.Printf("SetAccountToken: DB not available, using in-memory fallback for %s\n", username)
	tokensMu.Lock()
	defer tokensMu.Unlock()
	// Remove any existing token for this user
	if old, ok := userTokens[username]; ok {
		delete(inMemTokens, old)
	}
	inMemTokens[token] = username
	userTokens[username] = token
	return nil
}

// GetUsernameByToken returns the username associated with a session token.
func GetUsernameByToken(token string) (string, bool) {
	if DB != nil {
		var username string
		row := DB.QueryRow("SELECT Username FROM `442Account` WHERE Account_Token = ?", token)
		if err := row.Scan(&username); err == nil {
			return username, true
		} else if err == sql.ErrNoRows {
			// Token not found in DB; fall back to in-memory
			fmt.Printf("GetUsernameByToken: token not found in DB, checking in-memory\n")
		} else {
			// DB error: log and fall back to in-memory
			fmt.Printf("GetUsernameByToken: DB error: %v\n", err)
		}
	}

	tokensMu.RLock()
	defer tokensMu.RUnlock()
	u, ok := inMemTokens[token]
	return u, ok
}

// ClearAccountToken removes the stored token for a username (logout).
func ClearAccountToken(username string) error {
	if DB != nil {
		if _, err := DB.Exec("UPDATE `442Account` SET Account_Token = NULL WHERE Username = ?", username); err == nil {
			return nil
		}
		// fallthrough to in-memory on DB error
	}

	tokensMu.Lock()
	defer tokensMu.Unlock()
	if t, ok := userTokens[username]; ok {
		delete(userTokens, username)
		delete(inMemTokens, t)
	}
	return nil
}
