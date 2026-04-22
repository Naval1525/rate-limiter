# Distributed Rate Limiter

A horizontally-scaled, token-bucket rate limiter written in Go. Each API node is stateless; state lives in Redis, enforced atomically via a Lua script. Traffic is balanced across nodes by nginx. Metrics are scraped by Prometheus. Rate-limit decisions are streamed to Kafka for downstream analytics.

---

## Architecture

```
         ┌──────────┐
clients─▶│  nginx   │  (round-robin LB, :8080)
         └────┬─────┘
              │
     ┌────────┼────────┐
     ▼        ▼        ▼
   api1     api2     api3      (Go, :8080 internal)
     │        │        │
     ▼        ▼        ▼
  redis1   redis2   redis3     (per-node token-bucket state)
     │        │        │
     └────────┼────────┘
              ▼
           kafka                (rate-limit event stream)
              ▲
              │
         prometheus             (scrapes /metrics from each api, :9090)
```

### Why this layout

- **Sharded Redis, not clustered.** Each api owns a redis instance keyed by client — simpler ops, no cluster slot rebalancing. Trade-off: a client's bucket lives on one shard, so the nginx→api mapping is effectively stateless only because each client's quota is tracked per-node. If you need globally-consistent quotas across nodes, switch to a real Redis cluster and route all apis to it.
- **Lua on Redis, not Go-side logic.** `token_bucket.lua` runs atomically inside Redis, so two concurrent requests for the same key can't race the `read → decide → write` sequence.
- **Kafka for events, not for decisions.** The rate-limit verdict is returned synchronously from Redis; Kafka gets a fire-and-forget copy for audit, analytics, and alerting. If Kafka is down, requests still flow.

---

## Token Bucket Algorithm

Per client key, Redis stores a hash: `{ tokens, timestamp }`.

On each request:

1. Read `tokens`, `timestamp`.
2. Compute elapsed time since last refill; add `elapsed * refill_rate` tokens (capped at `capacity`).
3. If `tokens >= 1`: decrement, mark **allowed**. Else mark **blocked**.
4. Write back `{ tokens, now }`.

Entire sequence is one `EVAL` — atomic per key.

### Quotas

| API key header      | Capacity | Refill rate (tok/s) |
|---------------------|----------|---------------------|
| (none — by IP)      | 5        | 2                   |
| `X-API-Key: premium`| 20       | 10                  |

Defined in `internal/middleware/rate_limiter.go`.

---

## Endpoints

| Method | Path       | Description                              |
|--------|------------|------------------------------------------|
| GET    | `/hello`   | Protected route, returns `Hello 🚀`      |
| GET    | `/health`  | Liveness probe                           |
| GET    | `/metrics` | Prometheus exposition (Go runtime + custom counters) |

All requests under `/` go through the rate-limit middleware. `/metrics` is currently also wrapped — move it out of the middleware chain if you want Prometheus scrapes to bypass quotas.

---

## Running

```bash
docker compose up --build
```

Services:

| Service     | Host port | Purpose                   |
|-------------|-----------|---------------------------|
| nginx       | 8080      | Client-facing LB          |
| prometheus  | 9090      | Metrics UI                |
| redis1/2/3  | 6381–6383 | Per-node state            |
| kafka       | 9092      | Event stream              |

### Smoke test

```bash
# free tier — 5 tokens, refills 2/s
for i in {1..10}; do curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/hello; done

# premium — 20/10
for i in {1..25}; do curl -s -o /dev/null -w "%{http_code}\n" \
  -H "X-API-Key: premium" http://localhost:8080/hello; done
```

First N requests return `200`, the rest return `429 Too Many Requests`.

### Verify Prometheus

Visit `http://localhost:9090/targets` — all three api targets should show `up`. Then query:

```
requests_allowed_total
requests_blocked_total
rate(requests_blocked_total[1m])
```

---

## Project Layout

```
cmd/server/main.go              entry point; wires redis, lua, middleware, routes
internal/
  handler/                      route handlers (/hello, /health, test endpoints)
  middleware/
    rate_limiter.go             HTTP middleware — key extraction, Lua dispatch
    metrics.go                  Prometheus counters (allowed/blocked)
  repository/
    redis.go                    Redis client factory (reads REDIS_ADDR)
    lua.go                      loads + executes token_bucket.lua
    lua/token_bucket.lua        atomic bucket logic
nginx.conf                      upstream block for api1/api2/api3
prometheus.yml                  scrape config
docker-compose.yml              full stack
Dockerfile                      multi-stage Go build
```

---

## Configuration

Environment variables (read in `main.go` / `repository/redis.go`):

| Var            | Default       | Used by               |
|----------------|---------------|-----------------------|
| `REDIS_ADDR`   | `redis:6379`  | `NewRedisSimpleClient`|
| `KAFKA_BROKER` | `kafka:9092`  | Kafka producer (WIP)  |

Each api container gets its own `REDIS_ADDR` via `docker-compose.yml`.

---

## Metrics

Defined in `internal/middleware/metrics.go`:

- `requests_allowed_total` — counter, incremented before the Lua call
- `requests_blocked_total` — counter, incremented on `429`

**Heads up:** `InitMetrics()` is not called from `main.go` yet — register it on startup or these counters never appear in `/metrics`. Fix:

```go
// in main.go, before ListenAndServe
middleware.InitMetrics()
```

---

## Tech Stack

- Go 1.24
- `github.com/redis/go-redis/v9`
- `github.com/segmentio/kafka-go`
- `github.com/prometheus/client_golang`
- Redis 7, Kafka (Confluent), nginx, Prometheus
