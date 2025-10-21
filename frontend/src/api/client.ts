import type { Order, Trade, OrderBook, Ticker, Balance, PlaceOrderRequest } from '../types';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  private async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const url = `${this.baseUrl}${endpoint}`;
    console.log(`🌐 API Request: ${options?.method || 'GET'} ${url}`);
    
    const response = await fetch(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    });

    const data: ApiResponse<T> = await response.json();
    
    if (!data.success || !response.ok) {
      throw new Error(data.error || 'Request failed');
    }

    console.log(`✅ API Response: ${endpoint}`, data.data);
    return data.data as T;
  }

  // Orders
  async placeOrder(request: PlaceOrderRequest): Promise<Order> {
    return this.request<Order>('/api/v1/orders', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async cancelOrder(orderId: string, symbol: string): Promise<void> {
    return this.request<void>(`/api/v1/orders/${orderId}?symbol=${symbol}`, {
      method: 'DELETE',
    });
  }

  async getUserOrders(userId: string, limit = 50): Promise<Order[]> {
    return this.request<Order[]>(`/api/v1/users/${userId}/orders?limit=${limit}`);
  }

  // Trades
  async getRecentTrades(symbol: string, limit = 50): Promise<Trade[]> {
    return this.request<Trade[]>(`/api/v1/trades/${symbol}?limit=${limit}`);
  }

  async getUserTrades(userId: string, limit = 50): Promise<Trade[]> {
    return this.request<Trade[]>(`/api/v1/users/${userId}/trades?limit=${limit}`);
  }

  // Order Book
  async getOrderBook(symbol: string, depth = 20): Promise<OrderBook> {
    return this.request<OrderBook>(`/api/v1/orderbook/${symbol}?depth=${depth}`);
  }

  // Balances
  async getUserBalances(userId: string): Promise<Balance[]> {
    return this.request<Balance[]>(`/api/v1/users/${userId}/balances`);
  }

  // Tickers
  async getTicker(symbol: string): Promise<Ticker> {
    return this.request<Ticker>(`/api/v1/tickers/${symbol}`);
  }

  async getAllTickers(): Promise<Ticker[]> {
    return this.request<Ticker[]>('/api/v1/tickers');
  }

  // Symbols
  async getSymbols(): Promise<string[]> {
    return this.request<string[]>('/api/v1/symbols');
  }

  // Health check
  async healthCheck(): Promise<{ status: string }> {
    return this.request<{ status: string }>('/health');
  }
}

export const apiClient = new ApiClient(API_URL);
