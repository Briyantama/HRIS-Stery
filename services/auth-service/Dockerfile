# Build from repository root:
#   docker build -f services/auth-service/Dockerfile -t hris-auth-service .

FROM golang:1.25.0-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.work go.work.sum ./
COPY gen/go gen/go
COPY services/_shared services/_shared
COPY services/auth-service services/auth-service

WORKDIR /src/services/auth-service

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /auth-service ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /auth-service .

EXPOSE 50051

ENTRYPOINT ["/app/auth-service"]
