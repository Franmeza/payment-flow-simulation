package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/franmeza/payment-flow-simulation/internal/models"
	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

// Creates/opens the database
func New() *DB {
	conn, err := sql.Open("sqlite", "./payments.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	d := &DB{conn: conn}
	d.migrate()
	d.seed()
	return d
}

func (d *DB) migrate() {
	_, err := d.conn.Exec(`
		CREATE TABLE IF NOT EXISTS cards (
			uid          TEXT PRIMARY KEY,
			card_holder  TEXT NOT NULL,
			balance      REAL NOT NULL,
			status       TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS transactions (
			id           TEXT PRIMARY KEY,
			card_uid     TEXT NOT NULL,
			merchant_id  TEXT NOT NULL,
			amount       REAL NOT NULL,
			approved     INTEGER NOT NULL,
			reason       TEXT,
			timestamp    DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}

func (d *DB) seed() {
	// Only seed if table is empty
	var count int
	d.conn.QueryRow(`SELECT COUNT(*) FROM cards`).Scan(&count)
	if count > 0 {
		return
	}

	cards := []models.Card{
		{UID: "2E:F8:14:07", CardHolder: "Hely Cimer", Balance: 500.00, Status: "active"},
		{UID: "1B:5E:32:07", CardHolder: "Jane Smith", Balance: 23.50,  Status: "active"},
		{UID: "BLOCKED:01", CardHolder: "Bob Block", Balance: 100.00, Status: "blocked"},
	}

	for _, c := range cards {
		_, err := d.conn.Exec(
			`INSERT INTO cards (uid, card_holder, balance, status) VALUES (?, ?, ?, ?)`,
			c.UID, c.CardHolder, c.Balance, c.Status,
		)
		if err != nil {
			log.Printf("seed error for card %s: %v", c.UID, err)
		}
	}
	log.Println("Database seeded with test cards")
}

func (d *DB) GetCard(uid string) (*models.Card, error) {
	card := &models.Card{}
	err := d.conn.QueryRow(
		`SELECT uid, card_holder, balance, status FROM cards WHERE uid = ?`, uid,
	).Scan(&card.UID, &card.CardHolder, &card.Balance, &card.Status)
	if err != nil {
		return nil, err
	}
	return card, nil
}

func (d *DB) DeductBalance(uid string, amount float64) error {
	_, err := d.conn.Exec(
		`UPDATE cards SET balance = balance - ? WHERE uid = ?`, amount, uid,
	)
	return err
}

func (d *DB) SaveTransaction(t models.Transaction) error {
	approved := 0
	if t.Approved {
		approved = 1
	}
	_, err := d.conn.Exec(
		`INSERT INTO transactions (id, card_uid, merchant_id, amount, approved, reason)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		t.ID, t.CardUID, t.MerchantID, t.Amount, approved, t.Reason,
	)
	return err
}

func (d *DB) ListTransactions(limit int) ([]models.Transaction, error) {
	rows, err := d.conn.Query(
		`SELECT id, card_uid, merchant_id, amount, approved, COALESCE(reason, ''), timestamp
		 FROM transactions
		 ORDER BY datetime(timestamp) DESC
		 LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := make([]models.Transaction, 0, limit)
	for rows.Next() {
		var tx models.Transaction
		var approved int
		var timestampRaw string
		if err := rows.Scan(
			&tx.ID,
			&tx.CardUID,
			&tx.MerchantID,
			&tx.Amount,
			&approved,
			&tx.Reason,
			&timestampRaw,
		); err != nil {
			return nil, err
		}
		tx.Approved = approved == 1
		parsedTimestamp, err := parseSQLiteTimestamp(timestampRaw)
		if err != nil {
			return nil, err
		}
		tx.Timestamp = parsedTimestamp
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (d *DB) GetTransactionStats() (models.TransactionStats, error) {
	var stats models.TransactionStats
	err := d.conn.QueryRow(`
		SELECT
			COUNT(*) AS total_transactions,
			COALESCE(SUM(CASE WHEN approved = 1 THEN 1 ELSE 0 END), 0) AS approved_count,
			COALESCE(SUM(CASE WHEN approved = 0 THEN 1 ELSE 0 END), 0) AS declined_count,
			COALESCE(SUM(amount), 0) AS total_volume,
			COALESCE(SUM(CASE WHEN approved = 1 THEN amount ELSE 0 END), 0) AS approved_volume,
			COALESCE(SUM(CASE WHEN approved = 0 THEN amount ELSE 0 END), 0) AS declined_volume
		FROM transactions
	`).Scan(
		&stats.TotalTransactions,
		&stats.ApprovedCount,
		&stats.DeclinedCount,
		&stats.TotalVolume,
		&stats.ApprovedVolume,
		&stats.DeclinedVolume,
	)
	if err != nil {
		return stats, err
	}

	if stats.TotalTransactions > 0 {
		stats.ApprovalRate = float64(stats.ApprovedCount) / float64(stats.TotalTransactions)
	}

	return stats, nil
}

func parseSQLiteTimestamp(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, nil
	}

	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported sqlite timestamp format: %s", raw)
}