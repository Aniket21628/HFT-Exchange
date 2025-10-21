import type { Trade } from '../types';
import clsx from 'clsx';

interface TradeHistoryProps {
  trades: Trade[];
}

export function TradeHistory({ trades }: TradeHistoryProps) {
  return (
    <div className="bg-gray-900 rounded-lg p-4 h-full flex flex-col">
      <h2 className="text-lg font-semibold mb-4">Recent Trades</h2>
      
      <div className="flex-1 overflow-y-auto">
        <div className="grid grid-cols-3 gap-4 text-xs font-semibold text-gray-400 mb-2 sticky top-0 bg-gray-900 pb-2">
          <div>Price</div>
          <div className="text-right">Amount</div>
          <div className="text-right">Time</div>
        </div>
        
        <div className="space-y-1">
          {trades.length === 0 ? (
            <div className="text-center text-gray-500 py-8">No recent trades</div>
          ) : (
            trades.map((trade) => {
              const time = new Date(trade.executed_at).toLocaleTimeString();
              const isBuy = trade.buyer_id !== 'user-3'; // Simple heuristic
              
              return (
                <div 
                  key={trade.id} 
                  className="grid grid-cols-3 gap-4 text-sm hover:bg-gray-800 py-1"
                >
                  <div className={clsx(
                    'font-medium',
                    isBuy ? 'text-green-400' : 'text-red-400'
                  )}>
                    {trade.price.toFixed(2)}
                  </div>
                  <div className="text-right">{trade.quantity.toFixed(4)}</div>
                  <div className="text-right text-gray-400 text-xs">{time}</div>
                </div>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
}
