package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Memory represents conversation memory manager
type Memory struct {
	conn        *sql.DB
	maxMessages int
}

// Message represents a chat message
type Message struct {
	ID        int       `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Metadata  string    `json:"metadata"`
	Timestamp time.Time `json:"timestamp"`
}

// Session represents a conversation session
type Session struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Platform  string    `json:"platform"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewMemory creates a new Memory instance
func NewMemory(dbPath string, maxMessages int) (*Memory, error) {
	// Create database file if doesn't exist
	_, err := os.Create(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create database file: %w", err)
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	mem := &Memory{
		conn:        conn,
		maxMessages: maxMessages,
	}

	err = mem.initDB()
	if err != nil {
		return nil, err
	}

	return mem, nil
}

// initDB initializes database schema
func (m *Memory) initDB() error {
	// Create messages table
	_, err := m.conn.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			metadata TEXT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create messages table: %w", err)
	}

	// Create index on messages
	_, err = m.conn.Exec(`
		CREATE INDEX IF NOT EXISTS idx_messages_session_timestamp
		ON messages(session_id, timestamp)
	`)
	if err != nil {
		return fmt.Errorf("failed to create messages index: %w", err)
	}

	// Create sessions table
	_, err = m.conn.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			name TEXT,
			platform TEXT,
			user_id TEXT,
			metadata TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create sessions table: %w", err)
	}

	// Create long_term_memory table
	_, err = m.conn.Exec(`
		CREATE TABLE IF NOT EXISTS long_term_memory (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT UNIQUE NOT NULL,
			value TEXT NOT NULL,
			importance INTEGER DEFAULT 1,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create long_term_memory table: %w", err)
	}

	return nil
}

// AddMessage adds a message to memory
func (m *Memory) AddMessage(sessionID, role, content string, metadata map[string]interface{}) (int64, error) {
	metadataJSON, _ := json.Marshal(metadata)

	result, err := m.conn.Exec(`
		INSERT INTO messages (session_id, role, content, metadata)
		VALUES (?, ?, ?, ?)
	`, sessionID, role, content, string(metadataJSON))
	if err != nil {
		return 0, fmt.Errorf("failed to insert message: %w", err)
	}

	id, _ := result.LastInsertId()

	// Update session timestamp
	_, err = m.conn.Exec(`
		UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, sessionID)
	if err != nil {
		return 0, fmt.Errorf("failed to update session timestamp: %w", err)
	}

	return id, nil
}

// GetMessages retrieves messages for a session
func (m *Memory) GetMessages(sessionID string, limit int) ([]Message, error) {
	var query string
	var args []interface{}

	if limit > 0 {
		query = `SELECT id, session_id, role, content, metadata, timestamp 
		         FROM messages WHERE session_id = ? 
		         ORDER BY timestamp DESC LIMIT ?`
		args = []interface{}{sessionID, limit}
	} else {
		query = `SELECT id, session_id, role, content, metadata, timestamp 
		         FROM messages WHERE session_id = ? 
		         ORDER BY timestamp DESC`
		args = []interface{}{sessionID}
	}

	rows, err := m.conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var metadata string
		err := rows.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &metadata, &msg.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		msg.Metadata = metadata
		messages = append(messages, msg)
	}

	return messages, nil
}

// SetLongTerm stores information in long-term memory
func (m *Memory) SetLongTerm(key, value string, importance int) error {
	_, err := m.conn.Exec(`
		INSERT OR REPLACE INTO long_term_memory (key, value, importance, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, key, value, importance)
	if err != nil {
		return fmt.Errorf("failed to set long-term memory: %w", err)
	}
	return nil
}

// GetLongTerm retrieves information from long-term memory
func (m *Memory) GetLongTerm(key string) (string, error) {
	var value string
	err := m.conn.QueryRow(`
		SELECT value FROM long_term_memory WHERE key = ?
	`, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed to get long-term memory: %w", err)
	}
	return value, nil
}

// CreateSession creates or updates a session
func (m *Memory) CreateSession(id, name, platform, userID string) error {
	metadataJSON, _ := json.Marshal(map[string]interface{}{})

	_, err := m.conn.Exec(`
		INSERT OR REPLACE INTO sessions (id, name, platform, user_id, metadata)
		VALUES (?, ?, ?, ?, ?)
	`, id, name, platform, userID, string(metadataJSON))
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// GetSession retrieves session information
func (m *Memory) GetSession(id string) (*Session, error) {
	var session Session
	err := m.conn.QueryRow(`
		SELECT id, name, platform, user_id, created_at, updated_at
		FROM sessions WHERE id = ?
	`, id).Scan(&session.ID, &session.Name, &session.Platform, &session.UserID, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &session, nil
}

// Close closes the database connection
func (m *Memory) Close() error {
	if m.conn != nil {
		return m.conn.Close()
	}
	return nil
}

// TestMemory runs tests on the memory module
func TestMemory() {
	log.Println("Testing Memory module...")

	mem, err := NewMemory("test_memory.db", 100)
	if err != nil {
		log.Fatalf("Failed to create memory: %v", err)
	}
	defer mem.Close()

	// Create session
	err = mem.CreateSession("test_session", "Test User", "test", "user123")
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	log.Println("✓ Session created")

	// Add messages
	_, err = mem.AddMessage("test_session", "user", "Hello QuickBot!", nil)
	if err != nil {
		log.Fatalf("Failed to add message: %v", err)
	}
	_, err = mem.AddMessage("test_session", "assistant", "Hi there!", nil)
	if err != nil {
		log.Fatalf("Failed to add message: %v", err)
	}
	log.Println("✓ Messages added")

	// Get messages
	messages, err := mem.GetMessages("test_session", 10)
	if err != nil {
		log.Fatalf("Failed to get messages: %v", err)
	}
	log.Printf("✓ Retrieved %d messages", len(messages))

	// Set long-term memory
	err = mem.SetLongTerm("user_name", "Alice", 2)
	if err != nil {
		log.Fatalf("Failed to set long-term memory: %v", err)
	}
	log.Println("✓ Long-term memory set")

	// Get long-term memory
	value, err := mem.GetLongTerm("user_name")
	if err != nil {
		log.Fatalf("Failed to get long-term memory: %v", err)
	}
	log.Printf("� Retrieved long-term memory: %s", value)

	// Cleanup
	os.Remove("test_memory.db")
	log.Println("✓ Memory module tests passed")
}
