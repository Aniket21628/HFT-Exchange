package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/hft-exchange/backend/internal/domain"
)

type TradeRepository struct {
	db *sql.DB
}

func NewTradeRepository(db *sql.DB) *TradeRepository {
	return &TradeRepository{db: db}
}

func (r *TradeRepository) SaveTrade(trade *domain.Trade) error {
	query := `
		INSERT INTO trades (id, symbol, buy_order_id, sell_order_id, buyer_id, seller_id, 
			price, quantity, maker_order_id, taker_order_id, executed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.db.Exec(query, trade.ID, trade.Symbol, trade.BuyOrderID, trade.SellOrderID,
		trade.BuyerID, trade.SellerID, trade.Price, trade.Quantity, 
		trade.MakerOrderID, trade.TakerOrderID, trade.ExecutedAt)
	
	if err != nil {
		return fmt.Errorf("failed to save trade: %w", err)
	}
	return nil
}

func (r *TradeRepository) GetRecentTrades(symbol string, limit int) ([]*domain.Trade, error) {
	query := `
		SELECT id, symbol, buy_order_id, sell_order_id, buyer_id, seller_id,
			price, quantity, maker_order_id, taker_order_id, executed_at
		FROM trades 
		WHERE symbol = $1
		ORDER BY executed_at DESC
		LIMIT $2
	`
	
	rows, err := r.db.Query(query, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent trades: %w", err)
	}
	defer rows.Close()
	
	trades := make([]*domain.Trade, 0)
	for rows.Next() {
		trade := &domain.Trade{}
		var executedAt sql.NullString
		err := rows.Scan(
			&trade.ID, &trade.Symbol, &trade.BuyOrderID, &trade.SellOrderID,
			&trade.BuyerID, &trade.SellerID, &trade.Price, &trade.Quantity,
			&trade.MakerOrderID, &trade.TakerOrderID, &executedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}
		
		// Parse timestamp
		if executedAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", executedAt.String); err == nil {
				trade.ExecutedAt = t
			} else if t, err := time.Parse(time.RFC3339, executedAt.String); err == nil {
				trade.ExecutedAt = t
			}
		}
		
		trades = append(trades, trade)
	}
	
	return trades, nil
}

func (r *TradeRepository) GetUserTrades(userID string, limit int) ([]*domain.Trade, error) {
	query := `
		SELECT id, symbol, buy_order_id, sell_order_id, buyer_id, seller_id,
			price, quantity, maker_order_id, taker_order_id, executed_at
		FROM trades 
		WHERE buyer_id = $1 OR seller_id = $1
		ORDER BY executed_at DESC
		LIMIT $2
	`
	
	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user trades: %w", err)
	}
	defer rows.Close()
	
	trades := make([]*domain.Trade, 0)
	for rows.Next() {
		trade := &domain.Trade{}
		var executedAt sql.NullString
		err := rows.Scan(
			&trade.ID, &trade.Symbol, &trade.BuyOrderID, &trade.SellOrderID,
			&trade.BuyerID, &trade.SellerID, &trade.Price, &trade.Quantity,
			&trade.MakerOrderID, &trade.TakerOrderID, &executedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}
		
		// Parse timestamp
		if executedAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", executedAt.String); err == nil {
				trade.ExecutedAt = t
			} else if t, err := time.Parse(time.RFC3339, executedAt.String); err == nil {
				trade.ExecutedAt = t
			}
		}
		
		trades = append(trades, trade)
	}
	
	return trades, nil
}
