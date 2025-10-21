import { useState, useEffect, useCallback } from 'react';
import { useParams } from 'react-router-dom';
import { OrderBook } from '../components/OrderBook';
import { TradeHistory } from '../components/TradeHistory';
import { OrderForm } from '../components/OrderForm';
import { Portfolio } from '../components/Portfolio';
import { useWebSocket } from '../hooks/useWebSocket';
import { apiClient } from '../api/client';
import type { Order, Trade, OrderBook as OrderBookType, Ticker, Balance, WSMessage } from '../types';

export function TradingPage() {
  const { symbol } = useParams<{ symbol: string }>();
  const tradingSymbol = symbol || 'BTC-USD';

  const [orderBook, setOrderBook] = useState<OrderBookType | null>(null);
  const [trades, setTrades] = useState<Trade[]>([]);
  const [balances, setBalances] = useState<Balance[]>([]);
  const [orders, setOrders] = useState<Order[]>([]);
  const [ticker, setTicker] = useState<Ticker | null>(null);

  const currentPrice = ticker?.price || 0;

  // WebSocket message handler
  const handleWSMessage = useCallback((message: WSMessage) => {
    switch (message.type) {
      case 'orderbook':
        if (message.symbol === tradingSymbol) {
          console.log(`📦 [${tradingSymbol}] Received orderbook`);
          setOrderBook(message.data);
        }
        break;
      case 'trade':
        if (message.data.symbol === tradingSymbol) {
          setTrades(prev => {
            const tradeExists = prev.some(t => t.id === message.data.id);
            if (tradeExists) return prev;
            console.log(`📊 [${tradingSymbol}] Adding trade`);
            return [message.data, ...prev].slice(0, 20);
          });
        }
        break;
      case 'ticker':
        if (message.data.symbol === tradingSymbol) {
          console.log(`💹 [${tradingSymbol}] Updating ticker: $${message.data.price.toFixed(2)}`);
          setTicker(message.data);
        }
        break;
      case 'order_update':
        setOrders(prev => {
          const index = prev.findIndex(o => o.id === message.data.id);
          if (index >= 0) {
            const newOrders = [...prev];
            if (message.data.status === 'filled' || message.data.status === 'cancelled') {
              newOrders.splice(index, 1);
            } else {
              newOrders[index] = message.data;
            }
            return newOrders;
          }
          return [message.data, ...prev];
        });
        break;
    }
  }, [tradingSymbol]);

  const { isConnected } = useWebSocket(handleWSMessage);

  // Load initial data for this symbol
  useEffect(() => {
    const loadData = async () => {
      console.log(`🔄 [${tradingSymbol}] Loading initial data`);
      
      try {
        const [tickerData, balancesData, ordersData, tradesData, orderBookData] = await Promise.all([
          apiClient.getTicker(tradingSymbol),
          apiClient.getUserBalances('user-1'),
          apiClient.getUserOrders('user-1', 20),
          apiClient.getRecentTrades(tradingSymbol, 20),
          apiClient.getOrderBook(tradingSymbol, 20),
        ]);

        console.log(`✅ [${tradingSymbol}] Data loaded - Ticker: $${tickerData.price.toFixed(2)}`);

        setTicker(tickerData);
        setBalances(balancesData);
        setOrders(ordersData);
        setTrades(tradesData);
        setOrderBook(orderBookData);
      } catch (error) {
        console.error(`❌ [${tradingSymbol}] Failed to load data:`, error);
      }
    };

    loadData();
  }, [tradingSymbol]);

  // Refresh data periodically
  useEffect(() => {
    const interval = setInterval(async () => {
      try {
        const [balancesData, ordersData] = await Promise.all([
          apiClient.getUserBalances('user-1'),
          apiClient.getUserOrders('user-1', 20),
        ]);
        setBalances(balancesData);
        setOrders(ordersData);
      } catch (error) {
        console.error('Failed to refresh data:', error);
      }
    }, 2000);

    return () => clearInterval(interval);
  }, []);

  const handleOrderPlaced = async () => {
    try {
      const [ordersData, balancesData] = await Promise.all([
        apiClient.getUserOrders('user-1', 20),
        apiClient.getUserBalances('user-1'),
      ]);
      setOrders(ordersData);
      setBalances(balancesData);
    } catch (error) {
      console.error('Failed to refresh after order placement:', error);
    }
  };

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      {/* Header */}
      <header className="bg-gray-900 border-b border-gray-800">
        <div className="px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold">{tradingSymbol}</h1>
              {ticker && (
                <div className="flex items-center gap-4 mt-1">
                  <span className="text-3xl font-bold">${ticker.price.toFixed(2)}</span>
                  <span className={ticker.change_24h >= 0 ? 'text-green-400' : 'text-red-400'}>
                    {ticker.change_24h >= 0 ? '+' : ''}{ticker.change_24h.toFixed(2)}%
                  </span>
                </div>
              )}
            </div>
            <div className="flex items-center gap-4">
              <div className="flex items-center gap-2">
                <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`}></div>
                <span className="text-sm text-gray-400 capitalize">
                  {isConnected ? 'connected' : 'disconnected'}
                </span>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="p-6">
        <div className="grid grid-cols-12 gap-6 h-[calc(100vh-220px)]">
          {/* Left: Order Book */}
          <div className="col-span-3">
            <OrderBook orderBook={orderBook} />
          </div>

          {/* Center: Trades */}
          <div className="col-span-3">
            <TradeHistory trades={trades} />
          </div>

          {/* Right: Order Form */}
          <div className="col-span-3">
            <OrderForm 
              symbol={tradingSymbol}
              currentPrice={currentPrice}
              onOrderPlaced={handleOrderPlaced}
            />
          </div>

          {/* Right Panel: Portfolio */}
          <div className="col-span-3">
            <Portfolio balances={balances} orders={orders} />
          </div>
        </div>
      </div>
    </div>
  );
}
