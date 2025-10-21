package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderSide string
type OrderType string
type OrderStatus string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

const (
	OrderTypeLimit     OrderType = "LIMIT"
	OrderTypeMarket    OrderType = "MARKET"
	OrderTypeStopLimit OrderType = "STOP_LIMIT"
)

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusPartial   OrderStatus = "PARTIAL"
	OrderStatusFilled    OrderStatus = "FILLED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusRejected  OrderStatus = "REJECTED"
)

type Order struct {
	ID              string      `json:"id"`
	UserID          string      `json:"user_id"`
	Symbol          string      `json:"symbol"`
	Side            OrderSide   `json:"side"`
	Type            OrderType   `json:"type"`
	Quantity        float64     `json:"quantity"`
	Price           float64     `json:"price"`
	StopPrice       float64     `json:"stop_price,omitempty"`
	FilledQuantity  float64     `json:"filled_quantity"`
	RemainingQty    float64     `json:"remaining_qty"`
	Status          OrderStatus `json:"status"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	TimeInForce     string      `json:"time_in_force"` // GTC, IOC, FOK
}

type Trade struct {
	ID           string    `json:"id"`
	Symbol       string    `json:"symbol"`
	BuyOrderID   string    `json:"buy_order_id"`
	SellOrderID  string    `json:"sell_order_id"`
	BuyerID      string    `json:"buyer_id"`
	SellerID     string    `json:"seller_id"`
	Price        float64   `json:"price"`
	Quantity     float64   `json:"quantity"`
	ExecutedAt   time.Time `json:"executed_at"`
	MakerOrderID string    `json:"maker_order_id"`
	TakerOrderID string    `json:"taker_order_id"`
}

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Portfolio struct {
	UserID    string             `json:"user_id"`
	Balances  map[string]float64 `json:"balances"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type Position struct {
	UserID         string  `json:"user_id"`
	Symbol         string  `json:"symbol"`
	Quantity       float64 `json:"quantity"`
	AvgEntryPrice  float64 `json:"avg_entry_price"`
	CurrentPrice   float64 `json:"current_price"`
	UnrealizedPnL  float64 `json:"unrealized_pnl"`
	RealizedPnL    float64 `json:"realized_pnl"`
}

type Ticker struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	High24h   float64   `json:"high_24h"`
	Low24h    float64   `json:"low_24h"`
	Volume24h float64   `json:"volume_24h"`
	Change24h float64   `json:"change_24h"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrderBook struct {
	Symbol    string           `json:"symbol"`
	Bids      []OrderBookLevel `json:"bids"`
	Asks      []OrderBookLevel `json:"asks"`
	Timestamp time.Time        `json:"timestamp"`
}

type OrderBookLevel struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
	Orders   int     `json:"orders"`
}

func NewOrder(userID, symbol string, side OrderSide, orderType OrderType, quantity, price float64) *Order {
	now := time.Now()
	return &Order{
		ID:             uuid.New().String(),
		UserID:         userID,
		Symbol:         symbol,
		Side:           side,
		Type:           orderType,
		Quantity:       quantity,
		Price:          price,
		FilledQuantity: 0,
		RemainingQty:   quantity,
		Status:         OrderStatusPending,
		CreatedAt:      now,
		UpdatedAt:      now,
		TimeInForce:    "GTC",
	}
}

func NewTrade(symbol, buyOrderID, sellOrderID, buyerID, sellerID string, price, quantity float64, makerOrderID, takerOrderID string) *Trade {
	return &Trade{
		ID:           uuid.New().String(),
		Symbol:       symbol,
		BuyOrderID:   buyOrderID,
		SellOrderID:  sellOrderID,
		BuyerID:      buyerID,
		SellerID:     sellerID,
		Price:        price,
		Quantity:     quantity,
		ExecutedAt:   time.Now(),
		MakerOrderID: makerOrderID,
		TakerOrderID: takerOrderID,
	}
}
