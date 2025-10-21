package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type BalanceRepository struct {
	db *sql.DB
}

type Balance struct {
	UserID    string
	Asset     string
	Available float64
	Locked    float64
	UpdatedAt time.Time
}

func NewBalanceRepository(db *sql.DB) *BalanceRepository {
	return &BalanceRepository{db: db}
}

func (r *BalanceRepository) GetBalance(userID, asset string) (*Balance, error) {
	query := `
		SELECT user_id, asset, available, locked, updated_at
		FROM balances
		WHERE user_id = $1 AND asset = $2
	`
	
	balance := &Balance{}
	var updatedAt sql.NullString
	err := r.db.QueryRow(query, userID, asset).Scan(
		&balance.UserID, &balance.Asset, &balance.Available, 
		&balance.Locked, &updatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return &Balance{
				UserID:    userID,
				Asset:     asset,
				Available: 0,
				Locked:    0,
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}
	
	// Parse timestamp if valid
	if updatedAt.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", updatedAt.String); err == nil {
			balance.UpdatedAt = t
		} else if t, err := time.Parse(time.RFC3339, updatedAt.String); err == nil {
			balance.UpdatedAt = t
		}
	}
	
	return balance, nil
}

func (r *BalanceRepository) GetAllBalances(userID string) ([]*Balance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	query := `
		SELECT user_id, asset, available, locked, updated_at
		FROM balances
		WHERE user_id = $1
	`
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get balances: %w", err)
	}
	defer rows.Close()
	
	balances := make([]*Balance, 0)
	for rows.Next() {
		balance := &Balance{}
		var updatedAt sql.NullString
		err := rows.Scan(
			&balance.UserID, &balance.Asset, &balance.Available,
			&balance.Locked, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan balance: %w", err)
		}
		
		// Parse timestamp if valid
		if updatedAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", updatedAt.String); err == nil {
				balance.UpdatedAt = t
			} else if t, err := time.Parse(time.RFC3339, updatedAt.String); err == nil {
				balance.UpdatedAt = t
			}
		}
		
		balances = append(balances, balance)
	}
	
	return balances, nil
}

func (r *BalanceRepository) UpdateBalance(userID, asset string, available, locked float64) error {
	now := time.Now()
	query := `
		INSERT INTO balances (user_id, asset, available, locked, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, asset) 
		DO UPDATE SET available = $3, locked = $4, updated_at = $5
	`
	
	_, err := r.db.Exec(query, userID, asset, available, locked, now)
	if err != nil {
		return fmt.Errorf("failed to update balance for %s/%s (%.4f/%.4f): %w", userID, asset, available, locked, err)
	}
	return nil
}

func (r *BalanceRepository) LockBalance(userID, asset string, amount float64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	var available, locked float64
	err = tx.QueryRow(`
		SELECT available, locked FROM balances 
		WHERE user_id = $1 AND asset = $2
		FOR UPDATE
	`, userID, asset).Scan(&available, &locked)
	
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}
	
	if available < amount {
		return fmt.Errorf("insufficient balance")
	}
	
	_, err = tx.Exec(`
		UPDATE balances 
		SET available = available - $1, locked = locked + $1, updated_at = $4
		WHERE user_id = $2 AND asset = $3
	`, amount, userID, asset, time.Now())
	
	if err != nil {
		return fmt.Errorf("failed to lock balance: %w", err)
	}
	
	return tx.Commit()
}

func (r *BalanceRepository) UnlockBalance(userID, asset string, amount float64) error {
	query := `
		UPDATE balances 
		SET available = available + $1, locked = locked - $1, updated_at = $4
		WHERE user_id = $2 AND asset = $3
	`
	
	_, err := r.db.Exec(query, amount, userID, asset, time.Now())
	if err != nil {
		return fmt.Errorf("failed to unlock balance: %w", err)
	}
	
	return nil
}
