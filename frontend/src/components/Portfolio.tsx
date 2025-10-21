import type { Balance, Order } from '../types';
import { Wallet, Activity } from 'lucide-react';
import clsx from 'clsx';

interface PortfolioProps {
  balances: Balance[];
  orders: Order[];
}

export function Portfolio({ balances, orders }: PortfolioProps) {
  const totalValue = balances.reduce((sum, b) => sum + b.Available + b.Locked, 0);
  const openOrders = orders.filter(o => o.status === 'PENDING' || o.status === 'PARTIAL');

  return (
    <div className="bg-gray-900 rounded-lg p-4">
      <div className="flex items-center gap-2 mb-4">
        <Wallet className="text-primary-500" size={20} />
        <h2 className="text-lg font-semibold">Portfolio</h2>
      </div>

      {/* Total Value */}
      <div className="bg-gray-800 rounded-lg p-4 mb-4">
        <div className="text-sm text-gray-400 mb-1">Total Value</div>
        <div className="text-2xl font-bold">${totalValue.toFixed(2)}</div>
      </div>

      {/* Balances */}
      <div className="mb-6">
        <h3 className="text-sm font-semibold text-gray-400 mb-3">Assets</h3>
        <div className="space-y-2">
          {balances.length === 0 ? (
            <div className="text-center text-gray-500 py-4">No balances</div>
          ) : (
            balances.map((balance) => (
              <div 
                key={balance.Asset} 
                className="flex justify-between items-center bg-gray-800 rounded-lg p-3"
              >
                <div>
                  <div className="font-medium">{balance.Asset}</div>
                  <div className="text-xs text-gray-400">
                    Locked: {balance.Locked.toFixed(4)}
                  </div>
                </div>
                <div className="text-right">
                  <div className="font-semibold">{balance.Available.toFixed(4)}</div>
                  <div className="text-xs text-gray-400">Available</div>
                </div>
              </div>
            ))
          )}
        </div>
      </div>

      {/* Open Orders */}
      <div>
        <div className="flex items-center gap-2 mb-3">
          <Activity size={16} className="text-gray-400" />
          <h3 className="text-sm font-semibold text-gray-400">
            Open Orders ({openOrders.length})
          </h3>
        </div>
        <div className="space-y-2 max-h-64 overflow-y-auto">
          {openOrders.length === 0 ? (
            <div className="text-center text-gray-500 py-4 text-sm">No open orders</div>
          ) : (
            openOrders.map((order) => (
              <div 
                key={order.id} 
                className="bg-gray-800 rounded-lg p-3 text-sm"
              >
                <div className="flex justify-between items-center mb-2">
                  <span className={clsx(
                    'font-semibold px-2 py-0.5 rounded text-xs',
                    order.side === 'BUY' 
                      ? 'bg-green-500/20 text-green-400'
                      : 'bg-red-500/20 text-red-400'
                  )}>
                    {order.side}
                  </span>
                  <span className="text-gray-400 text-xs">{order.type}</span>
                </div>
                <div className="grid grid-cols-2 gap-2 text-xs">
                  <div>
                    <span className="text-gray-400">Price:</span>
                    <span className="ml-1 font-medium">${order.price.toFixed(2)}</span>
                  </div>
                  <div>
                    <span className="text-gray-400">Amount:</span>
                    <span className="ml-1 font-medium">{order.quantity.toFixed(4)}</span>
                  </div>
                  <div className="col-span-2">
                    <span className="text-gray-400">Filled:</span>
                    <span className="ml-1 font-medium">
                      {((order.filled_quantity / order.quantity) * 100).toFixed(0)}%
                    </span>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}
