package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
	driver string
}

func NewDB(connStr string) (*DB, error) {
	// Remove sqlite:// prefix if present
	dbPath := strings.TrimPrefix(connStr, "sqlite://")
	
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// SQLite works best with 1 connection
	db.SetMaxOpenConns(1)
	log.Printf("SQLite connection established: %s", dbPath)

	return &DB{db, "sqlite"}, nil
}

func (db *DB) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE NOT NULL,
		created_at TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS orders (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		symbol TEXT NOT NULL,
		side TEXT NOT NULL,
		type TEXT NOT NULL,
		quantity REAL NOT NULL,
		price REAL NOT NULL,
		stop_price REAL,
		filled_quantity REAL NOT NULL DEFAULT 0,
		remaining_qty REAL NOT NULL,
		status TEXT NOT NULL,
		time_in_force TEXT DEFAULT 'GTC',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
	CREATE INDEX IF NOT EXISTS idx_orders_symbol ON orders(symbol);
	CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
	CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at DESC);

	CREATE TABLE IF NOT EXISTS trades (
		id TEXT PRIMARY KEY,
		symbol TEXT NOT NULL,
		buy_order_id TEXT NOT NULL,
		sell_order_id TEXT NOT NULL,
		buyer_id TEXT NOT NULL,
		seller_id TEXT NOT NULL,
		price REAL NOT NULL,
		quantity REAL NOT NULL,
		maker_order_id TEXT NOT NULL,
		taker_order_id TEXT NOT NULL,
		executed_at TEXT NOT NULL,
		FOREIGN KEY (buy_order_id) REFERENCES orders(id),
		FOREIGN KEY (sell_order_id) REFERENCES orders(id),
		FOREIGN KEY (buyer_id) REFERENCES users(id),
		FOREIGN KEY (seller_id) REFERENCES users(id)
	);

	CREATE INDEX IF NOT EXISTS idx_trades_symbol ON trades(symbol);
	CREATE INDEX IF NOT EXISTS idx_trades_buyer_id ON trades(buyer_id);
	CREATE INDEX IF NOT EXISTS idx_trades_seller_id ON trades(seller_id);
	CREATE INDEX IF NOT EXISTS idx_trades_executed_at ON trades(executed_at DESC);

	CREATE TABLE IF NOT EXISTS balances (
		user_id TEXT NOT NULL,
		asset TEXT NOT NULL,
		available REAL NOT NULL DEFAULT 0,
		locked REAL NOT NULL DEFAULT 0,
		updated_at TEXT NOT NULL DEFAULT (datetime('now')),
		PRIMARY KEY (user_id, asset),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE INDEX IF NOT EXISTS idx_balances_user_id ON balances(user_id);

	CREATE TABLE IF NOT EXISTS positions (
		user_id TEXT NOT NULL,
		symbol TEXT NOT NULL,
		quantity REAL NOT NULL DEFAULT 0,
		avg_entry_price REAL NOT NULL DEFAULT 0,
		realized_pnl REAL NOT NULL DEFAULT 0,
		updated_at TEXT NOT NULL DEFAULT (datetime('now')),
		PRIMARY KEY (user_id, symbol),
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS tickers (
		symbol TEXT PRIMARY KEY,
		price REAL NOT NULL,
		high_24h REAL NOT NULL DEFAULT 0,
		low_24h REAL NOT NULL DEFAULT 0,
		volume_24h REAL NOT NULL DEFAULT 0,
		change_24h REAL NOT NULL DEFAULT 0,
		updated_at TEXT NOT NULL DEFAULT (datetime('now'))
	);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	log.Println("Database schema initialized")
	return nil
}

func (db *DB) SeedData() error {
	// Create demo users
	demoUsers := []struct {
		id       string
		username string
		email    string
	}{
		{"user-1", "trader1", "trader1@hft.com"},
		{"user-2", "trader2", "trader2@hft.com"},
		{"user-3", "marketmaker", "mm@hft.com"},
	}

	for _, user := range demoUsers {
		_, err := db.Exec(`
			INSERT INTO users (id, username, email, created_at)
			VALUES (?, ?, ?, datetime('now'))
			ON CONFLICT (id) DO NOTHING
		`, user.id, user.username, user.email)
		if err != nil {
			return fmt.Errorf("failed to seed user %s: %w", user.username, err)
		}

		// Give each user initial balances
		assets := []struct {
			asset  string
			amount float64
		}{
			{"USD", 100000.0},
			{"BTC", 1.0},
			{"ETH", 10.0},
			{"SOL", 100.0},
			{"USDC", 50000.0},
		}

		for _, asset := range assets {
			_, err := db.Exec(`
				INSERT INTO balances (user_id, asset, available, locked, updated_at)
				VALUES (?, ?, ?, 0, datetime('now'))
				ON CONFLICT (user_id, asset) DO NOTHING
			`, user.id, asset.asset, asset.amount)
			if err != nil {
				return fmt.Errorf("failed to seed balance for %s: %w", user.username, err)
			}
		}
	}

	// Initialize tickers
	tickers := []struct {
		symbol string
		price  float64
	}{
		{"BTC-USD", 45000.0},
		{"ETH-USD", 2500.0},
		{"SOL-USD", 100.0},
		{"USDC-USD", 1.0},
	}

	for _, ticker := range tickers {
		_, err := db.Exec(`
			INSERT INTO tickers (symbol, price, high_24h, low_24h, volume_24h, change_24h, updated_at)
			VALUES (?, ?, ?, ?, 0, 0, datetime('now'))
			ON CONFLICT (symbol) DO UPDATE SET price = ?, updated_at = datetime('now')
		`, ticker.symbol, ticker.price, ticker.price, ticker.price, ticker.price)
		if err != nil {
			return fmt.Errorf("failed to seed ticker %s: %w", ticker.symbol, err)
		}
	}

	log.Println("Database seeded with demo data")
	return nil
}

// Helper to convert time.Time to SQLite format
func (db *DB) TimeToString(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
