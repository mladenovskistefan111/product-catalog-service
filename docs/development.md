# Develop & test

## Running locally

Requirements: Go 1.25+, Docker, a Postgres seeded from [`product-catalog-db`](https://github.com/platform-boutique/app-product-catalog-db), [`grpcurl`](https://github.com/fullstorydev/grpcurl).

```bash
# 1. database (see ../product-catalog-db)
docker run -d --name catalog-postgres -p 5432:5432 \
  -e POSTGRES_DB=productcatalog -e POSTGRES_USER=catalog -e POSTGRES_PASSWORD=catalog \
  mladenovskistefan/product-catalog-db:latest

# 2. service
export DATABASE_URL="postgres://catalog:catalog@localhost:5432/productcatalog?sslmode=disable"
go run ./cmd/server

# or via Docker
docker build -t product-catalog-service .
docker run --rm --network host -e DATABASE_URL="$DATABASE_URL" product-catalog-service

go test ./...
```

### Manual gRPC testing (reflection is on)

```bash
grpcurl -plaintext localhost:3550 hipstershop.ProductCatalogService/ListProducts
grpcurl -plaintext -d '{"id": "OLJCESPC7Z"}' localhost:3550 hipstershop.ProductCatalogService/GetProduct
grpcurl -plaintext -d '{"query": "kitchen"}'  localhost:3550 hipstershop.ProductCatalogService/SearchProducts
grpcurl -plaintext localhost:3550 grpc.health.v1.Health/Check
curl -s localhost:9090/metrics | grep grpc_server
```

Seed product IDs: `OLJCESPC7Z` (Sunglasses), `66VCHSJNUP` (Tank Top), `1YMWWN1N4O` (Watch), `L9ECAV7KIM` (Loafers), `2ZYFJ3GM2N` (Hairdryer), `0PUK6V6EV0` (Candle Holder), `LS4PSXUNUM` (Salt & Pepper Shakers), `9SIQT8TOJO` (Bamboo Glass Jar), `6E92ZMYYFZ` (Mug).

## Project structure

```
├── cmd/server/main.go            # entrypoint: tracing, profiling, signals (SIGUSR1/2), gRPC server, metrics server
├── internal/catalog/
│   ├── loader.go                 # pgx pool (InitDB) + LoadCatalog (SELECT … FROM products)
│   └── service.go                # gRPC handlers, in-memory cache, reload logic
├── proto/                        # productcatalog.proto + generated Go
├── Dockerfile
├── go.mod / go.sum
└── docs/                         # empty placeholder
```
