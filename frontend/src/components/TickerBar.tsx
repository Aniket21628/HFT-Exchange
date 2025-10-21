import type { Ticker } from '../types';
import { TrendingUp, TrendingDown } from 'lucide-react';
import clsx from 'clsx';

interface TickerBarProps {
  tickers: Ticker[];
  selectedSymbol: string;
  onSelectSymbol: (symbol: string) => void;
}

export function TickerBar({ tickers, selectedSymbol, onSelectSymbol }: TickerBarProps) {
  return (
    <div className="bg-gray-900 border-b border-gray-800 px-4 py-3">
      <div className="flex gap-6 overflow-x-auto">
        {tickers.map((ticker) => {
          const isPositive = ticker.change_24h >= 0;
          const isSelected = ticker.symbol === selectedSymbol;

          return (
            <button
              key={ticker.symbol}
              onClick={() => onSelectSymbol(ticker.symbol)}
              className={clsx(
                'flex items-center gap-3 px-4 py-2 rounded-lg transition-colors whitespace-nowrap',
                isSelected 
                  ? 'bg-primary-600 text-white' 
                  : 'hover:bg-gray-800'
              )}
            >
              <div className="text-left">
                <div className="font-semibold">{ticker.symbol}</div>
                <div className="text-2xl font-bold">${ticker.price.toFixed(2)}</div>
              </div>
              <div className={clsx(
                'flex items-center gap-1 text-sm font-medium',
                isPositive ? 'text-green-400' : 'text-red-400'
              )}>
                {isPositive ? <TrendingUp size={16} /> : <TrendingDown size={16} />}
                <span>{Math.abs(ticker.change_24h).toFixed(2)}%</span>
              </div>
            </button>
          );
        })}
      </div>
    </div>
  );
}
