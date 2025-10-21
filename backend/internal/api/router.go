package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	ws "github.com/hft-exchange/backend/internal/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

func NewRouter(handler *Handler, hub *ws.Hub) http.Handler {
	r := mux.NewRouter()

	// Health check
	r.HandleFunc("/health", handler.HealthCheck).Methods("GET")

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Orders
	api.HandleFunc("/orders", handler.PlaceOrder).Methods("POST")
	api.HandleFunc("/orders/{id}", handler.CancelOrder).Methods("DELETE")
	api.HandleFunc("/users/{userId}/orders", handler.GetUserOrders).Methods("GET")

	// Trades
	api.HandleFunc("/trades/{symbol}", handler.GetRecentTrades).Methods("GET")
	api.HandleFunc("/users/{userId}/trades", handler.GetUserTrades).Methods("GET")

	// Order book
	api.HandleFunc("/orderbook/{symbol}", handler.GetOrderBook).Methods("GET")

	// Balances
	api.HandleFunc("/users/{userId}/balances", handler.GetUserBalances).Methods("GET")

	// Tickers
	api.HandleFunc("/tickers", handler.GetAllTickers).Methods("GET")
	api.HandleFunc("/tickers/{symbol}", handler.GetTicker).Methods("GET")

	// Symbols
	api.HandleFunc("/symbols", handler.GetSymbols).Methods("GET")

	// WebSocket
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(hub, w, r)
	})

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	return c.Handler(r)
}

func handleWebSocket(hub *ws.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := ws.NewClient(hub, conn)
	hub.Register <- client

	client.Start()
}
