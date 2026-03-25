# product-catalog-service

A gRPC service that exposes a product catalog for the platform-demo e-commerce platform. It serves product data from a PostgreSQL database and is part of a broader microservices platform built with full observability, GitOps, and internal developer platform tooling.

## Overview

The service exposes three gRPC methods:

| Method | Description |
|---|---|
| `ListProducts` | Returns all products in the catalog |
| `GetProduct` | Returns a single product by ID |
| `SearchProducts` | Searches products by name or description |

**Port:** `3550` (gRPC)  
**Protocol:** gRPC  
**Language:** Go  
**Database:** PostgreSQL (via `product-catalog-db`)

## Requirements

- Go 1.25+
- Docker
- A running PostgreSQL instance (see [product-catalog-db](https://github.com/mladenovskistefan111/product-catalog-db))
- `grpcurl` for manual testing

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | Yes | Postgres connection string e.g. `postgres://user:pass@host:5432/db?sslmode=disable` |
| `PORT` | No | gRPC server port (default: `3550`) |
| `ENABLE_TRACING` | No | Set to `1` to enable OpenTelemetry tracing |
| `COLLECTOR_SERVICE_ADDR` | No | OTel collector address e.g. `otel-collector:4317` (required if tracing enabled) |
| `EXTRA_LATENCY` | No | Inject artificial latency e.g. `200ms` (for chaos/load testing) |

## Running Locally

### 1. Start the database

Follow the instructions in [product-catalog-db](https://github.com/mladenovskistefan111/product-catalog-db) to start a PostgreSQL container with the schema and seed data already baked in.

### 2. Build and run the service

```bash
go build ./...
DATABASE_URL="postgres://catalog:catalog@localhost:5432/productcatalog?sslmode=disable" go run ./cmd/server
```

### 3. Run with Docker

```bash
docker build -t product-catalog-service .

docker run -p 3550:3550 \
  --network host \
  -e DATABASE_URL="postgres://catalog:catalog@localhost:5432/productcatalog?sslmode=disable" \
  product-catalog-service
```

## Testing

### Unit tests

```bash
go test ./...
```

### Manual gRPC testing

Install `grpcurl` then:

```bash
# list all products
grpcurl -plaintext localhost:3550 hipstershop.ProductCatalogService/ListProducts

# get a product by ID
grpcurl -plaintext -d '{"id": "OLJCESPC7Z"}' localhost:3550 hipstershop.ProductCatalogService/GetProduct

# search products
grpcurl -plaintext -d '{"query": "kitchen"}' localhost:3550 hipstershop.ProductCatalogService/SearchProducts

# health check
grpcurl -plaintext localhost:3550 grpc.health.v1.Health/Check
```

## Project Structure

```
├── cmd/server/         # Binary entrypoint
├── internal/catalog/   # Business logic — gRPC handlers and DB loader
├── proto/              # Proto definition and generated gRPC code
├── docs/               # Architecture decisions, runbooks, service contract
├── Dockerfile
├── go.mod
└── go.sum
```

## Documentation

See [`docs/`](./docs) for:

- Service contract and data model
- Architecture decision records
- Database schema
- Observability (metrics, traces, logs)
- Runbook

## Part Of

This service is part of [platform-demo](https://github.com/mladenovskistefan111) — a full platform engineering project featuring microservices, observability (LGTM stack), GitOps (Argo CD), policy enforcement (Kyverno), infrastructure provisioning (Crossplane), and an internal developer portal (Backstage).