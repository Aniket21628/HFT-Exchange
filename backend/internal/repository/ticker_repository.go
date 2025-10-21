package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/hft-exchange/backend/internal/domain"
)

type TickerRepository struct {
	db *sql.DB
}

func NewTickerRepository(db *sql.DB) *TickerRepository {
	return &TickerRepository{db: db}
}

func (r *TickerRepository) GetTicker(symbol string) (*domain.Ticker, error) {
	query := `
		SELECT symbol, price, high_24h, low_24h, volume_24h, change_24h, updated_at
		FROM tickers
		WHERE symbol = $1
	`
	
	ticker := &domain.Ticker{}
	var updatedAt sql.NullString
	err := r.db.QueryRow(query, symbol).Scan(
		&ticker.Symbol, &ticker.Price, &ticker.High24h, &ticker.Low24h,
		&ticker.Volume24h, &ticker.Change24h, &updatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get ticker: %w", err)
	}
	
	// Parse timestamp if valid
	if updatedAt.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", updatedAt.String); err == nil {
			ticker.UpdatedAt = t
		} else if t, err := time.Parse(time.RFC3339, updatedAt.String); err == nil {
			ticker.UpdatedAt = t
		}
	}
	
	return ticker, nil
}

func (r *TickerRepository) GetAllTickers() ([]*domain.Ticker, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	query := `
		SELECT symbol, price, high_24h, low_24h, volume_24h, change_24h, updated_at
		FROM tickers
	`
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickers: %w", err)
	}
	defer rows.Close()
	
	tickers := make([]*domain.Ticker, 0)
	for rows.Next() {
		ticker := &domain.Ticker{}
		var updatedAt sql.NullString
		err := rows.Scan(
			&ticker.Symbol, &ticker.Price, &ticker.High24h, &ticker.Low24h,
			&ticker.Volume24h, &ticker.Change24h, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ticker: %w", err)
		}
		
		// Parse timestamp if valid
		if updatedAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", updatedAt.String); err == nil {
				ticker.UpdatedAt = t
			} else if t, err := time.Parse(time.RFC3339, updatedAt.String); err == nil {
				ticker.UpdatedAt = t
			}
		}
		
		tickers = append(tickers, ticker)
	}
	
	return tickers, nil
}

func (r *TickerRepository) UpdateTicker(ticker *domain.Ticker) error {
	query := `
		UPDATE tickers
		SET price = $1, high_24h = $2, low_24h = $3, volume_24h = $4, 
		    change_24h = $5, updated_at = $6
		WHERE symbol = $7
	`
	
	_, err := r.db.Exec(query, ticker.Price, ticker.High24h, ticker.Low24h,
		ticker.Volume24h, ticker.Change24h, ticker.UpdatedAt, ticker.Symbol)
	
	if err != nil {
		return fmt.Errorf("failed to update ticker: %w", err)
	}
	
	return nil
}
