package db

import (
	"database/sql"
	"log"

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