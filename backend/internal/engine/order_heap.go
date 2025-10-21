package engine

import (
	"github.com/hft-exchange/backend/internal/domain"
)

type OrderHeap struct {
	orders []*domain.Order
	isBuy  bool
}

func (h *OrderHeap) Len() int { return len(h.orders) }

func (h *OrderHeap) Less(i, j int) bool {
	if h.isBuy {
		// For buy orders: higher price has priority
		if h.orders[i].Price != h.orders[j].Price {
			return h.orders[i].Price > h.orders[j].Price
		}
	} else {
		// For sell orders: lower price has priority
		if h.orders[i].Price != h.orders[j].Price {
			return h.orders[i].Price < h.orders[j].Price
		}
	}
	// If prices are equal, earlier timestamp has priority (FIFO)
	return h.orders[i].CreatedAt.Before(h.orders[j].CreatedAt)
}

func (h *OrderHeap) Swap(i, j int) {
	h.orders[i], h.orders[j] = h.orders[j], h.orders[i]
}

func (h *OrderHeap) Push(x interface{}) {
	h.orders = append(h.orders, x.(*domain.Order))
}

func (h *OrderHeap) Pop() interface{} {
	old := h.orders
	n := len(old)
	x := old[n-1]
	h.orders = old[0 : n-1]
	return x
}
