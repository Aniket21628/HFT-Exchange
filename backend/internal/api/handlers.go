package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/hft-exchange/backend/internal/domain"
	"github.com/hft-exchange/backend/internal/engine"
	"github.com/hft-exchange/backend/internal/repository"
)

type Handler struct {
	exchange     *engine.Exchange
	orderRepo    *repository.OrderRepository
	tradeRepo    *repository.TradeRepository
	balanceRepo  *repository.BalanceRepository
	tickerRepo   *repository.TickerRepository
}

func NewHandler(
	exchange *engine.Exchange,
	orderRepo *repository.OrderRepository,
	tradeRepo *repository.TradeRepository,
	balanceRepo *repository.BalanceRepository,
	tickerRepo *repository.TickerRepository,
) *Handler {
	return &Handler{
		exchange:    exchange,
		orderRepo:   orderRepo,
		tradeRepo:   tradeRepo,
		balanceRepo: balanceRepo,
		tickerRepo:  tickerRepo,
	}
}

type PlaceOrderRequest struct {
	UserID    string  `json:"user_id"`
	Symbol    string  `json:"symbol"`
	Side      string  `json:"side"`
	Type      string  `json:"type"`
	Quantity  float64 `json:"quantity"`
	Price     float64 `json:"price"`
	StopPrice float64 `json:"stop_price,omitempty"`
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var req PlaceOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, Response{Success: false, Error: "Invalid request body"})
		return
	}

	order := domain.NewOrder(
		req.UserID,
		req.Symbol,
		domain.OrderSide(req.Side),
		domain.OrderType(req.Type),
		req.Quantity,
		req.Price,
	)

	if req.StopPrice > 0 {
		order.StopPrice = req.StopPrice
	}

	if err := h.exchange.SubmitOrder(order); err != nil {
		respondJSON(w, http.StatusInternalServerError, Response{Success: false, Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, Response{Success: true, Data: order})
}

func (h *Handler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]
	symbol := r.URL.Query().Get("symbol")

	success := h.exchange.CancelOrder(orderID, symbol)
	if !success {
		respondJSON(w, http.StatusNotFound, Response{Success: false, Error: "Order not found"})
		return
	}

	respondJSON(w, http.StatusOK, Response{Success: true})
}

func (h *Handler) GetOrderBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]
	
	depthStr := r.URL.Query().Get("depth")
	depth := 20
	if depthStr != "" {
		if d, err := strconv.Atoi(depthStr); err == nil {
			depth = d
		}
	}

	orderBook := h.exchange.GetOrderBook(symbol, depth)
	respondJSON(w, http.StatusOK, Response{Success: true, Data: orderBook})
}

func (h *Handler) GetRecentTrades(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]
	
	limitStr := r.URL.Query().Get("limit")
	limit := 20 // Default to 20 trades (was 50)
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
			// Cap at 20 max to prevent UI overflow
			if limit > 20 {
				limit = 20
			}
		}
	}

	trades, err := h.tradeRepo.GetRecentTrades(symbol, limit)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, Response{Success: false, Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, Response{Success: true, Data: trades})
}

func (h *Handler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]
	
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	orders, err := h.orderRepo.GetOrdersByUser(userID, limit)
	if err != nil {
		log.Printf("ERROR getting orders: %v", err)
		respondJSON(w, http.StatusInternalServerError, Response{Success: false, Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, Response{Success: true, Data: orders})
}

func (h *Handler) GetUserTrades(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]
	
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	trades, err := h.tradeRepo.GetUserTrades(userID, limit)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, Response{Success: false, Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, Response{Success: true, Data: trades})
}

func (h *Handler) GetUserBalances(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	balances, err := h.balanceRepo.GetAllBalances(userID)
	if err != nil {
		log.Printf("ERROR getting balances: %v", err)
		respondJSON(w, http.StatusInternalServerError, Response{Success: false, Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, Response{Success: true, Data: balances})
}

func (h *Handler) GetTicker(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]

	ticker, err := h.tickerRepo.GetTicker(symbol)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, Response{Success: false, Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, Response{Success: true, Data: ticker})
}

func (h *Handler) GetAllTickers(w http.ResponseWriter, r *http.Request) {
	tickers, err := h.tickerRepo.GetAllTickers()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, Response{Success: false, Error: err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, Response{Success: true, Data: tickers})
}

func (h *Handler) GetSymbols(w http.ResponseWriter, r *http.Request) {
	symbols := h.exchange.GetAllSymbols()
	respondJSON(w, http.StatusOK, Response{Success: true, Data: symbols})
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, Response{Success: true, Data: map[string]string{"status": "healthy"}})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}
