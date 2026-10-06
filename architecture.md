# Architecture

This document describes the target architecture of the Go gRPC GraphQL Microservices platform and how far each part has come. The diagrams use Mermaid, which GitHub renders directly.

## Legend

| Colour | Meaning |
|---|---|
| 🟩 Green | Implemented in the repository today |
| 🟦 Blue | Being built in October 2026 (core scope) |
| 🟨 Amber | October stretch goal |
| ⬜ Grey (dashed) | Planned for later (November onward) |

Update the colours in the diagram as each part ships.

## 1. System overview

```mermaid
flowchart TD
  Web["Next.js storefront<br/>search + AI chat"]
  Gateway["GraphQL Gateway"]
  AI["ai-service<br/>FastAPI + LangGraph"]
  Account["Account Service"]
  Catalog["Catalog Service"]
  Order["Order Service"]
  PGA[("Postgres: accounts")]
  PGO[("Postgres: orders")]
  Redis[("Redis<br/>catalog cache + chat memory")]
  ES[("Elasticsearch<br/>products index")]
  VEC[("Elasticsearch<br/>ai_embeddings index")]
  Kafka{{"Kafka<br/>order.placed"}}
  AISum["AI summary consumer<br/>(Python)"]
  Inv["Inventory Consumer"]
  Notif["Notification Consumer"]
  RMQ{{"RabbitMQ<br/>send-email queue"}}
  Email["Email Worker"]
  SMTP["Gmail / SMTP"]
  Prom[("Prometheus")]
  Graf["Grafana Dashboards"]

  Web -->|GraphQL| Gateway
  Web -->|"chat (SSE)"| AI

  Gateway -->|gRPC| Account
  Gateway -->|gRPC| Catalog
  Gateway -->|gRPC| Order

  Order -->|gRPC verify user| Account
  Order -->|gRPC verify product| Catalog

  Account --> PGA
  Order --> PGO
  Catalog --> Redis
  Redis -.->|on cache miss| ES

  AI -->|"tools + product sync (GraphQL)"| Gateway
  AI -->|hybrid search| VEC
  AI -->|"chat memory, LLM cache"| Redis

  Order -->|"outbox, then publish"| Kafka
  Kafka --> AISum
  Kafka --> Inv
  Kafka --> Notif
  Inv -->|update stock| ES
  Notif -->|enqueue task| RMQ
  RMQ --> Email
  Email -->|SMTP| SMTP

  Gateway & Account & Catalog & Order & AI -.->|/metrics| Prom
  Prom --> Graf

  classDef done fill:#1f6f43,stroke:#2ea043,color:#ffffff;
  classDef core fill:#0b4f8a,stroke:#388bfd,color:#ffffff;
  classDef stretch fill:#8a6100,stroke:#d29922,color:#ffffff;
  classDef later fill:#444c56,stroke:#768390,color:#ffffff,stroke-dasharray: 4 3;

  class Gateway,Account,Catalog,Order,PGA,PGO,ES done;
  class Web,AI,VEC,Redis,Prom,Graf core;
  class Kafka,AISum stretch;
  class Inv,Notif,RMQ,Email,SMTP later;
```

## 2. What changes compared to the base repository

| Area | In the repo today | October change | How it works | Target date |
|---|---|---|---|---|
| Elasticsearch + Compose | Client and image versions can mismatch | Fix the compatibility, verify the full stack end to end | Catalog talks to Elasticsearch through the Go client, so versions must match | Oct 7 |
| Health and lifecycle | No health checks or graceful shutdown | gRPC health + readiness endpoints, Compose healthchecks, SIGTERM handling | Containers only receive traffic when ready, and finish in-flight requests when stopped | Oct 8 |
| Errors | Mixed error handling | Consistent gRPC status codes (NotFound, InvalidArgument, Unavailable, Internal) | Gateway can map each code to a clean GraphQL error | Oct 9 |
| CI | None | GitHub Actions: format, vet, test, build images, push to ECR | Every push proves the stack still builds and passes tests | Oct 9 |
| Redis | Only in the diagram | Cache-aside in the catalog service | Read: Redis hit returns fast, miss reads Elasticsearch then fills Redis with a TTL. Write: invalidate the key | Oct 10-12 |
| AWS deploy | Local only | ECR + EC2 running Docker Compose, security groups | CI pushes images, the instance pulls and runs them. Only 80/443 are public | Oct 13-14 |
| Observability | Planned | `/metrics` on every service, Prometheus, Grafana | Latency, errors, cache hit rate on one dashboard | Oct 15 |
| Search | Basic name/description search | Fuzzy match, filters, aggregations, GraphQL `search` query | Powers the storefront search page | Oct 16 |
| ai-service (new) | Does not exist | Python FastAPI + LangGraph agent + RAG | See section 5 | Oct 12-26 |
| Storefront (new) | Does not exist | Next.js: search page + streaming AI chat | Calls the gateway (GraphQL) and ai-service (streaming) | Oct 17-23 |
| HTTPS | None | Caddy reverse proxy + domain | Terminates TLS in front of the gateway and ai-service | Oct 20 |
| Order status | Not stored | Status field on orders (PENDING / CONFIRMED / FAILED / CANCELLED) | Lets the agent and clients report order state | Oct 21 |
| Outbox + Kafka | Planned | Outbox table, publisher, `order.placed` topic, a logging consumer | Order and event are saved in one transaction, so no event is lost | Oct 22-24 (stretch) |
| Stock reservation, inventory consumer, RabbitMQ, email worker, gRPC auth/TLS, tracing | Planned | Not in October | Continue in November | Later |

## 3. Reading data (catalog with Redis)

```mermaid
sequenceDiagram
  participant GW as Gateway
  participant CAT as Catalog Service
  participant R as Redis
  participant ES as Elasticsearch

  GW->>CAT: gRPC GetProduct or ListProducts
  CAT->>R: lookup key
  alt cache hit
    R-->>CAT: cached value
  else cache miss
    CAT->>ES: query products index
    ES-->>CAT: result
    CAT->>R: store with TTL
  end
  CAT-->>GW: response
```

## 4. Creating an order

```mermaid
sequenceDiagram
  autonumber
  actor C as Client
  participant GW as GraphQL Gateway
  participant ORD as Order Service
  participant ACC as Account Service
  participant CAT as Catalog Service
  participant PG as Postgres orders
  participant K as Kafka

  C->>GW: createOrder mutation
  GW->>ORD: gRPC CreateOrder
  ORD->>ACC: gRPC verify account
  ORD->>CAT: gRPC get products and prices
  ORD->>ORD: validate quantities, calculate total, snapshot prices
  ORD->>PG: one transaction saves order and outbox row
  ORD-->>GW: order with status
  GW-->>C: GraphQL response
  Note over ORD,K: Stretch goal. After commit, the outbox publisher sends order.placed
  ORD->>K: order.placed (async)
```

The synchronous part answers the caller immediately. Side effects (stock, email, AI summary) happen asynchronously from the event.

## 5. AI chat flow (ai-service)

```mermaid
sequenceDiagram
  autonumber
  actor U as User
  participant W as Next.js storefront
  participant AI as ai-service
  participant R as Redis
  participant L as LLM API
  participant V as ES ai_embeddings
  participant GW as GraphQL Gateway

  U->>W: Find a phone under 20k with good battery
  W->>AI: POST /chat (streamed)
  AI->>R: load conversation memory
  AI->>L: prompt and available tools
  L-->>AI: call search_products
  AI->>V: hybrid search (keywords and vectors)
  V-->>AI: candidate product IDs
  AI->>GW: GraphQL products by ID for live price and stock
  GW-->>AI: product details
  AI->>L: tool results
  L-->>AI: answer with product IDs
  AI-->>W: stream answer
  W-->>U: recommendations
  U->>W: Order the second one
  W->>AI: POST /chat
  AI->>L: prompt and available tools
  L-->>AI: call create_order
  AI-->>W: pause and ask user to confirm
  U->>W: Confirm
  W->>AI: confirm
  AI->>GW: createOrder mutation
  GW-->>AI: order result
  AI-->>W: confirmation message
```

How it works in plain words:

1. **RAG:** Product text is turned into embeddings and stored in the ai-service's own index. A question retrieves the closest products, so the model answers from real catalog data.
2. **Agent tools:** The LLM does not touch any database. It asks the ai-service to run tools (search products, get product, get order, create order). Each tool is a normal GraphQL call to the gateway.
3. **Human in the loop:** Any action that changes data (creating an order) pauses until the user confirms in the UI.
4. **Evals:** A fixed set of test conversations checks that the agent picks the right tool and answers from the catalog. The pass rate is tracked over time.

## 6. Deployment

```mermaid
flowchart LR
  Dev["git push"] --> GA["GitHub Actions<br/>test + build images"]
  GA --> ECR[("AWS ECR<br/>image registry")]
  GA -->|deploy| EC2
  User["User"] --> Vercel["Vercel<br/>Next.js storefront"]
  Vercel -->|HTTPS| Caddy
  ECR -->|pull images| EC2
  TF["Terraform (stretch)"] -.->|creates| EC2
  TF -.-> ECR

  subgraph EC2["AWS EC2 (docker compose)"]
    Caddy["Caddy<br/>HTTPS reverse proxy"]
    GW2["GraphQL Gateway"]
    AI2["ai-service"]
    SVC["Account, Catalog, Order"]
    INFRA["Postgres x2, Redis, Elasticsearch<br/>Prometheus, Grafana"]
    Caddy --> GW2
    Caddy --> AI2
    AI2 --> GW2
    GW2 --> SVC
    SVC --> INFRA
  end
```

Network rules: only ports 80 and 443 are public. SSH (22) is limited to your own IP. Databases, Redis, Elasticsearch, Prometheus and Grafana are never exposed publicly.

## 7. Design rules (extended)

The existing rules still apply. Two are added for the AI service:

- **Database ownership:** a service never reads another service's database directly.
- **Explicit contracts:** gRPC protobuf definitions are the boundary between services.
- **Thin gateway:** GraphQL translates and composes; domain rules stay inside services.
- **Synchronous core, asynchronous side effects:** order validation and persistence happen in the request; stock, email and AI summaries react to events.
- **ai-service is just another client:** it reads product and order data only through the gateway, never from the catalog or order databases.
- **ai-service owns its own index:** the embeddings live in an `ai_embeddings` index owned by the ai-service, filled from the catalog through the gateway.
- **Observable behaviour:** every service exposes `/metrics`.
