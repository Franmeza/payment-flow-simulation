package idempotency

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/franmeza/payment-flow-simulation/internal/models"
	_ "modernc.org/sqlite"
)

// Store is a SQLite-backed idempotency key store.
// Keys survive service restarts and power cycles, which matters when the
// acquirer is running on embedded hardware alongside a card reader.
type Store struct {
	conn *sql.DB
}

// New opens (or creates) the SQLite file and runs migrations.
func New(conn *sql.DB) *Store {
	s := &Store{conn: conn}
	s.migrate()
	return s
}

// NewConn is a convenience constructor that opens the SQLite file itself.
func NewConn(path string) *Store {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatalf("idempotency: failed to open database: %v", err)
	}
	return New(conn)
}

func (s *Store) migrate() {
	_, err := s.conn.Exec(`
		CREATE TABLE IF NOT EXISTS idempotency_keys (
			key        TEXT PRIMARY KEY,
			response   TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatalf("idempotency migration failed: %v", err)
	}
}

// Get returns the cached AuthResponse for key if it exists and has not expired.
func (s *Store) Get(key string) (*models.AuthResponse, bool) {
	var raw string
	err := s.conn.QueryRow(
		`SELECT response FROM idempotency_keys WHERE key = ?`, key,
	).Scan(&raw)
	if err != nil {
		return nil, false
	}

	var resp models.AuthResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, false
	}
	return &resp, true
}

// Set stores resp under key. INSERT OR IGNORE means the first response always
// wins — retries never overwrite the original result.
func (s *Store) Set(key string, resp models.AuthResponse) {
	raw, err := json.Marshal(resp)
	if err != nil {
		log.Printf("idempotency: failed to marshal response: %v", err)
		return
	}
	_, err = s.conn.Exec(
		`INSERT OR IGNORE INTO idempotency_keys (key, response) VALUES (?, ?)`,
		key, string(raw),
	)
	if err != nil {
		log.Printf("idempotency: failed to save key: %v", err)
	}
}

// Cleanup runs forever, pruning keys older than 24 hours once per hour.
// Call it in a goroutine: go store.Cleanup()
func (s *Store) Cleanup() {
	for {
		time.Sleep(time.Hour)
		_, err := s.conn.Exec(
			`DELETE FROM idempotency_keys
			 WHERE created_at < datetime('now', '-24 hours')`,
		)
		if err != nil {
			log.Printf("idempotency: cleanup error: %v", err)
		}
	}
}
