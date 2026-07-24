# ReserveKit

ReserveKit is a small, production-minded inventory reservation system for teams selling limited stock. A browser dashboard calls a Go reservation API; that API uses gRPC to reserve inventory from a separate inventory service.

## What it demonstrates

- A clear HTTP-to-gRPC service boundary using protobuf contracts
- Idempotent reservation creation and explicit `pending`, `confirmed`, `released`, and `expired` states
- PostgreSQL-backed inventory and reservation data
- Correlation IDs, structured logs, health checks, Docker Compose, and tests
- A concise React + TypeScript operations dashboard

## Architecture

```text
React dashboard -> Reservation API (HTTP :8080) -> Inventory service (gRPC :9090)
                                      |                     |
                                      +------ PostgreSQL ---+
```

## Planned local start

```bash
cp .env.example .env
docker compose up --build
```

The project is intentionally limited to two Go services and one web client so that the core distributed-systems choices remain easy to follow.
