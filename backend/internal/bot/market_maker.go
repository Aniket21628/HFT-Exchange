package bot

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/hft-exchange/backend/internal/domain"
)

type MarketMaker struct {
	userID         string
	exchange       ExchangeInterface
	priceSimulator PriceSimulator
	ctx            context.Context
	cancel         context.CancelFunc
}

type ExchangeInterface interface {
	SubmitOrder(order *domain.Order) error
	GetOrderBook(symbol string, depth int) *domain.OrderBook
}

type PriceSimulator interface {
	GetCurrentPrice(symbol string) float64
}

func NewMarketMaker(userID string, exchange ExchangeInterface, priceSimulator PriceSimulator) *MarketMaker {
	ctx, cancel := context.WithCancel(context.Background())
	return &MarketMaker{
		userID:         userID,
		exchange:       exchange,
		priceSimulator: priceSimulator,
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (mm *MarketMaker) Start() {
	symbols := []string{"BTC-USD", "ETH-USD", "SOL-USD"}
	
	for _, symbol := range symbols {
		go mm.makeMarket(symbol)
	}
	
	log.Printf("Market maker started for user: %s", mm.userID)
}

func (mm *MarketMaker) makeMarket(symbol string) {
	ticker := time.NewTicker(15 * time.Second) // Slower market making for demo (was 5s)
	defer ticker.Stop()
	
	for {
		select {
		case <-mm.ctx.Done():
			return
		case <-ticker.C:
			mm.placeOrders(symbol)
		}
	}
}

func (mm *MarketMaker) placeOrders(symbol string) {
	currentPrice := mm.priceSimulator.GetCurrentPrice(symbol)
	if currentPrice == 0 {
		return
	}
	
	// Place orders with spread around current price
	spread := mm.getSpread(symbol)
	orderCount := 1 // Place 1 order on each side (reduced from 3 for demo)
	
	for i := 0; i < orderCount; i++ {
		// Buy orders (below current price)
		buyPriceOffset := spread * float64(i+1)
		buyPrice := currentPrice * (1 - buyPriceOffset)
		buyQuantity := mm.getRandomQuantity(symbol)
		
		buyOrder := domain.NewOrder(
			mm.userID,
			symbol,
			domain.OrderSideBuy,
			domain.OrderTypeLimit,
			buyQuantity,
			mm.roundPrice(buyPrice, symbol),
		)
		
		if err := mm.exchange.SubmitOrder(buyOrder); err != nil {
			log.Printf("MM failed to place buy order: %v", err)
		}
		
		// Sell orders (above current price)
		sellPriceOffset := spread * float64(i+1)
		sellPrice := currentPrice * (1 + sellPriceOffset)
		sellQuantity := mm.getRandomQuantity(symbol)
		
		sellOrder := domain.NewOrder(
			mm.userID,
			symbol,
			domain.OrderSideSell,
			domain.OrderTypeLimit,
			sellQuantity,
			mm.roundPrice(sellPrice, symbol),
		)
		
		if err := mm.exchange.SubmitOrder(sellOrder); err != nil {
			log.Printf("MM failed to place sell order: %v", err)
		}
	}
}

func (mm *MarketMaker) getSpread(symbol string) float64 {
	switch symbol {
	case "BTC-USD":
		return 0.001 // 0.1% spread
	case "ETH-USD":
		return 0.0015 // 0.15% spread
	case "SOL-USD":
		return 0.002 // 0.2% spread
	default:
		return 0.002
	}
}

func (mm *MarketMaker) getRandomQuantity(symbol string) float64 {
	base := 0.01
	if symbol == "SOL-USD" {
		base = 0.1
	}
	return base * (1 + rand.Float64())
}

func (mm *MarketMaker) roundPrice(price float64, symbol string) float64 {
	precision := 2.0
	if symbol == "BTC-USD" || symbol == "ETH-USD" {
		precision = 2.0
	}
	multiplier := 1.0
	for i := 0; i < int(precision); i++ {
		multiplier *= 10
	}
	return float64(int(price*multiplier)) / multiplier
}

func (mm *MarketMaker) Stop() {
	mm.cancel()
	log.Printf("Market maker stopped for user: %s", mm.userID)
}
