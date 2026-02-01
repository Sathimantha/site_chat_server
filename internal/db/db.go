package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() error {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		log.Fatal("Required environment variable DB_PATH is not set. " +
			"Please set DB_PATH in .env or in your environment.\n" +
			"Example: DB_PATH=./data/chats.db")
	}

	// Ensure the directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return err
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// Create table if not exists
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

	// Add token columns if they don't exist yet (SQLite ignores if column exists)
	DB.Exec("ALTER TABLE messages ADD COLUMN input_tokens INTEGER DEFAULT 0")
	DB.Exec("ALTER TABLE messages ADD COLUMN output_tokens INTEGER DEFAULT 0")

	log.Printf("Database initialized → %s", dbPath)
	return nil
}

func Close() {
	if DB != nil {
		_ = DB.Close()
	}
}

// SaveMessage ...
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

// GetHistory ...
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
