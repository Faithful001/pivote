# Pivote

Pivote is a backend API for the Pivote voting platform. It was built to handle the full lifecycle of a voting event: creating the program, enrolling participants via tokenized invite links, collecting votes in real time, and automatically expiring programs when the deadline passes.

The name is a blend of "pivot" and "e" for election. Voting is fundamentally about pivoting, and a good election shifts the direction of things. The name felt like a natural fit.

---

## What It Does

At its core, Pivote lets administrators create voting programs, invite users into them, and collect votes against a set of candidates. A few things that shaped how the system works:

- **Programs have a deadline.** When an admin activates a program and sets a `voting_ends_at` time, the server begins streaming a live countdown to every connected client using Server-Sent Events. When time runs out, the program is automatically expired by a background broadcaster process.

- **Votes propagate in real time.** When a user casts or toggles a vote, the result is broadcast to all connected clients over Socket.IO so that leaderboard views stay current without polling.

- **Invitations are token-based.** Users join a program by requesting an email invite. The link they receive contains a short-lived access token. Submitting that token through the join endpoint enrolls them in the program.

- **Email delivery is decoupled.** Transactional and notification emails go through a RabbitMQ queue and are processed by background consumer workers. This keeps the HTTP request cycle fast and makes the email pipeline independently scalable.

- **OTP-based authentication.** Registration and login flows use a one-time password delivered by email. A background cleanup worker periodically purges expired OTP records.

---

## Tech Stack

| Concern                 | Technology               |
| ----------------------- | ------------------------ |
| Language                | Go 1.24                  |
| HTTP framework          | Gin                      |
| Database                | PostgreSQL 15 via GORM   |
| Caching / rate limiting | Redis 7                  |
| Message broker          | RabbitMQ 3               |
| Real-time (WebSocket)   | Socket.IO (go-socket.io) |
| Real-time (SSE)         | Native SSE via Gin       |
| Email delivery          | Resend                   |
| Auth                    | JWT (golang-jwt/jwt v5)  |
| Dev hot-reload          | Air                      |
| Containerization        | Docker Compose           |

---

## Project Structure

```
pivote/
├── cmd/
│   └── api/
│       └── main.go           # Application entry point
├── internal/
│   ├── domains/              # Business logic organized by domain
│   │   ├── auth/
│   │   ├── candidate/
│   │   ├── otp/
│   │   ├── program/
│   │   ├── user/
│   │   ├── vote/
│   │   └── workspace/
│   ├── infra/                # Infrastructure adapters
│   │   ├── db/               # PostgreSQL connection and migrations
│   │   ├── rabbitmq/         # RabbitMQ client and consumer helpers
│   │   ├── redis/            # Redis client
│   │   ├── sse/              # SSE broadcaster manager
│   │   └── websocket/        # Socket.IO server
│   ├── middlewares/          # Auth, CORS, rate limiting
│   ├── router/               # Route registration
│   ├── types/                # Shared types
│   ├── utils/                # Utility helpers
│   └── workers/              # Background workers
│       ├── email/            # RabbitMQ email consumers
│       └── otp/              # OTP cleanup worker
├── docker-compose.yml
├── .env.example
└── .air.toml
```

---

## API Overview

All routes are prefixed with `/api/v1`. Role-based middleware guards each group: `admin` for management operations, `user` for participant-facing actions, and both for read operations that either party may need.

### Auth

| Method | Path             | Description                    |
| ------ | ---------------- | ------------------------------ |
| POST   | `/auth/register` | Register a new user            |
| POST   | `/auth/login`    | Authenticate and receive a JWT |

### OTPs

| Method | Path           | Description                |
| ------ | -------------- | -------------------------- |
| POST   | `/otps/verify` | Verify a one-time password |

### Users

| Method | Path         | Access      |
| ------ | ------------ | ----------- |
| GET    | `/users`     | Admin, User |
| GET    | `/users/:id` | Admin, User |

### Programs

| Method | Path                         | Access      | Description                             |
| ------ | ---------------------------- | ----------- | --------------------------------------- |
| POST   | `/programs`                  | Admin       | Create a voting program                 |
| GET    | `/programs`                  | Admin, User | List programs (filterable by workspace) |
| GET    | `/programs/:id`              | Admin, User | Get a single program                    |
| PUT    | `/programs/:id`              | Admin       | Update program details                  |
| DELETE | `/programs/:id`              | Admin       | Delete a program                        |
| PATCH  | `/programs/:id/toggle`       | Admin       | Activate or deactivate voting           |
| POST   | `/programs/:id/request-join` | Public      | Request an email invite                 |
| POST   | `/programs/:id/join`         | Public      | Join with a token from the invite email |
| GET    | `/programs/:id/countdown`    | Admin, User | SSE stream of the countdown timer       |

### Candidates

| Method | Path              | Access      |
| ------ | ----------------- | ----------- |
| POST   | `/candidates`     | Admin       |
| GET    | `/candidates`     | Admin, User |
| GET    | `/candidates/:id` | Admin, User |
| PUT    | `/candidates/:id` | Admin       |
| DELETE | `/candidates/:id` | Admin       |

### Votes

| Method | Path                             | Access      | Description                           |
| ------ | -------------------------------- | ----------- | ------------------------------------- |
| POST   | `/votes/toggle`                  | User        | Cast or retract a vote (rate limited) |
| GET    | `/votes/program/:program_id`     | Admin, User | Get votes for a program               |
| GET    | `/votes/candidate/:candidate_id` | Admin, User | Get votes for a candidate             |

### Workspaces

Workspaces group programs under an organizational boundary. Routes are registered under `/api/v1/workspaces`.

### WebSocket

Socket.IO is served at `/socket.io/*any` over both GET and POST. Vote events are emitted here as they happen.

### Health Check

```
GET /health
```

Returns the server version and a status confirmation.

---

## Getting Started

### Prerequisites

- Go 1.24 or later
- Docker and Docker Compose
- A [Resend](https://resend.com) API key for email delivery

### 1. Clone the repository

```bash
git clone https://github.com/your-username/pivote.git
cd pivote
```

### 2. Start infrastructure services

The `docker-compose.yml` file defines PostgreSQL, RabbitMQ, and Redis. Start all three with:

```bash
docker compose up -d
```

RabbitMQ's management UI will be available at `http://localhost:15672`.

### 3. Configure environment variables

Copy the example file and fill in your values:

```bash
cp .env.example .env
```

```env
ENV=development
POSTGRES_USER=your-db-user
POSTGRES_PASSWORD=your-db-pwd
POSTGRES_DB=your-db-name
DATABASE_URL=postgres://your-db-user:your-db-pwd@127.0.0.1:5435/pivotedb?sslmode=disable
REDIS_URL=redis://localhost:6379
RABBITMQ_URL=amqp://your-rabbitmq-user:your-rabbitmq-pwd@127.0.0.1:5672/
RABBITMQ_DEFAULT_USER=your-rabbitmq-user
RABBITMQ_DEFAULT_PASS=your-rabbitmq-pwd
JWT_SECRET=your-secret-key
RESEND_API_KEY=your-resend-api-key
```

### 4. Run the server

**With hot-reload (recommended for development):**

```bash
air
```

**Without hot-reload:**

```bash
go run ./cmd/api/main.go
```

The server starts on port `8000`. You can verify it is running by hitting the health endpoint:

```bash
curl http://localhost:8000/health
```

---

## How the Real-Time Layer Works

There are two real-time channels and they serve different purposes.

**SSE countdown stream:** When a client connects to `GET /programs/:id/countdown`, the server upgrades the connection to a persistent SSE stream. The `BroadcasterManager` maintains an in-memory map of active programs to subscriber channels. Every active program gets a goroutine that ticks every second and pushes the remaining time to all registered subscriber channels. When the deadline passes, the broadcaster automatically triggers program expiration and closes all subscriber channels, which causes the HTTP handlers to flush a final `countdown: 0` event and terminate cleanly.

Program state changes (activation, deadline updates) are propagated across the system via a RabbitMQ fanout exchange named `program.events`. Each running instance binds an exclusive, auto-deleted queue to this exchange on startup and listens in a background goroutine. This means in a multi-instance deployment, all instances stay in sync with program state without any shared in-memory state.

**Socket.IO vote events:** When a vote is cast or toggled, the vote controller emits an event over the Socket.IO server. Any client subscribed to the relevant room receives the update immediately, which is how live leaderboards stay current.

---

## Rate Limiting

Rate limiting is applied at the middleware level using Redis. The two main surfaces that are protected:

- Vote toggle: 20 requests per minute per user
- Program join: 10 attempts per 15 minutes per IP
- Join link request: 5 attempts per 15 minutes per IP
- Countdown stream: 30 connections per minute per user

---

## Contributing

Contributions are welcome. If you find a bug or want to propose a change, please open an issue first so we can discuss the approach before any code is written. Pull requests without an associated issue may be closed without review.

When submitting a pull request:

- Keep commits focused. One logical change per commit is easier to review and revert if needed.
- Match the existing code style. There is no linter configuration checked in yet, but the project follows standard Go conventions.
- Write a clear PR description explaining what changed and why.

---

## License

Pivote is released under the MIT License. See the [LICENSE](./LICENSE) file for details.
