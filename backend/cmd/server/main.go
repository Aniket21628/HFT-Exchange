package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/hft-exchange/backend/internal/api"
	"github.com/hft-exchange/backend/internal/bot"
	"github.com/hft-exchange/backend/internal/cache"
	"github.com/hft-exchange/backend/internal/database"
	"github.com/hft-exchange/backend/internal/domain"
	"github.com/hft-exchange/backend/internal/engine"
	"github.com/hft-exchange/backend/internal/pricefeed"
	"github.com/hft-exchange/backend/internal/repository"
	"github.com/hft-exchange/backend/internal/websocket"
)

// balanceStoreAdapter adapts BalanceRepository to engine.BalanceStore interface
type balanceStoreAdapter struct {
	repo *repository.BalanceRepository
}

func (a *balanceStoreAdapter) GetBalance(userID, asset string) (available, locked float64, err error) {
	balance, err := a.repo.GetBalance(userID, asset)
	if err != nil {
		return 0, 0, err
	}
	return balance.Available, balance.Locked, nil
}

func (a *balanceStoreAdapter) UpdateBalance(userID, asset string, available, locked float64) error {
	return a.repo.UpdateBalance(userID, asset, available, locked)
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Database connection
	dbURL := getEnv("DATABASE_URL", "sqlite://./hft_exchange.db")
	db, err := database.NewDB(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize schema
	if err := db.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Seed demo data
	if err := db.SeedData(); err != nil {
		log.Printf("Warning: Failed to seed data: %v", err)
	}

	// Redis connection
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379/0")
	redisCache, err := cache.NewRedisCache(redisURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v. Continuing without cache.", err)
		redisCache = nil
	}
	if redisCache != nil {
		defer redisCache.Close()
	}

	// Initialize repositories
	orderRepo := repository.NewOrderRepository(db.DB)
	tradeRepo := repository.NewTradeRepository(db.DB)
	balanceRepo := repository.NewBalanceRepository(db.DB)
	tickerRepo := repository.NewTickerRepository(db.DB)

	// Create balance store adapter
	balanceStore := &balanceStoreAdapter{repo: balanceRepo}

	// Initialize exchange
	exchange := engine.NewExchange(tradeRepo, orderRepo, balanceStore)
	exchange.Start()
	defer exchange.Stop()

	// Initialize WebSocket hub (moved up to use in trade callback)
	hub := websocket.NewHub()
	go hub.Run()

	// Set up trade broadcasting callback
	exchange.SetOnTradeCallback(func(trade *domain.Trade) {
		hub.BroadcastTrade(trade)
	})

	// Initialize price simulator
	priceSimulator := pricefeed.NewPriceSimulator(tickerRepo)
	priceSimulator.Start()
	defer priceSimulator.Stop()

	// Connect price updates to exchange and websocket
	priceSimulator.AddUpdateHandler(func(symbol string, price float64) {
		exchange.UpdatePrice(symbol, price)
		
		// Get ticker and broadcast (DB is already updated by simulator)
		if ticker, err := tickerRepo.GetTicker(symbol); err == nil {
			hub.BroadcastTicker(ticker)
		} else {
			log.Printf("❌ Failed to get ticker %s: %v", symbol, err)
		}
		
		// Cache and broadcast order book
		orderBook := exchange.GetOrderBook(symbol, 20)
		if redisCache != nil {
			redisCache.CacheOrderBook(symbol, orderBook)
		}
		hub.BroadcastOrderBook(symbol, orderBook)
	})

	// Start market maker bot
	marketMaker := bot.NewMarketMaker("user-3", exchange, priceSimulator)
	marketMaker.Start()
	defer marketMaker.Stop()

	// Trade broadcasting is now handled by the matching engine directly
	// This polling approach was causing duplicate broadcasts

	// Initialize API handlers
	handler := api.NewHandler(exchange, orderRepo, tradeRepo, balanceRepo, tickerRepo)
	router := api.NewRouter(handler, hub)

	// HTTP server
	port := getEnv("PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
