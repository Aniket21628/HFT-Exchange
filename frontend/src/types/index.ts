export type OrderSide = 'BUY' | 'SELL';
export type OrderType = 'LIMIT' | 'MARKET' | 'STOP_LIMIT';
export type OrderStatus = 'PENDING' | 'PARTIAL' | 'FILLED' | 'CANCELLED' | 'REJECTED';

export interface Order {
  id: string;
  user_id: string;
  symbol: string;
  side: OrderSide;
  type: OrderType;
  quantity: number;
  price: number;
  stop_price?: number;
  filled_quantity: number;
  remaining_qty: number;
  status: OrderStatus;
  created_at: string;
  updated_at: string;
  time_in_force: string;
}

export interface Trade {
  id: string;
  symbol: string;
  buy_order_id: string;
  sell_order_id: string;
  buyer_id: string;
  seller_id: string;
  price: number;
  quantity: number;
  executed_at: string;
  maker_order_id: string;
  taker_order_id: string;
}

export interface OrderBookLevel {
  price: number;
  quantity: number;
  orders: number;
}

export interface OrderBook {
  symbol: string;
  bids: OrderBookLevel[];
  asks: OrderBookLevel[];
  timestamp: string;
}

export interface Ticker {
  symbol: string;
  price: number;
  high_24h: number;
  low_24h: number;
  volume_24h: number;
  change_24h: number;
  updated_at: string;
}

export interface Balance {
  UserID: string;
  Asset: string;
  Available: number;
  Locked: number;
  UpdatedAt: string;
}

export interface WSMessage {
  type: 'orderbook' | 'trade' | 'ticker' | 'order_update';
  symbol?: string;
  data: any;
}

export interface PlaceOrderRequest {
  user_id: string;
  symbol: string;
  side: OrderSide;
  type: OrderType;
  quantity: number;
  price: number;
  stop_price?: number;
}
