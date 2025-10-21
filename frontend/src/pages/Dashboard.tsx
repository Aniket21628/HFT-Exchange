import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { TrendingUp, TrendingDown, Activity } from 'lucide-react';
import clsx from 'clsx';
import { apiClient } from '../api/client';
import { useWebSocket } from '../hooks/useWebSocket';
import type { Ticker, WSMessage } from '../types';

export function Dashboard() {
  const navigate = useNavigate();
  const [tickers, setTickers] = useState<Ticker[]>([]);

  // WebSocket message handler - only update tickers
  const handleWSMessage = (message: WSMessage) => {
    if (message.type === 'ticker') {
      setTickers(prev => {
        const index = prev.findIndex(t => t.symbol === message.data.symbol);
        if (index >= 0) {
          const newTickers = [...prev];
          newTickers[index] = message.data;
          return newTickers;
        }
        return prev;
      });
    }
  };

  const { isConnected } = useWebSocket(handleWSMessage);

  // Load tickers
  useEffect(() => {
    const loadTickers = async () => {
      try {
        const tickersData = await apiClient.getAllTickers();
        setTickers(tickersData);
      } catch (error) {
        console.error('Failed to load tickers:', error);
      }
    };

    loadTickers();
  }, []);

  const handleSelectSymbol = (symbol: string) => {
    navigate(`/trade/${symbol}`);
  };

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      {/* Header */}
      <header className="bg-gray-900 border-b border-gray-800">
        <div className="px-6 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold">HFT Exchange</h1>
              <p className="text-gray-400 mt-1">High-Frequency Trading Platform</p>
            </div>
            <div className="flex items-center gap-2">
              <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`}></div>
              <span className="text-sm text-gray-400 capitalize">
                {isConnected ? 'connected' : 'disconnected'}
              </span>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="p-8">
        <div className="mb-8">
          <h2 className="text-2xl font-semibold mb-6 flex items-center gap-3">
            <Activity className="text-primary-500" />
            Select Trading Pair
          </h2>
          
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            {tickers.map((ticker) => {
              const isPositive = ticker.change_24h >= 0;
              
              return (
                <button
                  key={ticker.symbol}
                  onClick={() => handleSelectSymbol(ticker.symbol)}
                  className="bg-gray-900 hover:bg-gray-800 border border-gray-800 hover:border-primary-600 rounded-xl p-6 transition-all transform hover:scale-105"
                >
                  <div className="flex items-start justify-between mb-4">
                    <div>
                      <h3 className="text-xl font-bold">{ticker.symbol}</h3>
                      <p className="text-gray-500 text-sm mt-1">
                        {ticker.symbol.split('-')[0]} / {ticker.symbol.split('-')[1]}
                      </p>
                    </div>
                    <div className={clsx(
                      'p-2 rounded-lg',
                      isPositive ? 'bg-green-500/10' : 'bg-red-500/10'
                    )}>
                      {isPositive ? 
                        <TrendingUp className="text-green-400" size={24} /> : 
                        <TrendingDown className="text-red-400" size={24} />
                      }
                    </div>
                  </div>

                  <div className="space-y-2">
                    <div>
                      <p className="text-3xl font-bold">${ticker.price.toFixed(2)}</p>
                    </div>
                    
                    <div className="flex items-center justify-between pt-3 border-t border-gray-800">
                      <span className="text-gray-500 text-sm">24h Change</span>
                      <span className={clsx(
                        'font-semibold',
                        isPositive ? 'text-green-400' : 'text-red-400'
                      )}>
                        {isPositive ? '+' : ''}{ticker.change_24h.toFixed(2)}%
                      </span>
                    </div>

                    <div className="flex items-center justify-between">
                      <span className="text-gray-500 text-sm">24h High</span>
                      <span className="text-gray-300">${ticker.high_24h.toFixed(2)}</span>
                    </div>

                    <div className="flex items-center justify-between">
                      <span className="text-gray-500 text-sm">24h Low</span>
                      <span className="text-gray-300">${ticker.low_24h.toFixed(2)}</span>
                    </div>
                  </div>

                  <div className="mt-4 pt-4 border-t border-gray-800">
                    <span className="text-primary-500 font-semibold text-sm">
                      Start Trading →
                    </span>
                  </div>
                </button>
              );
            })}
          </div>
        </div>

        {/* Info Section */}
        <div className="mt-12 bg-gray-900/50 border border-gray-800 rounded-xl p-6">
          <h3 className="text-xl font-semibold mb-4">Platform Features</h3>
          <div className="grid md:grid-cols-3 gap-6">
            <div>
              <h4 className="font-semibold text-primary-500 mb-2">Real-Time Trading</h4>
              <p className="text-gray-400 text-sm">
                Execute trades instantly with our high-frequency matching engine
              </p>
            </div>
            <div>
              <h4 className="font-semibold text-primary-500 mb-2">Live Order Book</h4>
              <p className="text-gray-400 text-sm">
                View market depth and liquidity in real-time via WebSocket
              </p>
            </div>
            <div>
              <h4 className="font-semibold text-primary-500 mb-2">Advanced Orders</h4>
              <p className="text-gray-400 text-sm">
                Market orders, limit orders, and stop-limit orders supported
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
