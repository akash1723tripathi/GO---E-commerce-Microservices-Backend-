# Backend Implementation Status

Last reviewed: 2026-09-01

## Project goal

This repository is building an e-commerce backend using independently deployable Go microservices. GraphQL is the public API, and the services communicate internally through gRPC.

The intended backend contains:

```text
Client
  |
  v
GraphQL Gateway
  |
  +--> Account Service ----> PostgreSQL
  +--> Catalog Service ----> Search/Product Store
  +--> Order Service ------> PostgreSQL
                              |
                              +--> Kafka: order.placed
                                      |
                                      +--> Inventory Consumer
                                      +--> Notification Consumer --> RabbitMQ --> Email Worker --> SMTP
```

## Implemented

### Repository and contracts

- Account, catalog, order, and GraphQL gateway directories exist as separate runtime boundaries.
- Go module and dependency configuration are present.
- Account and catalog protobuf contracts exist.
- Order protobuf contract was added.
- Generated Go gRPC/protobuf code is checked in for the services.

### Account service

- Account creation through the service layer.
- Account lookup by ID.
- Paginated account listing.
- PostgreSQL repository.
- PostgreSQL schema migration in `account/up.sql`.
- gRPC server and client.
- Reflection-enabled gRPC server.
- Container build configuration.

### Catalog service

- Product creation.
- Product lookup by ID.
- Product listing with pagination.
- Product lookup by IDs.
- Product search by name and description.
- Elasticsearch repository.
- gRPC server and client.
- Reflection-enabled gRPC server.
- Container build configuration.

### Order service

- Order creation through the service layer.
- Account verification through the account service.
- Product verification through the catalog service.
- Quantity validation.
- Duplicate-product validation.
- Order total calculation using catalog prices.
- Product price and description snapshots stored with each order.
- Order lookup by ID.
- Order history lookup by account ID.
- Pagination for order history.
- PostgreSQL repository using orders and order-products tables.
- Transactional persistence of an order and its line items.
- PostgreSQL schema migration in `order/up.sql`.
- gRPC server and client.
- Reflection-enabled gRPC server.
- Unit tests for order validation, total calculation, product snapshots, and persistence calls.
- Container build configuration.

### GraphQL gateway

- GraphQL schema includes accounts, products, orders, pagination, and mutations.
- Account, catalog, and order gRPC clients are connected by the gateway.
- Account queries are implemented.
- Product queries are implemented.
- Account creation mutation is implemented.
- Product creation mutation is implemented.
- Order creation mutation is implemented.
- Account order lookup is implemented through the order service.
- Request timeouts are applied to mutations.
- GraphQL playground and `/graphql` HTTP endpoints are configured.

### Local deployment

- Docker Compose definitions exist for:
  - GraphQL gateway
  - Account service
  - Catalog service
  - Order service
  - Account PostgreSQL
  - Order PostgreSQL
  - Elasticsearch
- Account and order SQL migrations are mounted into their PostgreSQL containers.
- Service-to-service URLs are configured through environment variables.
- Docker Compose configuration validates successfully.

### Verification

The current Go test command succeeds:

```bash
go test ./...
```

The Compose configuration also validates with:

```bash
docker compose config
```

## Partially implemented or requiring correction

These areas exist but are not production-ready:

- Catalog persistence uses Elasticsearch directly. A durable source of truth for product stock is still required.
- The catalog repository uses older Elasticsearch typed APIs and must be checked against the Elasticsearch image version used in Compose.
- Docker Compose has startup dependencies but does not yet have proper readiness health checks.
- gRPC clients currently use insecure transport.
- Services do not yet expose standard liveness and readiness endpoints.
- Error responses are not consistently mapped to gRPC status codes.
- Services do not yet implement graceful shutdown.
- There are no end-to-end tests running against real PostgreSQL and Elasticsearch containers.
- Database migrations are mounted for local development but there is no dedicated migration versioning process.
- Production secrets, TLS, backups, resource limits, and observability are not configured.

## Required remaining backend work

### 1. Add order state and lifecycle

Add an order status such as:

```text
PENDING
CONFIRMED
FAILED
CANCELLED
```

The order should initially be created as `PENDING`. It should become `CONFIRMED` only after inventory is successfully reserved.

### 2. Add reliable order events

Add an outbox table to the order database. Order creation should persist the order and an `order.placed` event in the same PostgreSQL transaction.

Add an outbox publisher that sends unpublished events to Kafka. This prevents the order database and Kafka from becoming inconsistent if one operation succeeds and the other fails.

### 3. Add Kafka

Add Kafka configuration and a topic such as:

```text
order.placed
```

The event should include the order ID, account ID, product IDs, quantities, total, and creation time. Consumers must be idempotent so that retrying the same event does not reserve stock twice or send duplicate notifications.

### 4. Add stock ownership and reservation

The catalog domain needs stock data and an atomic stock-reservation operation. Add a gRPC method such as `ReserveStock`.

The inventory consumer should:

1. Consume `order.placed`.
2. Reserve stock atomically.
3. Retry temporary failures.
4. Mark the order confirmed or failed.
5. Avoid processing the same event more than once.

Elasticsearch should be treated as a search projection, not the authoritative store for stock quantities.

### 5. Add notification processing

Add a notification consumer that reads order events from Kafka and publishes email jobs to RabbitMQ.

Add an email worker that:

- Consumes RabbitMQ jobs.
- Sends email through SMTP.
- Acknowledges messages only after successful delivery.
- Retries temporary SMTP failures.
- Moves permanently failing messages to a dead-letter queue.

### 6. Add Redis catalog caching

Add Redis for frequently requested product and product-list reads.

Required behavior:

- Read from Redis first.
- Read from the catalog store on cache miss.
- Populate Redis after a successful store read.
- Invalidate or refresh cache entries after product or stock changes.
- Use short expiration times for stock-related data.

### 7. Add service reliability features

Every service should add:

- gRPC health checks.
- Readiness checks for dependencies.
- Graceful shutdown on `SIGTERM`.
- Request IDs and structured logging.
- Timeouts on outbound calls.
- Bounded retries for safe operations.
- Circuit breaking for unavailable dependencies.
- Correct gRPC status codes such as `NotFound`, `InvalidArgument`, `Unavailable`, and `Internal`.

### 8. Add security

- Replace insecure gRPC transport with TLS or a private trusted network.
- Add authentication at the GraphQL boundary.
- Add authorization checks for account and order access.
- Store secrets outside Git and Compose source files.
- Restrict public access to PostgreSQL, Elasticsearch, Kafka, and RabbitMQ.
- Add HTTPS through a reverse proxy.

### 9. Add observability

Expose Prometheus metrics for:

- Request count and latency.
- gRPC errors.
- Database errors.
- Catalog cache hit/miss rates.
- Kafka consumer lag.
- RabbitMQ queue depth.
- Email delivery failures.

Add Grafana dashboards and distributed tracing so a GraphQL request can be followed across services.

### 10. Add automated testing and delivery

Add:

- Repository tests using a test database.
- gRPC server/client integration tests.
- GraphQL API tests.
- Kafka consumer tests.
- RabbitMQ worker tests.
- End-to-end order-flow tests.
- CI checks for formatting, tests, protobuf generation, and Docker builds.
- Image publishing and deployment automation.

## Recommended implementation sequence

1. Fix and standardize Elasticsearch compatibility.
2. Add order status and status update operations.
3. Add the PostgreSQL outbox table and publisher.
4. Add Kafka and the `order.placed` event.
5. Add authoritative product stock and `ReserveStock`.
6. Add the inventory consumer.
7. Add order confirmation/failure handling.
8. Add Kafka-to-RabbitMQ notification processing.
9. Add the SMTP email worker.
10. Add Redis caching.
11. Add health checks, graceful shutdown, retries, metrics, and tracing.
12. Add security, backups, CI/CD, and production deployment configuration.

## Current completion estimate

The synchronous foundation is implemented:

```text
GraphQL -> gRPC -> Account/Catalog/Order -> PostgreSQL/Search Store
```

The asynchronous and production platform is still pending:

```text
Order -> Outbox -> Kafka -> Inventory/Notifications -> RabbitMQ -> Email
```

The next highest-value implementation is the transactional outbox plus Kafka integration, followed by stock reservation.
