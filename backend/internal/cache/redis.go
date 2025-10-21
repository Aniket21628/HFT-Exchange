package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/hft-exchange/backend/internal/domain"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisCache(url string) (*RedisCache, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)
	
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

func (r *RedisCache) CacheOrderBook(symbol string, orderBook *domain.OrderBook) error {
	data, err := json.Marshal(orderBook)
	if err != nil {
		return fmt.Errorf("failed to marshal order book: %w", err)
	}

	key := fmt.Sprintf("orderbook:%s", symbol)
	return r.client.Set(r.ctx, key, data, 5*time.Second).Err()
}

func (r *RedisCache) GetOrderBook(symbol string) (*domain.OrderBook, error) {
	key := fmt.Sprintf("orderbook:%s", symbol)
	data, err := r.client.Get(r.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get order book: %w", err)
	}

	var orderBook domain.OrderBook
	if err := json.Unmarshal(data, &orderBook); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order book: %w", err)
	}

	return &orderBook, nil
}

func (r *RedisCache) CacheTicker(symbol string, ticker *domain.Ticker) error {
	data, err := json.Marshal(ticker)
	if err != nil {
		return fmt.Errorf("failed to marshal ticker: %w", err)
	}

	key := fmt.Sprintf("ticker:%s", symbol)
	return r.client.Set(r.ctx, key, data, 10*time.Second).Err()
}

func (r *RedisCache) GetTicker(symbol string) (*domain.Ticker, error) {
	key := fmt.Sprintf("ticker:%s", symbol)
	data, err := r.client.Get(r.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get ticker: %w", err)
	}

	var ticker domain.Ticker
	if err := json.Unmarshal(data, &ticker); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ticker: %w", err)
	}

	return &ticker, nil
}

func (r *RedisCache) PublishTrade(trade *domain.Trade) error {
	data, err := json.Marshal(trade)
	if err != nil {
		return fmt.Errorf("failed to marshal trade: %w", err)
	}

	channel := fmt.Sprintf("trades:%s", trade.Symbol)
	return r.client.Publish(r.ctx, channel, data).Err()
}

func (r *RedisCache) SubscribeTrades(symbol string) *redis.PubSub {
	channel := fmt.Sprintf("trades:%s", symbol)
	return r.client.Subscribe(r.ctx, channel)
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}
