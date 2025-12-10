package data_access

import (
	"context"
	"database/sql"
	"time"
)

// ChatMessage represents a row in the chat table.
type ChatMessage struct {
	Account_Token string
	Username      string
	Message       string
	Chat_Date     time.Time
}

// returns the most recent messages (limit controlled).
func GetMessages(ctx context.Context, limit int) ([]ChatMessage, error) {
	if DB == nil {
		return nil, sql.ErrConnDone
	}
	if limit <= 0 {
		limit = 100
	}

	rows, err := DB.QueryContext(ctx,
		`SELECT account_token, username, message, chat_date
         FROM 442Chat
         ORDER BY chat_date DESC
         LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ChatMessage
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.Account_Token, &m.Username, &m.Message, &m.Chat_Date); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// InsertMessage inserts a new chat message and returns the inserted ID.
func InsertMessage(ctx context.Context, accountToken, username, message string, chatDate string) (int64, error) {
	if DB == nil {
		return 0, sql.ErrConnDone
	}
	res, err := DB.ExecContext(ctx,
		`INSERT INTO 442Chat (account_token, username, message, chat_date) VALUES (?, ?, ?, ?)`,
		accountToken, username, message, chatDate)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateMessage updates the text of an existing message.
func UpdateMessageByAccountAndDate(ctx context.Context, accountToken string, chatDate string, newText string) error {
	if DB == nil {
		return sql.ErrConnDone
	}
	_, err := DB.ExecContext(ctx, `UPDATE 442Chat SET message = ? WHERE account_token = ? AND chat_date = ?`, newText, accountToken, chatDate)
	return err
}

// DeleteMessage removes a message by id.
func DeleteMessageByAccountAndDate(ctx context.Context, accountToken string, chatDate string) error {
	if DB == nil {
		return sql.ErrConnDone
	}
	_, err := DB.ExecContext(ctx, `DELETE FROM 442Chat WHERE account_token = ? AND chat_date = ?`, accountToken, chatDate)
	return err
}
