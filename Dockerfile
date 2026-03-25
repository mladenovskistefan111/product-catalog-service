# --- Stage 1: Build ---
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Download dependencies first (cached layer — only re-runs if go.mod/go.sum change)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a fully static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /product-catalog-service \
    ./cmd/server

# --- Stage 2: Run ---
# scratch = zero base image, nothing but your binary
# Safe because CGO_ENABLED=0 produces a fully static binary with no OS deps
FROM scratch

WORKDIR /app

# Copy the binary from the builder
COPY --from=builder /product-catalog-service .

# Single goroutine stack trace on panic (good for debugging in prod)
ENV GOTRACEBACK=single

# gRPC port
EXPOSE 3550

ENTRYPOINT ["/app/product-catalog-service"]