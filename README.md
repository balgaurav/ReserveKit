# ReserveKit

ReserveKit is a compact inventory-reservation platform that demonstrates a real HTTP-to-gRPC service boundary without burying the core workflow under infrastructure. Operators can inspect stock, create a time-limited hold, and confirm or release it from a responsive React dashboard.

## Why this project exists

Limited-stock systems must prevent duplicate holds, handle client retries safely, and make state transitions understandable. ReserveKit focuses on those rules:

- idempotent HTTP creation through `Idempotency-Key`
- atomic, retry-safe inventory holds in a dedicated Go service
- generated protobuf contracts and grpc-go client/server code
- explicit `pending`, `confirmed`, `released`, and `expired` states
- correlation IDs and useful HTTP/gRPC failure mapping
- integration coverage across the complete HTTP → gRPC path

## Architecture

```text
Browser
  │ HTTP /api
  ▼
React dashboard :5173 ──► Reservation API :8080
                                  │
                                  │ protobuf / gRPC
                                  ▼
                         Inventory service :9090
```

Both services use service-owned in-memory stores for a deterministic one-command demo. The boundaries are deliberately shaped so Postgres adapters can replace them without changing the public contracts. See [architecture decisions](docs/architecture.md) for consistency rules and production tradeoffs.

## Run the complete stack

```bash
docker compose up --build
```

Open `http://localhost:5173`. The dashboard includes seeded inventory for keyboards, monitors, and webcams.

### Run without Docker

```bash
# terminal 1
go run ./cmd/inventory

# terminal 2
go run ./cmd/reservation-api

# terminal 3
cd web
npm install
npm run dev
```

## API examples

```bash
curl 'http://localhost:8080/api/inventory?sku=keyboard'

curl -X POST 'http://localhost:8080/api/reservations' \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: checkout-42' \
  -d '{"sku":"keyboard","quantity":2,"ttlSeconds":900}'

curl -X POST 'http://localhost:8080/api/reservations/RESERVATION_ID/confirm'
curl -X POST 'http://localhost:8080/api/reservations/RESERVATION_ID/release'
```

## Repository map

```text
cmd/inventory/             gRPC service entry point
cmd/reservation-api/       HTTP service entry point
internal/inventory/        atomic stock and hold rules
internal/reservation/      lifecycle, HTTP handlers, gRPC client
proto/                     source protobuf contract
gen/                       generated Go protobuf code
web/                       React + TypeScript operations UI
deploy/                    production-oriented container builds
```

## Verification

```bash
go test ./...
cd web && npm ci && npm run build
```

GitHub Actions executes both checks on every push and pull request.

## Production evolution

The next production steps are separate Postgres databases per service, a transactional outbox and reconciliation worker, OpenTelemetry traces, mTLS between services, and authenticated user-facing APIs. Those are documented as extensions rather than implied features of this self-contained demo.
