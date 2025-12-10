package business_logic

import (
	"othello/data_access"
)

// listUsers retrieves all users that are online
func ListUsers() ([]string, error) {
	userlist, err := data_access.OnlineUsers()
	if err != nil {
		return nil, err
	} else {
		var users []string
		for _, u := range userlist {
			users = append(users, u.Username)
		}
		return users, nil
	}
}
