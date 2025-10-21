import type { OrderBook as OrderBookType } from '../types';

interface OrderBookProps {
  orderBook: OrderBookType | null;
}

export function OrderBook({ orderBook }: OrderBookProps) {
  if (!orderBook) {
    return (
      <div className="bg-gray-900 rounded-lg p-4 h-full flex items-center justify-center">
        <p className="text-gray-500">Loading order book...</p>
      </div>
    );
  }

  const maxBidQty = Math.max(...orderBook.bids.map(b => b.quantity), 0);
  const maxAskQty = Math.max(...orderBook.asks.map(a => a.quantity), 0);

  return (
    <div className="bg-gray-900 rounded-lg p-4 h-full flex flex-col">
      <h2 className="text-lg font-semibold mb-4">Order Book - {orderBook.symbol}</h2>
      
      <div className="flex-1 overflow-hidden flex flex-col">
        {/* Asks (Sell orders) */}
        <div className="flex-1 overflow-y-auto mb-2">
          <div className="grid grid-cols-3 gap-2 text-xs font-semibold text-gray-400 mb-2 sticky top-0 bg-gray-900 pb-2">
            <div>Price</div>
            <div className="text-right">Size</div>
            <div className="text-right">Total</div>
          </div>
          <div className="space-y-1">
            {[...orderBook.asks].reverse().slice(0, 15).map((ask, idx) => {
              const barWidth = (ask.quantity / maxAskQty) * 100;
              return (
                <div key={idx} className="relative grid grid-cols-3 gap-2 text-sm hover:bg-gray-800 cursor-pointer">
                  <div 
                    className="absolute inset-0 bg-red-500/10" 
                    style={{ width: `${barWidth}%` }}
                  />
                  <div className="relative text-red-400">{ask.price.toFixed(2)}</div>
                  <div className="relative text-right">{ask.quantity.toFixed(4)}</div>
                  <div className="relative text-right text-gray-400">{(ask.price * ask.quantity).toFixed(2)}</div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Spread */}
        <div className="py-3 text-center border-y border-gray-800">
          {orderBook.bids.length > 0 && orderBook.asks.length > 0 && (
            <div className="text-sm">
              <span className="text-gray-400">Spread: </span>
              <span className="text-yellow-400 font-semibold">
                {(orderBook.asks[0].price - orderBook.bids[0].price).toFixed(2)}
              </span>
            </div>
          )}
        </div>

        {/* Bids (Buy orders) */}
        <div className="flex-1 overflow-y-auto mt-2">
          <div className="space-y-1">
            {orderBook.bids.slice(0, 15).map((bid, idx) => {
              const barWidth = (bid.quantity / maxBidQty) * 100;
              return (
                <div key={idx} className="relative grid grid-cols-3 gap-2 text-sm hover:bg-gray-800 cursor-pointer">
                  <div 
                    className="absolute inset-0 bg-green-500/10" 
                    style={{ width: `${barWidth}%` }}
                  />
                  <div className="relative text-green-400">{bid.price.toFixed(2)}</div>
                  <div className="relative text-right">{bid.quantity.toFixed(4)}</div>
                  <div className="relative text-right text-gray-400">{(bid.price * bid.quantity).toFixed(2)}</div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    </div>
  );
}
