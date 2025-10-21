package pricefeed

import (
	"context"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/hft-exchange/backend/internal/domain"
)

type PriceUpdateHandler func(symbol string, price float64)

type PriceSimulator struct {
	prices           map[string]float64
	mu               sync.RWMutex
	updateHandlers   []PriceUpdateHandler
	tickerRepo       TickerRepository
	ctx              context.Context
	cancel           context.CancelFunc
}

type TickerRepository interface {
	GetTicker(symbol string) (*domain.Ticker, error)
	UpdateTicker(ticker *domain.Ticker) error
}

func NewPriceSimulator(tickerRepo TickerRepository) *PriceSimulator {
	ctx, cancel := context.WithCancel(context.Background())
	return &PriceSimulator{
		prices:         make(map[string]float64),
		updateHandlers: make([]PriceUpdateHandler, 0),
		tickerRepo:     tickerRepo,
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (ps *PriceSimulator) Start() {
	symbols := []string{"BTC-USD", "ETH-USD", "SOL-USD", "USDC-USD"}
	
	// Initialize prices from database
	for _, symbol := range symbols {
		ticker, err := ps.tickerRepo.GetTicker(symbol)
		if err == nil {
			ps.mu.Lock()
			ps.prices[symbol] = ticker.Price
			ps.mu.Unlock()
		}
	}
	
	// Start price simulation for each symbol
	for _, symbol := range symbols {
		go ps.simulatePrice(symbol)
	}
	
	log.Println("Price simulator started")
}

func (ps *PriceSimulator) simulatePrice(symbol string) {
	ticker := time.NewTicker(3 * time.Second) // Slower updates for demo (was 100ms)
	defer ticker.Stop()
	
	// Different volatility for different assets
	volatility := ps.getVolatility(symbol)
	
	for {
		select {
		case <-ps.ctx.Done():
			return
		case <-ticker.C:
			ps.mu.Lock()
			currentPrice := ps.prices[symbol]
			
			// Geometric Brownian Motion for realistic price movement
			dt := 0.1 / 3600 // 100ms in hours
			drift := 0.0     // No drift for stable simulation
			
			randomShock := rand.NormFloat64()
			priceChange := currentPrice * (drift*dt + volatility*math.Sqrt(dt)*randomShock)
			newPrice := currentPrice + priceChange
			
			// Ensure price doesn't go negative or too extreme
			if newPrice < currentPrice*0.95 {
				newPrice = currentPrice * 0.95
			}
			if newPrice > currentPrice*1.05 {
				newPrice = currentPrice * 1.05
			}
			
			// Special case for stablecoins
			if symbol == "USDC-USD" {
				newPrice = 1.0 + (rand.Float64()-0.5)*0.001 // Very small fluctuation
			}
			
			ps.prices[symbol] = newPrice
			ps.mu.Unlock()
			
			// Update database FIRST (synchronously) before notifying handlers
			ps.updateTickerInDB(symbol, newPrice)
			
			// Notify handlers AFTER DB is updated
			for _, handler := range ps.updateHandlers {
				go handler(symbol, newPrice)
			}
		}
	}
}

func (ps *PriceSimulator) getVolatility(symbol string) float64 {
	switch symbol {
	case "BTC-USD":
		return 0.02
	case "ETH-USD":
		return 0.025
	case "SOL-USD":
		return 0.03
	case "USDC-USD":
		return 0.0001
	default:
		return 0.02
	}
}

func (ps *PriceSimulator) updateTickerInDB(symbol string, price float64) {
	ticker, err := ps.tickerRepo.GetTicker(symbol)
	if err != nil {
		log.Printf("Failed to get ticker %s: %v", symbol, err)
		return
	}
	
	// Store old price for change calculation
	oldPrice := ticker.Price
	ticker.Price = price
	ticker.UpdatedAt = time.Now()
	
	// Update 24h high/low
	if price > ticker.High24h || ticker.High24h == 0 {
		ticker.High24h = price
	}
	if price < ticker.Low24h || ticker.Low24h == 0 {
		ticker.Low24h = price
	}
	
	// Calculate 24h change percentage
	// For demo: use the midpoint of 24h range as baseline
	if ticker.High24h > 0 && ticker.Low24h > 0 {
		baseline := (ticker.High24h + ticker.Low24h) / 2
		if baseline > 0 {
			ticker.Change24h = ((price - baseline) / baseline) * 100
		}
	} else if oldPrice > 0 {
		// Fallback: use price change from last update
		ticker.Change24h = ((price - oldPrice) / oldPrice) * 100
	}
	
	if err := ps.tickerRepo.UpdateTicker(ticker); err != nil {
		log.Printf("Failed to update ticker %s: %v", symbol, err)
	}
}

func (ps *PriceSimulator) GetCurrentPrice(symbol string) float64 {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.prices[symbol]
}

func (ps *PriceSimulator) AddUpdateHandler(handler PriceUpdateHandler) {
	ps.updateHandlers = append(ps.updateHandlers, handler)
}

func (ps *PriceSimulator) Stop() {
	ps.cancel()
}
