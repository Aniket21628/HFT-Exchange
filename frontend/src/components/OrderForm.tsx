import { useState } from 'react';
import type { OrderSide, OrderType, PlaceOrderRequest } from '../types';
import { apiClient } from '../api/client';
import clsx from 'clsx';

interface OrderFormProps {
  symbol: string;
  currentPrice: number;
  onOrderPlaced?: () => void;
}

export function OrderForm({ symbol, currentPrice, onOrderPlaced }: OrderFormProps) {
  const [side, setSide] = useState<OrderSide>('BUY');
  const [orderType, setOrderType] = useState<OrderType>('LIMIT');
  const [price, setPrice] = useState(currentPrice.toString());
  const [stopPrice, setStopPrice] = useState(currentPrice.toString());
  const [quantity, setQuantity] = useState('0.01');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(false);
    setLoading(true);

    try {
      const request: PlaceOrderRequest = {
        user_id: 'user-1', // In real app, get from auth
        symbol,
        side,
        type: orderType,
        quantity: parseFloat(quantity),
        price: orderType === 'MARKET' ? 0 : parseFloat(price),
      };

      if (orderType === 'STOP_LIMIT') {
        request.stop_price = parseFloat(stopPrice);
      }

      await apiClient.placeOrder(request);
      setSuccess(true);
      setQuantity('0.01');
      
      if (onOrderPlaced) {
        onOrderPlaced();
      }

      setTimeout(() => setSuccess(false), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to place order');
    } finally {
      setLoading(false);
    }
  };

  const total = orderType === 'MARKET' 
    ? parseFloat(quantity) * currentPrice 
    : parseFloat(quantity) * parseFloat(price || '0');

  return (
    <div className="bg-gray-900 rounded-lg p-4">
      <h2 className="text-lg font-semibold mb-4">Place Order</h2>

      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Side selector */}
        <div className="grid grid-cols-2 gap-2">
          <button
            type="button"
            onClick={() => setSide('BUY')}
            className={clsx(
              'py-2 rounded-lg font-semibold transition-colors',
              side === 'BUY'
                ? 'bg-green-600 text-white'
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            )}
          >
            Buy
          </button>
          <button
            type="button"
            onClick={() => setSide('SELL')}
            className={clsx(
              'py-2 rounded-lg font-semibold transition-colors',
              side === 'SELL'
                ? 'bg-red-600 text-white'
                : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
            )}
          >
            Sell
          </button>
        </div>

        {/* Order type */}
        <div>
          <label className="block text-sm font-medium mb-2">Order Type</label>
          <select
            value={orderType}
            onChange={(e) => setOrderType(e.target.value as OrderType)}
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 focus:outline-none focus:border-primary-500"
          >
            <option value="LIMIT">Limit</option>
            <option value="MARKET">Market</option>
            <option value="STOP_LIMIT">Stop Limit</option>
          </select>
        </div>

        {/* Stop Price (for Stop-Limit) */}
        {orderType === 'STOP_LIMIT' && (
          <div>
            <label className="block text-sm font-medium mb-2">
              Stop Price (Trigger)
              <span className="text-xs text-gray-400 ml-2">
                {side === 'BUY' ? '(Activate when price ≥)' : '(Activate when price ≤)'}
              </span>
            </label>
            <input
              type="number"
              step="0.01"
              value={stopPrice}
              onChange={(e) => setStopPrice(e.target.value)}
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 focus:outline-none focus:border-primary-500"
              required
              placeholder={currentPrice.toFixed(2)}
            />
          </div>
        )}

        {/* Limit Price */}
        {orderType !== 'MARKET' && (
          <div>
            <label className="block text-sm font-medium mb-2">
              {orderType === 'STOP_LIMIT' ? 'Limit Price (Execute at)' : 'Price'}
            </label>
            <input
              type="number"
              step="0.01"
              value={price}
              onChange={(e) => setPrice(e.target.value)}
              className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 focus:outline-none focus:border-primary-500"
              required
              placeholder={currentPrice.toFixed(2)}
            />
          </div>
        )}

        {/* Quantity */}
        <div>
          <label className="block text-sm font-medium mb-2">Amount</label>
          <input
            type="number"
            step="0.0001"
            value={quantity}
            onChange={(e) => setQuantity(e.target.value)}
            className="w-full bg-gray-800 border border-gray-700 rounded-lg px-3 py-2 focus:outline-none focus:border-primary-500"
            required
            min="0"
          />
        </div>

        {/* Total */}
        <div className="bg-gray-800 rounded-lg p-3">
          <div className="flex justify-between text-sm">
            <span className="text-gray-400">Total</span>
            <span className="font-semibold">{total.toFixed(2)} USD</span>
          </div>
        </div>

        {/* Submit button */}
        <button
          type="submit"
          disabled={loading}
          className={clsx(
            'w-full py-3 rounded-lg font-semibold transition-colors',
            side === 'BUY'
              ? 'bg-green-600 hover:bg-green-700 text-white'
              : 'bg-red-600 hover:bg-red-700 text-white',
            loading && 'opacity-50 cursor-not-allowed'
          )}
        >
          {loading ? 'Placing Order...' : `${side} ${symbol.split('-')[0]}`}
        </button>

        {/* Error/Success messages */}
        {error && (
          <div className="bg-red-500/10 border border-red-500 text-red-400 rounded-lg p-3 text-sm">
            {error}
          </div>
        )}
        {success && (
          <div className="bg-green-500/10 border border-green-500 text-green-400 rounded-lg p-3 text-sm">
            Order placed successfully!
          </div>
        )}
      </form>
    </div>
  );
}
