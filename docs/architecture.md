# Architecture decisions

## Service responsibilities

`reservation-api` owns the browser-facing HTTP contract and the reservation lifecycle. `inventory` owns stock quantities and exposes a narrow gRPC API for availability and holds.

## Reservation flow

1. The client sends a product, quantity, expiry time, and idempotency key.
2. The reservation API validates the request and forwards a hold request over gRPC.
3. Inventory atomically decrements available stock only once for the supplied reservation ID.
4. The reservation API stores the lifecycle state and returns the result with the correlation ID.

The initial implementation favors an explicit, observable workflow over a hidden distributed transaction. A production evolution could add an outbox and a reconciliation worker.
