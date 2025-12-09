package business_logic

import (
	"checkers/data_access"
)

// listUsers retrieves all users that are online
func ListUsers() ([]string, error) {
	return data_access.OnlineUsers()
}
