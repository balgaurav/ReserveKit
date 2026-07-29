# Architecture decisions

## Service responsibilities

`reservation-api` owns the browser-facing HTTP contract and the reservation lifecycle. `inventory` owns stock quantities and exposes a narrow gRPC API for availability and holds.

## Reservation flow

1. The client sends a product, quantity, expiry time, and idempotency key.
2. The reservation API validates the request and forwards a hold request over gRPC.
3. Inventory atomically decrements available stock only once for the supplied reservation ID.
4. The reservation API stores the lifecycle state and returns the result with the correlation ID.

The demo uses service-owned in-memory stores so it starts without cloud credentials or external data. The API and gRPC boundaries, validation, idempotency rules, and generated contracts are real. A production evolution would add separate Postgres databases, an outbox, a reconciliation worker, and authenticated service transport.

## Consistency model

- Inventory holds are atomic within the inventory service and idempotent by reservation ID.
- HTTP creation is idempotent by the `Idempotency-Key` header.
- Expired reservations release their hold when confirmation is attempted.
- Confirmation never changes inventory because the pending hold already owns the stock.
- Release is idempotent across API retries.
