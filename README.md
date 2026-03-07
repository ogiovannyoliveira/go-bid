# GoBid

A real-time online auction platform built with Go. Users can create auctions, place bids, and receive live updates through WebSocket connections.

---

## What it does

- **User accounts** – sign up, log in and log out with session-based authentication
- **Product auctions** – create auctions with a base price and an end time
- **Live bidding** – connect to an auction room via WebSocket and place bids in real time
- **Automatic auction expiry** – rooms close automatically when the configured end time is reached and all connected clients are notified

---

## Tech stack

| Tool | Purpose |
|---|---|
| [Go](https://go.dev/) | Language |
| [Chi](https://github.com/go-chi/chi) | HTTP router & middleware |
| [pgx](https://github.com/jackc/pgx) | PostgreSQL driver |
| [sqlc](https://sqlc.dev/) | SQL-to-Go code generation |
| [tern](https://github.com/jackc/tern) | Database migrations |
| [scs](https://github.com/alexedwards/scs) | Session management (PostgreSQL-backed) |
| [gorilla/websocket](https://github.com/gorilla/websocket) | WebSocket protocol |
| [gorilla/csrf](https://github.com/gorilla/csrf) | CSRF protection |
| [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) | Password hashing |
| [godotenv](https://github.com/joho/godotenv) | `.env` file loading |
| [air](https://github.com/air-verse/air) | Live reload for development |
| [Docker / Compose](https://docs.docker.com/compose/) | PostgreSQL container |

---

## Project structure

```
go-bid/
├── cmd/
│   ├── api/
│   │   └── main.go              # Server entry point (port 3080)
│   └── terndotenv/
│       └── main.go              # Migration runner (loads .env before tern)
├── internal/
│   ├── api/
│   │   ├── api.go               # Core Api struct (router, sessions, services)
│   │   ├── routes.go            # Route registration and middleware stack
│   │   ├── auth.go              # AuthMiddleware (session guard)
│   │   ├── user_handlers.go     # POST /signup, /login, /logout
│   │   ├── product_handlers.go  # POST /products
│   │   └── auction_handlers.go  # GET /products/ws/subscribe/{product_id}
│   ├── services/
│   │   ├── users_service.go     # Create user, authenticate
│   │   ├── products_service.go  # Create product, get by ID
│   │   ├── bids_service.go      # Validate and place bids
│   │   └── auction_services.go  # Auction lobby, rooms, WebSocket clients
│   ├── store/pgstore/
│   │   ├── db.go                # Database connection pool
│   │   ├── models.go            # sqlc-generated models
│   │   ├── users.sql.go         # User queries
│   │   ├── products.sql.go      # Product queries
│   │   ├── bids.sql.go          # Bid queries
│   │   ├── migrations/          # SQL migration files (tern)
│   │   └── sqlc.yml             # sqlc configuration
│   ├── use_case/
│   │   ├── user/                # Request structs and validation for user ops
│   │   └── product/             # Request structs and validation for product ops
│   ├── validator/
│   │   └── validator.go         # Field validators (email, length, blank)
│   └── jsonutils/
│       └── json_utils.go        # Generic JSON encode/decode helpers
├── compose.yml                  # PostgreSQL service
├── Makefile                     # Development commands
├── .env.example                 # Environment variable template
└── go.mod
```

---

## Getting started

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- [air](https://github.com/air-verse/air) – `go install github.com/air-verse/air@latest`
- [tern](https://github.com/jackc/tern) – `go install github.com/jackc/tern/v2@latest`
- [sqlc](https://sqlc.dev/) – `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

### 1. Configure environment

```sh
cp .env.example .env
# Edit .env with your database credentials
```

**.env variables:**

| Variable | Description |
|---|---|
| `ENV` | Environment (`local` disables CSRF) |
| `GOBID_DATABASE_HOST` | PostgreSQL host |
| `GOBID_DATABASE_PORT` | PostgreSQL port (default: `5432`) |
| `GOBID_DATABASE_NAME` | Database name |
| `GOBID_DATABASE_USER` | Database user |
| `GOBID_DATABASE_PASSWORD` | Database password |
| `GOBID_CSRF_KEY` | Secret key for CSRF protection |

### 2. Start the database

```sh
make compose-up
```

### 3. Run migrations

```sh
make migrations
```

### 4. Start the server

```sh
make run
```

The API will be available at `http://localhost:3080`.

---

## API reference

All routes are prefixed with `/api/v1`.

### Users

| Method | Route | Auth | Description |
|---|---|---|---|
| `POST` | `/users/signup` | No | Create a new account |
| `POST` | `/users/login` | No | Log in and start a session |
| `POST` | `/users/logout` | Yes | End the current session |

### Products / Auctions

| Method | Route | Auth | Description |
|---|---|---|---|
| `POST` | `/products` | Yes | Create a new product auction |
| `GET` | `/products/ws/subscribe/{product_id}` | Yes | WebSocket – join an auction room |

### WebSocket message types

After connecting to an auction room, communicate using JSON messages with a `kind` field:

| Kind | Direction | Description |
|---|---|---|
| `0` – `PlaceBid` | Client → Server | Place a bid: `{"kind":0,"amount":150.0,"user_id":"..."}` |
| `1` – `SuccessfullyPlacedBid` | Server → Client | Your bid was accepted |
| `2` – `FailedToPlaceBid` | Server → Client | Bid rejected (too low) |
| `3` – `InvalidJSON` | Server → Client | Malformed message received |
| `4` – `NewBidPlaced` | Server → All | Another user placed a bid |
| `5` – `AuctionFinished` | Server → All | Auction ended |

---

## Development commands

```sh
make compose-up              # Start PostgreSQL container
make compose-down            # Stop containers and remove volumes
make run                     # Run with hot reload (air)
make migrations              # Apply pending migrations
make migration-create name=X # Create a new migration file
make migration-rollback      # Roll back the last migration
make sqlc                    # Regenerate database code with sqlc
```
