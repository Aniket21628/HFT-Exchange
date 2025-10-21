package engine

import (
	"container/heap"
	"log"
	"sync"
	"time"

	"github.com/hft-exchange/backend/internal/domain"
)

type MatchingEngine struct {
	symbol       string
	buyOrders    *OrderHeap
	sellOrders   *OrderHeap
	mu           sync.RWMutex
	tradeChan    chan *domain.Trade
	orderUpdates chan *domain.Order
	stopLimitOrders []*domain.Order
}

func NewMatchingEngine(symbol string) *MatchingEngine {
	me := &MatchingEngine{
		symbol:       symbol,
		buyOrders:    &OrderHeap{isBuy: true},
		sellOrders:   &OrderHeap{isBuy: false},
		tradeChan:    make(chan *domain.Trade, 1000),
		orderUpdates: make(chan *domain.Order, 1000),
		stopLimitOrders: make([]*domain.Order, 0),
	}
	heap.Init(me.buyOrders)
	heap.Init(me.sellOrders)
	return me
}

func (me *MatchingEngine) ProcessOrder(order *domain.Order) {
	me.mu.Lock()
	defer me.mu.Unlock()

	if order.Type == domain.OrderTypeStopLimit {
		log.Printf("🛑 Stop-Limit order placed: %s %s %.4f @ Stop:$%.2f Limit:$%.2f", 
			order.Side, order.Symbol, order.Quantity, order.StopPrice, order.Price)
		me.stopLimitOrders = append(me.stopLimitOrders, order)
		return
	}

	if order.Type == domain.OrderTypeMarket {
		log.Printf("⚡ Market order: %s %s %.4f", order.Side, order.Symbol, order.Quantity)
		me.matchMarketOrder(order)
	} else {
		log.Printf("🎯 Limit order: %s %s %.4f @ $%.2f", order.Side, order.Symbol, order.Quantity, order.Price)
		me.matchLimitOrder(order)
	}
}

func (me *MatchingEngine) matchLimitOrder(order *domain.Order) {
	var oppositeBook *OrderHeap
	if order.Side == domain.OrderSideBuy {
		oppositeBook = me.sellOrders
	} else {
		oppositeBook = me.buyOrders
	}

	for oppositeBook.Len() > 0 && order.RemainingQty > 0 {
		topOrder := oppositeBook.orders[0]

		canMatch := false
		if order.Side == domain.OrderSideBuy {
			canMatch = order.Price >= topOrder.Price
		} else {
			canMatch = order.Price <= topOrder.Price
		}

		if !canMatch {
			break
		}

		matchQty := min(order.RemainingQty, topOrder.RemainingQty)
		tradePrice := topOrder.Price

		me.executeTrade(order, topOrder, matchQty, tradePrice)

		if topOrder.RemainingQty == 0 {
			heap.Pop(oppositeBook)
		} else {
			heap.Fix(oppositeBook, 0)
		}
	}

	if order.RemainingQty > 0 && order.TimeInForce == "GTC" {
		if order.Side == domain.OrderSideBuy {
			heap.Push(me.buyOrders, order)
		} else {
			heap.Push(me.sellOrders, order)
		}
		me.orderUpdates <- order
	} else if order.RemainingQty > 0 {
		order.Status = domain.OrderStatusCancelled
		me.orderUpdates <- order
	}
}

func (me *MatchingEngine) matchMarketOrder(order *domain.Order) {
	var oppositeBook *OrderHeap
	if order.Side == domain.OrderSideBuy {
		oppositeBook = me.sellOrders
	} else {
		oppositeBook = me.buyOrders
	}

	for oppositeBook.Len() > 0 && order.RemainingQty > 0 {
		topOrder := oppositeBook.orders[0]
		matchQty := min(order.RemainingQty, topOrder.RemainingQty)
		tradePrice := topOrder.Price

		me.executeTrade(order, topOrder, matchQty, tradePrice)

		if topOrder.RemainingQty == 0 {
			heap.Pop(oppositeBook)
		} else {
			heap.Fix(oppositeBook, 0)
		}
	}

	if order.RemainingQty > 0 {
		order.Status = domain.OrderStatusPartial
	}
	me.orderUpdates <- order
}

func (me *MatchingEngine) executeTrade(order1, order2 *domain.Order, quantity, price float64) {
	order1.FilledQuantity += quantity
	order1.RemainingQty -= quantity
	order2.FilledQuantity += quantity
	order2.RemainingQty -= quantity

	if order1.RemainingQty == 0 {
		order1.Status = domain.OrderStatusFilled
	} else {
		order1.Status = domain.OrderStatusPartial
	}

	if order2.RemainingQty == 0 {
		order2.Status = domain.OrderStatusFilled
	} else {
		order2.Status = domain.OrderStatusPartial
	}

	order1.UpdatedAt = time.Now()
	order2.UpdatedAt = time.Now()

	var buyOrderID, sellOrderID, buyerID, sellerID string
	if order1.Side == domain.OrderSideBuy {
		buyOrderID = order1.ID
		sellOrderID = order2.ID
		buyerID = order1.UserID
		sellerID = order2.UserID
	} else {
		buyOrderID = order2.ID
		sellOrderID = order1.ID
		buyerID = order2.UserID
		sellerID = order1.UserID
	}

	makerOrderID := order2.ID
	takerOrderID := order1.ID

	trade := domain.NewTrade(me.symbol, buyOrderID, sellOrderID, buyerID, sellerID, price, quantity, makerOrderID, takerOrderID)
	me.tradeChan <- trade
	me.orderUpdates <- order1
	me.orderUpdates <- order2
}

func (me *MatchingEngine) CancelOrder(orderID string) bool {
	me.mu.Lock()
	defer me.mu.Unlock()

	if me.cancelFromHeap(me.buyOrders, orderID) {
		return true
	}
	if me.cancelFromHeap(me.sellOrders, orderID) {
		return true
	}
	return false
}

func (me *MatchingEngine) cancelFromHeap(h *OrderHeap, orderID string) bool {
	for i, order := range h.orders {
		if order.ID == orderID {
			heap.Remove(h, i)
			order.Status = domain.OrderStatusCancelled
			order.UpdatedAt = time.Now()
			me.orderUpdates <- order
			return true
		}
	}
	return false
}

func (me *MatchingEngine) GetOrderBook(depth int) *domain.OrderBook {
	me.mu.RLock()
	defer me.mu.RUnlock()

	bids := make([]domain.OrderBookLevel, 0)
	asks := make([]domain.OrderBookLevel, 0)

	bidMap := make(map[float64]*domain.OrderBookLevel)
	for _, order := range me.buyOrders.orders {
		if level, exists := bidMap[order.Price]; exists {
			level.Quantity += order.RemainingQty
			level.Orders++
		} else {
			bidMap[order.Price] = &domain.OrderBookLevel{
				Price:    order.Price,
				Quantity: order.RemainingQty,
				Orders:   1,
			}
		}
	}

	askMap := make(map[float64]*domain.OrderBookLevel)
	for _, order := range me.sellOrders.orders {
		if level, exists := askMap[order.Price]; exists {
			level.Quantity += order.RemainingQty
			level.Orders++
		} else {
			askMap[order.Price] = &domain.OrderBookLevel{
				Price:    order.Price,
				Quantity: order.RemainingQty,
				Orders:   1,
			}
		}
	}

	for _, level := range bidMap {
		bids = append(bids, *level)
		if len(bids) >= depth {
			break
		}
	}

	for _, level := range askMap {
		asks = append(asks, *level)
		if len(asks) >= depth {
			break
		}
	}

	return &domain.OrderBook{
		Symbol:    me.symbol,
		Bids:      bids,
		Asks:      asks,
		Timestamp: time.Now(),
	}
}

func (me *MatchingEngine) CheckStopOrders(currentPrice float64) {
	me.mu.Lock()
	defer me.mu.Unlock()

	triggered := make([]*domain.Order, 0)
	remaining := make([]*domain.Order, 0)

	for _, order := range me.stopLimitOrders {
		shouldTrigger := false
		if order.Side == domain.OrderSideBuy && currentPrice >= order.StopPrice {
			shouldTrigger = true
		} else if order.Side == domain.OrderSideSell && currentPrice <= order.StopPrice {
			shouldTrigger = true
		}

		if shouldTrigger {
			log.Printf("🔔 Stop-Limit TRIGGERED: %s %s %.4f @ Stop:$%.2f → Now Limit:$%.2f (Current:$%.2f)", 
				order.Side, order.Symbol, order.Quantity, order.StopPrice, order.Price, currentPrice)
			order.Type = domain.OrderTypeLimit
			triggered = append(triggered, order)
		} else {
			remaining = append(remaining, order)
		}
	}

	me.stopLimitOrders = remaining

	me.mu.Unlock()
	for _, order := range triggered {
		me.ProcessOrder(order)
	}
	me.mu.Lock()
}

func (me *MatchingEngine) TradeChan() <-chan *domain.Trade {
	return me.tradeChan
}

func (me *MatchingEngine) OrderUpdatesChan() <-chan *domain.Order {
	return me.orderUpdates
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
