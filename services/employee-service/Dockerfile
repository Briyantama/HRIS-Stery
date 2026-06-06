# ============================================================
# Employee-Service Multi-Stage Dockerfile
# ============================================================
# Build:  docker build -f services/employee-service/Dockerfile -t hris-employee-service .
# Run:    docker run -e DATABASE_URL=... -e NATS_URL=... -p 50052:50052 hris-employee-service
# ============================================================

# ──────────────────────────────────────────────────────────
# Stage 1: Builder
# ──────────────────────────────────────────────────────────
FROM golang:1.25.0-alpine AS builder

WORKDIR /app

# Copy go.work and module files
COPY go.work go.work.sum ./
COPY services/employee-service/go.mod services/employee-service/go.sum ./services/employee-service/
COPY services/_shared/go.mod services/_shared/go.sum ./services/_shared/
COPY services/auth-service/go.mod services/auth-service/go.sum ./services/auth-service/
COPY gen/go/go.mod gen/go/go.sum ./gen/go/

# Download dependencies
RUN go mod download

# Copy source
COPY services/employee-service ./services/employee-service
COPY services/_shared ./services/_shared
COPY services/auth-service ./services/auth-service
COPY gen/go ./gen/go

# Build binary
RUN cd services/employee-service && \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/bin/employee-service \
    ./cmd/server

# ──────────────────────────────────────────────────────────
# Stage 2: Runtime
# ──────────────────────────────────────────────────────────
FROM alpine:3.20

# Install ca-certificates for TLS/HTTPS
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /app/bin/employee-service /usr/local/bin/employee-service

# Runtime configuration
EXPOSE 50052
ENTRYPOINT ["employee-service"]

# Default environment (override at runtime)
ENV GRPC_PORT=50052 \
    DATABASE_URL=postgres://hris_app:hris_app_secret@localhost:6432/hris_db?sslmode=disable \
    NATS_URL=nats://localhost:4222
