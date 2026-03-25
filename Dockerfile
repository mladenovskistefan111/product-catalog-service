# --- Stage 1: Build ---
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /product-catalog-service \
    ./cmd/server

# --- Stage 2: Run ---
# Switch from scratch to alpine so pprof/pyroscope have what they need
FROM alpine:3.21

WORKDIR /app

COPY --from=builder /product-catalog-service .

ENV GOTRACEBACK=single

EXPOSE 3550 9090

ENTRYPOINT ["/app/product-catalog-service"]