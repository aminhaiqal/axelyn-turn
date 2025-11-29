# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Axelyn-turn is a modular, event-driven queue system designed for real-world businesses where multiple users, staff, and processes interact concurrently. Built with a microservice architecture focusing on speed, concurrency, and extendability.

## Build and Development Commands

### Queue-Core Service (Go)

The main service is `queue-core` written in Go. All commands should be run from `services/queue-core/`:

```bash
# Build the service
make build

# Run the service
make run

# Format code
make fmt

# Lint code (requires golangci-lint)
make lint

# Run all tests (unit, integration, load)
make test

# Run specific test types
make unit           # Unit tests only
make integration    # Integration tests only
make load          # Load tests only
```

### Running Tests

Tests are organized in `services/queue-core/tests/`:
- `tests/unit/` - Unit tests with mocked dependencies
- `tests/integration/` - Integration tests requiring real DB/Redis
- `tests/load/` - Load/performance tests

To run a single test file:
```bash
go test -v ./tests/unit/service_test.go
```

## Architecture

### Microservices Structure

The system is organized as separate services under `services/`:
- **queue-core**: Core queue management service (Go) - currently implemented
- **queue-worker**: Worker service for processing tickets - placeholder
- **queue-analytics**: Analytics and reporting service - placeholder
- **notifications**: Notification service - placeholder
- **frontend**: Web interface - placeholder

### Queue-Core Architecture

Queue-core is the primary service implementing the ticket queue system.

**Key Components:**

1. **HTTP API** (`internal/api/router.go`)
   - `POST /tickets` - Create new ticket
   - `GET /tickets/waiting?queue_id=N` - List waiting tickets for a queue
   - `GET /ws?queue_id=N` - WebSocket connection for real-time updates

2. **Ticket Service** (`internal/services/ticket_service.go`)
   - Creates tickets and publishes events to Redis Stream + PubSub
   - Runs a background Dispatcher goroutine that continuously reserves waiting tickets
   - The Dispatcher uses `ReserveNext()` to atomically transition tickets from "waiting" to "processing"

3. **Repository Layer** (`internal/repositories/ticket_repo.go`)
   - Database operations with optimistic locking via version field
   - `ReserveNext()` uses `FOR UPDATE SKIP LOCKED` for concurrent reservation
   - `UpdateStatus()` enforces optimistic locking with version checks

4. **Database** (`internal/db/`)
   - PostgreSQL connection via pgx driver (`postgres.go`)
   - Upstash Redis connection for streams and pub/sub (`upstash.go`)
   - Migrations in `internal/db/migrations/`

### Data Flow

1. **Ticket Creation**: HTTP POST → TicketService.CreateTicket() → DB insert → Redis Stream event → PubSub broadcast
2. **Ticket Reservation**: Dispatcher (background goroutine) → Repository.ReserveNext() → Status "waiting" → "processing" → Redis Stream event → PubSub broadcast
3. **Real-time Updates**: WebSocket clients subscribe to Redis PubSub channel `queue.<queue_id>.broadcast`

### Event Streaming

The system uses Redis for two purposes:
- **Redis Streams**: Durable event log for workers to consume (e.g., `queue.stream`, `queue.<id>.events`)
- **Redis PubSub**: Fast broadcast to WebSocket clients (e.g., `queue.<queue_id>.broadcast`)

Events published:
- `ticket.created` - New ticket added
- `ticket.reserved` - Ticket moved to processing
- Worker updates can be published back via `PublishWorkerUpdate()`

### Concurrency and Locking

- **Optimistic Locking**: All tickets have a `version` field incremented on each update
- **Database-level Locking**: `ReserveNext()` uses PostgreSQL's `FOR UPDATE SKIP LOCKED` to prevent race conditions
- **Status Transitions**: Updates validate current status and version to prevent conflicts

### Database Schema

**tickets table**:
- `id` (BIGSERIAL): Primary key
- `queue_id` (BIGINT): Which queue this ticket belongs to
- `customer_name` (TEXT): Customer identifier
- `status` (VARCHAR): "waiting", "processing", "done"
- `priority` (INT): Priority level (default 1)
- `estimated_time` (INT): Estimated processing time in seconds
- `version` (BIGINT): Optimistic locking version
- `created_at`, `updated_at` (TIMESTAMPTZ): Timestamps

**ticket_history table**: Archive table for completed tickets

Indexes optimize queries for `queue_id + status` combinations.

## Environment Variables

Queue-core requires:
- `DATABASE_URL`: PostgreSQL connection string
- `UPSTASH_REDIS_URL`: Redis connection URL (format: `rediss://user:token@host:port`)

Optional:
- Place in `.env` file in `services/queue-core/` for local development

## Testing Approach

Tests use:
- `testify/assert` for assertions
- `redismock` for mocking Redis operations
- `sqlmock` for mocking database operations

Example test structure can be found in `tests/unit/service_test.go`.

## Deployment

Deployment configurations in `deploy/`:
- `k8s/` - Kubernetes manifests
- `prod/` - Production deployment configs
- `staging/` - Staging deployment configs

Infrastructure in `infra/`:
- `docker-compose.yml` - Local development setup
- `traefik/` - Traefik reverse proxy configuration
