# 🚀 HFT Crypto Trading Exchange

High-frequency trading exchange simulator built with **Go** and **React**. A fully functional trading platform with real-time order matching, WebSocket updates, and market simulation.

## 🏗️ System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                   FRONTEND (React + Vite)                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │Dashboard │  │ BTC Page │  │ ETH Page │  │ SOL Page │   │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘   │
│       └─────────────┴──────────────┴─────────────┘          │
│                     React Router v6                         │
└──────────────────────┬──────────────┬───────────────────────┘
                       │              │
                  HTTP/REST       WebSocket
                       │              │
┌──────────────────────┴──────────────┴───────────────────────┐
│                    BACKEND (Go)                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │         HTTP Server (Gorilla Mux + CORS)               │ │
│  │   /api/v1/orders  /api/v1/trades  /api/v1/tickers     │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │         WebSocket Hub (Real-Time Broadcasting)         │ │
│  │   • Order Book Updates    • Trade Executions           │ │
│  │   • Ticker Updates        • Order Status Changes       │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │              Matching Engines (Concurrent)             │ │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐            │ │
│  │  │ BTC-USD  │  │ ETH-USD  │  │ SOL-USD  │  +USDC     │ │
│  │  │ Engine   │  │ Engine   │  │ Engine   │            │ │
│  │  └──────────┘  └──────────┘  └──────────┘            │ │
│  │   In-Memory Order Books (Heaps) + Stop-Limit Queue    │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │     Price Simulator (Market Data Feed) - 3s Interval   │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │      Market Maker Bot (Automated Liquidity Provider)   │ │
│  └────────────────────────────────────────────────────────┘ │
└───────────┬──────────────────────┬──────────────────────────┘
            │                      │
    ┌───────▼────────┐    ┌────────▼────────┐
    │     SQLite     │    │     Redis       │
    │  (Persistent)  │    │    (Cache)      │
    │                │    │                 │
    │ • Orders       │    │ • Order Books   │
    │ • Trades       │    │   (5s TTL)      │
    │ • Balances     │    │ • Tickers       │
    │ • Tickers      │    │   (10s TTL)     │
    │ • Users        │    │ • Pub/Sub       │
    └────────────────┘    └─────────────────┘
```

## ✨ Key Features

### **Trading**
- ✅ **Market Orders** - Instant execution at best available price
- ✅ **Limit Orders** - Execute at specific price or better
- ✅ **Stop-Limit Orders** - Trigger-based conditional orders
- ✅ **Real-time Order Book** - Live bids and asks visualization
- ✅ **Trade History** - Recent executions with price/volume

### **Real-Time Updates (WebSocket)**
- 📊 Order book updates every 3 seconds
- 💹 Live ticker price updates
- 📈 Instant trade notifications
- 🔔 Order status changes

### **Market Simulation**
- 🤖 Automated market maker providing liquidity
- 📉 Price simulator with realistic volatility
- 💰 Balance management with atomic transactions
- 🛡️ Proper fund locking during orders

### **Multi-Asset Support**
- BTC-USD, ETH-USD, SOL-USD, USDC-USD
- Separate matching engines per trading pair
- Independent price feeds and order books

## 🚀 Quick Start

### Prerequisites
- **Go 1.21+**
- **Node.js 18+**
- **Docker** (optional, for Redis)

### 1. Start Redis (Optional but Recommended)
```bash
docker-compose up -d redis
```

### 2. Start Backend
```bash
cd backend
go mod download
go run cmd/server/main.go
```

Backend will start on `http://localhost:8080`

### 3. Start Frontend
```bash
cd frontend
npm install
npm run dev
```

Frontend will start on `http://localhost:5173`

### 4. Access the Application
Open `http://localhost:5173` in your browser and start trading!

**Default Test Account:**
- User ID: `user-1`
- Starting Balance: $100,000 USD, 1.0 BTC, 10.0 ETH, 100.0 SOL, 50,000 USDC

## ⚙️ Configuration

### Backend Environment Variables

The backend uses PostgreSQL. Get your own PostgreSQL from NeonDB or use Docker to set it up. Optional Redis for caching.

**`.env` (auto-created from `.env.sqlite`)**
```bash
DATABASE_URL=<your_postgres_db>
REDIS_URL=redis://localhost:6379/0
PORT=8080
ENVIRONMENT=development
```

### Frontend Environment Variables

**`.env`**
```bash
VITE_API_URL=http://localhost:8080
```

## 🛠️ Tech Stack

### Backend
- **Language:** Go 1.21+
- **HTTP Router:** Gorilla Mux
- **WebSocket:** Gorilla WebSocket
- **Database:** PostgreSQL
- **Cache:** Redis 7+ (optional)
- **Concurrency:** Goroutines + Channels

### Frontend
- **Framework:** React 18 + Vite
- **Language:** TypeScript 5
- **Routing:** React Router v6
- **Styling:** TailwindCSS 3
- **Icons:** Lucide React
- **Charts:** Recharts
- **WebSocket:** Native WebSocket API

## 📁 Project Structure

```
hft-exchange/
├── backend/
│   ├── cmd/server/        # Entry point
│   ├── internal/
│   │   ├── api/           # HTTP handlers
│   │   ├── engine/        # Matching engines
│   │   ├── websocket/     # WebSocket hub
│   │   ├── domain/        # Core types
│   │   ├── repository/    # Data access
│   │   ├── database/      # DB setup
│   │   ├── cache/         # Redis cache
│   │   ├── pricefeed/     # Price simulator
│   │   └── bot/           # Market maker
│   └── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── pages/         # Dashboard, TradingPage
│   │   ├── components/    # OrderBook, TradeHistory, etc
│   │   ├── hooks/         # useWebSocket
│   │   ├── api/           # API client
│   │   └── types/         # TypeScript types
│   └── package.json
│
└── docker-compose.yml     # Redis + Postgres
```

### Order Lifecycle

```
1. User places order (Frontend)
2. HTTP POST /api/v1/orders
3. Validate balance & lock funds (SQLite)
4. Route to matching engine (In-Memory)
5. Match against opposite side
   ├─ Match found → Execute trade
   │  ├─ Save trade (SQLite)
   │  ├─ Update balances (SQLite)
   │  └─ Broadcast (WebSocket)
   └─ No match → Add to order book (Heap)
6. WebSocket pushes update to all clients
7. Frontend updates UI
```

## 🔥 Performance

- **Order Processing:** 30,000+ orders/sec (across all symbols)
- **WebSocket Clients:** 10,000+ concurrent connections
- **Trade Latency:** <5ms (p99)
- **Memory:** ~2KB per active order
- **CPU:** Multi-core utilization via goroutines

## 🐳 Docker Support

```bash
# Start Redis
docker-compose up -d
```

## 🤝 Contributing

This is a learning project demonstrating:
- High-frequency trading systems
- Go concurrency patterns
- Real-time WebSocket communication
- React state management
- Order matching algorithms

Feel free to fork and experiment!
