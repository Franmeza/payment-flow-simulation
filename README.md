# Simulated Payment Flow + Merchant Dashboard

A hardware + software simulation of a real card payment flow with a merchant-facing dashboard.

A physical NFC card is tapped on an ESP32 terminal, which triggers a full authorization cycle across three simulated Go services: acquirer, card network router, and issuer. A React dashboard visualizes live transaction data and processing-fee metrics from SQLite.

---

## Architecture

```
[ESP32 + PN532 Terminal]
        |
        | HTTP POST /authorize
        v
[Acquirer Service :8080]
        |
        | HTTP POST /route
        v
[Network Router :8081]
        |
        | HTTP POST /authorize
        v
[Issuer Service :8082]
        |
        | approve / decline
        v
[Network Router :8081]
        |
        v
[Acquirer Service :8080]
        |
        | HTTP response
        v
[ESP32 Display — APPROVED / DECLINED]

[React Dashboard (Vite) :5173]
        |
        | GET /api/transactions, /api/stats (proxied)
        v
[Issuer Service :8082]
```

---

## Demo

![Payment Flow Simulation Demo](demo_hq.gif)

## Hardware

| Component | Details |
|---|---|
| ESP32 board | With onboard SSD1306 OLED display |
| PN532 NFC module | POCREATION kit — I2C mode |
| MIFARE Classic 1K card | UID: 2E:F8:14:07|
| MIFARE keyfob | UID: 1B:5E:32:07 |

**Wiring (I2C):**

| PN532 Pin | ESP32 Pin |
|---|---|
| VCC | 3.3V |
| GND | GND |
| SDA | D21 |
| SCL | D22 |

---

## Services

### Acquirer (port 8080)
Simulates the merchant's bank. Receives the auth request from the terminal, validates the merchant exists and is active, then forwards the request to the network router. Returns the final approve/decline back to the terminal. Supports optional idempotency via an `idempotency_key` field — duplicate requests within 24 hours return the cached response with an `Idempotent-Replayed: true` header.

### Network Router (port 8081)
Simulates Visa/Mastercard. Routes the request to the correct issuer based on card prefix. Adds a small artificial latency to simulate real network conditions.

### Issuer (port 8082)
Simulates the cardholder's bank. Holds the card database, checks balance, runs fraud rules, makes the approve/decline decision, and exposes read-only dashboard endpoints:

- `GET /transactions?limit=100` — recent transaction history
- `GET /stats` — aggregated transaction metrics

### Dashboard UI (port 5173)
A React + TypeScript app that displays:

- approval/decline KPIs
- gross approved volume
- estimated processing fees (`1.68% + $0.08`)
- estimated net volume
- paginated transaction history with mobile-optimized cards

---

## Fraud Rules (Issuer)

- Unknown card UID → decline
- Blocked or non-active card status → decline
- Amount ≤ $0 → decline (invalid amount)
- Amount > $1,000 → decline (transaction limit)
- Insufficient balance → decline
- ≥ 3 attempts in 60 seconds (same card) → decline (velocity check)

---

## Project Structure

```
payment-flow-simulation/
├── cmd/
│   ├── acquirer/       # Acquirer service (port 8080)
│   ├── network/        # Network router (port 8081)
│   └── issuer/         # Issuer service (port 8082)
├── internal/
│   ├── models/         # Shared structs (Card, Merchant, AuthRequest, AuthResponse, Transaction)
│   ├── db/             # SQLite card store (migrations, seed, CRUD)
│   ├── rules/          # Authorization rule engine
│   ├── fraud/          # In-memory velocity tracker (sliding 60s window)
│   ├── idempotency/    # SQLite idempotency store with 24h expiry
│   ├── cardutil/       # Card display helpers
│   ├── httputil/       # Health check handler
│   └── logger/         # Structured zap logger
├── ui/                 # React merchant dashboard (Vite + TypeScript)
├── go.mod
├── payments.db (not pushed)
├── idempotency.db (not pushed)
└── README.md
```

---

## Getting Started

### Prerequisites

- Go 1.26+
- Node.js 20+ and npm
- Arduino IDE with ESP32 board support
- Adafruit PN532 library
- ThingPulse SSD1306Wire library

### Run the backend

```bash
# Terminal 1 — Issuer
go run ./cmd/issuer

# Terminal 2 — Network router
go run ./cmd/network

# Terminal 3 — Acquirer
go run ./cmd/acquirer
```

### Run the dashboard UI

```bash
cd ui
npm install
npm run dev
```

The Vite dev server proxies `/api/*` to the issuer service at `http://localhost:8082`.

### Flash the ESP32

Open `esp32/terminal.ino` in Arduino IDE, set your WiFi credentials and acquirer IP, then upload to the board.

### Test without hardware

```bash
curl -X POST http://localhost:8080/authorize \
  -H "Content-Type: application/json" \
  -d '{"card_uid":"2E:F8:14:07","merchant_id":"M001","amount":24.99}'
```

---

## Simulated Accounts

| Card | Holder | Balance | Status |
|---|---|---|---|
| 2E:F8:14:07 | Hely Cimer | $500.00 | active |
| 1B:5E:32:07 | Jane Smith | $23.50 | active |
| BLOCKED:01 | Bob Block | $100.00 | blocked |

---

## Simulated Merchants

| Merchant ID | Name | Status |
|---|---|---|
| M001 | Kingly | active |
| M002 | Kinjo Sushi and Grill | active |
| M003 | Blocked Merchant | blocked |

---

## Tech Stack

| Layer | Technology |
|---|---|
| Terminal firmware | C++ (Arduino) |
| Backend services | Go 1.26 |
| Frontend dashboard | React 19 + TypeScript + Vite |
| Data fetching | TanStack Query |
| Database | SQLite (`modernc.org/sqlite`) |
| Logging | `go.uber.org/zap` |
| Transport | HTTP/JSON |
| Hardware | ESP32, PN532, SSD1306 |

---

## Why this project

This project simulates end-to-end payment authorization with realistic service boundaries and introduces a merchant-style operations dashboard for analytics. It demonstrates payment domain fundamentals, backend service design in Go, and frontend data visualization with React.

---

## Author

Built by Francisco — Calgary, AB  
Portfolio project