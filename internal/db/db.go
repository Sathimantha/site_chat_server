package db

import (
	"database/sql"
	"log"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// Init creates the database file and table if they don't exist
func Init() error {
	dbPath := filepath.Join(".", "chats.db")
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// Create table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id      TEXT NOT NULL,
			role            TEXT NOT NULL,
			content         TEXT NOT NULL,
			timestamp       DATETIME DEFAULT CURRENT_TIMESTAMP,
			input_tokens    INTEGER DEFAULT 0,
			output_tokens   INTEGER DEFAULT 0
		)
	`)
	if err != nil {
		return err
	}

	// Try to add columns if they are missing (SQLite ignores if already exists)
	DB.Exec("ALTER TABLE messages ADD COLUMN input_tokens INTEGER DEFAULT 0")
	DB.Exec("ALTER TABLE messages ADD COLUMN output_tokens INTEGER DEFAULT 0")

	log.Printf("Database initialized → %s", dbPath)
	return nil
}

// Close cleans up the database connection
func Close() {
	if DB != nil {
		_ = DB.Close()
	}
}

// SaveMessage stores a chat message with token usage
func SaveMessage(sessionID, role, content string, inputTokens, outputTokens int) error {
	stmt, err := DB.Prepare(`
		INSERT INTO messages (session_id, role, content, input_tokens, output_tokens)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(sessionID, role, content, inputTokens, outputTokens)
	return err
}

// GetHistory returns all messages for a session, ordered by time
func GetHistory(sessionID string) ([]map[string]string, error) {
	rows, err := DB.Query(`
		SELECT role, content
		FROM messages
		WHERE session_id = ?
		ORDER BY id ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []map[string]string

	for rows.Next() {
		var role, content string
		if err := rows.Scan(&role, &content); err != nil {
			return nil, err
		}
		history = append(history, map[string]string{
			"role":    role,
			"content": content,
		})
	}

	return history, rows.Err()
}
