package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client connected. Total clients: %d", len(h.clients))

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("Client disconnected. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastOrderBook(symbol string, orderBook interface{}) {
	data := map[string]interface{}{
		"type":    "orderbook",
		"symbol":  symbol,
		"data":    orderBook,
	}
	
	message, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal orderbook: %v", err)
		return
	}
	
	h.broadcast <- message
}

func (h *Hub) BroadcastTrade(trade interface{}) {
	data := map[string]interface{}{
		"type": "trade",
		"data": trade,
	}
	
	message, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal trade: %v", err)
		return
	}
	
	h.broadcast <- message
}

func (h *Hub) BroadcastTicker(ticker interface{}) {
	data := map[string]interface{}{
		"type": "ticker",
		"data": ticker,
	}
	
	message, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal ticker: %v", err)
		return
	}
	
	h.broadcast <- message
}

func (h *Hub) BroadcastOrderUpdate(order interface{}) {
	data := map[string]interface{}{
		"type": "order_update",
		"data": order,
	}
	
	message, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal order update: %v", err)
		return
	}
	
	h.broadcast <- message
}

func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
