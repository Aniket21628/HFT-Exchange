package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/hft-exchange/backend/internal/domain"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) SaveOrder(order *domain.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	query := `
		INSERT INTO orders (id, user_id, symbol, side, type, quantity, price, stop_price, 
			filled_quantity, remaining_qty, status, time_in_force, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.db.ExecContext(ctx, query, order.ID, order.UserID, order.Symbol, string(order.Side), string(order.Type),
		order.Quantity, order.Price, order.StopPrice, order.FilledQuantity, order.RemainingQty,
		string(order.Status), order.TimeInForce, order.CreatedAt, order.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}
	return nil
}

func (r *OrderRepository) UpdateOrder(order *domain.Order) error {
	query := `
		UPDATE orders 
		SET filled_quantity = $1, remaining_qty = $2, status = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(query, order.FilledQuantity, order.RemainingQty, order.Status, 
		order.UpdatedAt, order.ID)
	
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}
	return nil
}

func (r *OrderRepository) GetOrderByID(orderID string) (*domain.Order, error) {
	query := `
		SELECT id, user_id, symbol, side, type, quantity, price, stop_price,
			filled_quantity, remaining_qty, status, time_in_force, created_at, updated_at
		FROM orders WHERE id = $1
	`
	
	order := &domain.Order{}
	var stopPrice sql.NullFloat64
	var createdAt, updatedAt sql.NullString
	
	err := r.db.QueryRow(query, orderID).Scan(
		&order.ID, &order.UserID, &order.Symbol, &order.Side, &order.Type,
		&order.Quantity, &order.Price, &stopPrice, &order.FilledQuantity,
		&order.RemainingQty, &order.Status, &order.TimeInForce,
		&createdAt, &updatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	
	if stopPrice.Valid {
		order.StopPrice = stopPrice.Float64
	}
	
	// Parse timestamps
	if createdAt.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", createdAt.String); err == nil {
			order.CreatedAt = t
		} else if t, err := time.Parse(time.RFC3339, createdAt.String); err == nil {
			order.CreatedAt = t
		}
	}
	if updatedAt.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", updatedAt.String); err == nil {
			order.UpdatedAt = t
		} else if t, err := time.Parse(time.RFC3339, updatedAt.String); err == nil {
			order.UpdatedAt = t
		}
	}
	
	return order, nil
}

func (r *OrderRepository) GetOrdersByUser(userID string, limit int) ([]*domain.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	query := `
		SELECT id, user_id, symbol, side, type, quantity, price, stop_price,
			filled_quantity, remaining_qty, status, time_in_force, created_at, updated_at
		FROM orders WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}
	defer rows.Close()
	
	orders := make([]*domain.Order, 0)
	for rows.Next() {
		order := &domain.Order{}
		var stopPrice sql.NullFloat64
		var createdAt, updatedAt sql.NullString
		
		err := rows.Scan(
			&order.ID, &order.UserID, &order.Symbol, &order.Side, &order.Type,
			&order.Quantity, &order.Price, &stopPrice, &order.FilledQuantity,
			&order.RemainingQty, &order.Status, &order.TimeInForce,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		
		if stopPrice.Valid {
			order.StopPrice = stopPrice.Float64
		}
		
		// Parse timestamps
		if createdAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", createdAt.String); err == nil {
				order.CreatedAt = t
			} else if t, err := time.Parse(time.RFC3339, createdAt.String); err == nil {
				order.CreatedAt = t
			}
		}
		if updatedAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", updatedAt.String); err == nil {
				order.UpdatedAt = t
			} else if t, err := time.Parse(time.RFC3339, updatedAt.String); err == nil {
				order.UpdatedAt = t
			}
		}
		
		orders = append(orders, order)
	}
	
	return orders, nil
}

func (r *OrderRepository) GetOpenOrders(symbol string) ([]*domain.Order, error) {
	query := `
		SELECT id, user_id, symbol, side, type, quantity, price, stop_price,
			filled_quantity, remaining_qty, status, time_in_force, created_at, updated_at
		FROM orders 
		WHERE symbol = $1 AND status IN ('PENDING', 'PARTIAL')
		ORDER BY created_at ASC
	`
	
	rows, err := r.db.Query(query, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get open orders: %w", err)
	}
	defer rows.Close()
	
	orders := make([]*domain.Order, 0)
	for rows.Next() {
		order := &domain.Order{}
		var stopPrice sql.NullFloat64
		var createdAt, updatedAt sql.NullString
		
		err := rows.Scan(
			&order.ID, &order.UserID, &order.Symbol, &order.Side, &order.Type,
			&order.Quantity, &order.Price, &stopPrice, &order.FilledQuantity,
			&order.RemainingQty, &order.Status, &order.TimeInForce,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		
		if stopPrice.Valid {
			order.StopPrice = stopPrice.Float64
		}
		
		// Parse timestamps
		if createdAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", createdAt.String); err == nil {
				order.CreatedAt = t
			} else if t, err := time.Parse(time.RFC3339, createdAt.String); err == nil {
				order.CreatedAt = t
			}
		}
		if updatedAt.Valid {
			if t, err := time.Parse("2006-01-02 15:04:05", updatedAt.String); err == nil {
				order.UpdatedAt = t
			} else if t, err := time.Parse(time.RFC3339, updatedAt.String); err == nil {
				order.UpdatedAt = t
			}
		}
		
		orders = append(orders, order)
	}
	
	return orders, nil
}
