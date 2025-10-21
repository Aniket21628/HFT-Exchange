package engine

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/hft-exchange/backend/internal/domain"
)

type Exchange struct {
	engines      map[string]*MatchingEngine
	mu           sync.RWMutex
	tradeStore   TradeStore
	orderStore   OrderStore
	balanceStore BalanceStore
	ctx          context.Context
	cancel       context.CancelFunc
	onTrade      func(*domain.Trade)  // Callback when trade executes
}

type TradeStore interface {
	SaveTrade(trade *domain.Trade) error
}

type OrderStore interface {
	SaveOrder(order *domain.Order) error
	UpdateOrder(order *domain.Order) error
	GetOrderByID(orderID string) (*domain.Order, error)
}

type BalanceStore interface {
	GetBalance(userID, asset string) (available, locked float64, err error)
	UpdateBalance(userID, asset string, available, locked float64) error
}

func NewExchange(tradeStore TradeStore, orderStore OrderStore, balanceStore BalanceStore) *Exchange {
	ctx, cancel := context.WithCancel(context.Background())
	ex := &Exchange{
		engines:      make(map[string]*MatchingEngine),
		tradeStore:   tradeStore,
		orderStore:   orderStore,
		balanceStore: balanceStore,
		ctx:          ctx,
		cancel:       cancel,
	}
	return ex
}

func (ex *Exchange) Start() {
	symbols := []string{"BTC-USD", "ETH-USD", "SOL-USD", "USDC-USD"}
	
	for _, symbol := range symbols {
		ex.AddSymbol(symbol)
	}

	go ex.processAllTrades()
	go ex.processAllOrderUpdates()
}

func (ex *Exchange) AddSymbol(symbol string) {
	ex.mu.Lock()
	defer ex.mu.Unlock()

	if _, exists := ex.engines[symbol]; !exists {
		engine := NewMatchingEngine(symbol)
		ex.engines[symbol] = engine
		log.Printf("Added trading pair: %s", symbol)
	}
}

func (ex *Exchange) SubmitOrder(order *domain.Order) error {
	ex.mu.RLock()
	engine, exists := ex.engines[order.Symbol]
	ex.mu.RUnlock()

	if !exists {
		return nil
	}

	if err := ex.orderStore.SaveOrder(order); err != nil {
		return err
	}

	go engine.ProcessOrder(order)
	return nil
}

func (ex *Exchange) CancelOrder(orderID, symbol string) bool {
	ex.mu.RLock()
	engine, exists := ex.engines[symbol]
	ex.mu.RUnlock()

	if !exists {
		return false
	}

	return engine.CancelOrder(orderID)
}

func (ex *Exchange) GetOrderBook(symbol string, depth int) *domain.OrderBook {
	ex.mu.RLock()
	engine, exists := ex.engines[symbol]
	ex.mu.RUnlock()

	if !exists {
		return &domain.OrderBook{
			Symbol:    symbol,
			Bids:      []domain.OrderBookLevel{},
			Asks:      []domain.OrderBookLevel{},
			Timestamp: time.Now(),
		}
	}

	return engine.GetOrderBook(depth)
}

func (ex *Exchange) processAllTrades() {
	for {
		select {
		case <-ex.ctx.Done():
			return
		default:
			ex.mu.RLock()
			for _, engine := range ex.engines {
				select {
				case trade := <-engine.TradeChan():
					if err := ex.tradeStore.SaveTrade(trade); err != nil {
						log.Printf("Failed to save trade: %v", err)
					}
					// Settle balances for the trade
					if err := ex.settleTrade(trade); err != nil {
						log.Printf("Failed to settle trade balances: %v", err)
					}
					// Broadcast trade via callback
					if ex.onTrade != nil {
						ex.onTrade(trade)
					}
				default:
				}
			}
			ex.mu.RUnlock()
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (ex *Exchange) processAllOrderUpdates() {
	for {
		select {
		case <-ex.ctx.Done():
			return
		default:
			ex.mu.RLock()
			for _, engine := range ex.engines {
				select {
				case order := <-engine.OrderUpdatesChan():
					if err := ex.orderStore.UpdateOrder(order); err != nil {
						log.Printf("Failed to update order: %v", err)
					}
				default:
				}
			}
			ex.mu.RUnlock()
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (ex *Exchange) UpdatePrice(symbol string, price float64) {
	ex.mu.RLock()
	engine, exists := ex.engines[symbol]
	ex.mu.RUnlock()

	if exists {
		engine.CheckStopOrders(price)
	}
}

func (ex *Exchange) Stop() {
	ex.cancel()
}

// SetOnTradeCallback sets the callback to be called when a trade executes
func (ex *Exchange) SetOnTradeCallback(callback func(*domain.Trade)) {
	ex.onTrade = callback
}

// settleTrade updates balances for buyer and seller after a trade
func (ex *Exchange) settleTrade(trade *domain.Trade) error {
	// Parse symbol to get base and quote assets (e.g., "BTC-USD" -> "BTC", "USD")
	baseAsset, quoteAsset := ex.parseSymbol(trade.Symbol)
	
	tradeValue := trade.Price * trade.Quantity
	log.Printf("💰 Settling trade: %s bought %.4f %s @ %.2f from %s (total: %.2f %s)", 
		trade.BuyerID, trade.Quantity, baseAsset, trade.Price, trade.SellerID, tradeValue, quoteAsset)
	
	// Update buyer balances: -quote asset (USD), +base asset (BTC)
	buyerQuoteAvail, buyerQuoteLocked, err := ex.balanceStore.GetBalance(trade.BuyerID, quoteAsset)
	if err != nil {
		return err
	}
	buyerBaseAvail, buyerBaseLocked, err := ex.balanceStore.GetBalance(trade.BuyerID, baseAsset)
	if err != nil {
		return err
	}
	
	// Buyer: reduce available quote (USD), increase available base (BTC)
	log.Printf("  Buyer %s before: %s=%.4f(avail) %.4f(locked), %s=%.4f(avail) %.4f(locked)", 
		trade.BuyerID, quoteAsset, buyerQuoteAvail, buyerQuoteLocked, baseAsset, buyerBaseAvail, buyerBaseLocked)
	
	newBuyerQuoteAvail := buyerQuoteAvail - tradeValue  // DEDUCT USD from available
	newBuyerQuoteLocked := buyerQuoteLocked              // Keep locked as-is for now
	if err := ex.balanceStore.UpdateBalance(trade.BuyerID, quoteAsset, newBuyerQuoteAvail, newBuyerQuoteLocked); err != nil {
		return err
	}
	
	newBuyerBaseAvail := buyerBaseAvail + trade.Quantity  // ADD BTC to available
	newBuyerBaseLocked := buyerBaseLocked
	if err := ex.balanceStore.UpdateBalance(trade.BuyerID, baseAsset, newBuyerBaseAvail, newBuyerBaseLocked); err != nil {
		return err
	}
	
	log.Printf("  Buyer %s after: %s=%.4f(avail) %.4f(locked), %s=%.4f(avail) %.4f(locked)", 
		trade.BuyerID, quoteAsset, newBuyerQuoteAvail, newBuyerQuoteLocked, baseAsset, newBuyerBaseAvail, newBuyerBaseLocked)
	
	// Update seller balances: +quote asset (USD), -base asset (BTC)
	sellerQuoteAvail, sellerQuoteLocked, err := ex.balanceStore.GetBalance(trade.SellerID, quoteAsset)
	if err != nil {
		return err
	}
	sellerBaseAvail, sellerBaseLocked, err := ex.balanceStore.GetBalance(trade.SellerID, baseAsset)
	if err != nil {
		return err
	}
	
	// Seller: increase available quote (USD), reduce available base (BTC)
	log.Printf("  Seller %s before: %s=%.4f(avail) %.4f(locked), %s=%.4f(avail) %.4f(locked)", 
		trade.SellerID, quoteAsset, sellerQuoteAvail, sellerQuoteLocked, baseAsset, sellerBaseAvail, sellerBaseLocked)
	
	newSellerQuoteAvail := sellerQuoteAvail + tradeValue  // ADD USD to available
	newSellerQuoteLocked := sellerQuoteLocked
	if err := ex.balanceStore.UpdateBalance(trade.SellerID, quoteAsset, newSellerQuoteAvail, newSellerQuoteLocked); err != nil {
		return err
	}
	
	newSellerBaseAvail := sellerBaseAvail - trade.Quantity  // DEDUCT BTC from available
	newSellerBaseLocked := sellerBaseLocked
	if err := ex.balanceStore.UpdateBalance(trade.SellerID, baseAsset, newSellerBaseAvail, newSellerBaseLocked); err != nil {
		return err
	}
	
	log.Printf("  Seller %s after: %s=%.4f(avail) %.4f(locked), %s=%.4f(avail) %.4f(locked)", 
		trade.SellerID, quoteAsset, newSellerQuoteAvail, newSellerQuoteLocked, baseAsset, newSellerBaseAvail, newSellerBaseLocked)
	
	return nil
}

// parseSymbol splits a symbol like "BTC-USD" into base and quote assets
func (ex *Exchange) parseSymbol(symbol string) (base, quote string) {
	// Simple split on "-"
	parts := []rune(symbol)
	for i, r := range parts {
		if r == '-' {
			return string(parts[:i]), string(parts[i+1:])
		}
	}
	return symbol, "USD" // fallback
}

func (ex *Exchange) GetAllSymbols() []string {
	ex.mu.RLock()
	defer ex.mu.RUnlock()

	symbols := make([]string, 0, len(ex.engines))
	for symbol := range ex.engines {
		symbols = append(symbols, symbol)
	}
	return symbols
}
