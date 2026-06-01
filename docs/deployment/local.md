# Local Development Deployment

## Prerequisites

- Docker & Docker Compose
- Go 1.24+
- PostgreSQL 16+ (local or via Docker)
- NATS 2.10+ (local or via Docker)
- grpcurl (for testing)

## Quick Start (Docker Compose)

```bash
# 1. Copy environment template
cp .env.example .env

# 2. Update .env with local values (optional - defaults are fine for local)
nano .env

# 3. Build and start all services
docker-compose -f docker-compose.production.yml build
docker-compose -f docker-compose.production.yml up -d

# 4. Verify services are healthy
sleep 5
docker-compose -f docker-compose.production.yml ps

# 5. Check service health
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
grpcurl -plaintext localhost:50052 grpc.health.v1.Health/Check
```

## Manual Setup (No Docker)

```bash
# 1. Start PostgreSQL
brew services start postgresql@16  # macOS
# or
sudo systemctl start postgresql    # Linux

# 2. Create databases
createdb -U postgres hris
psql -U postgres hris << EOF
  CREATE SCHEMA auth;
  CREATE SCHEMA employee;
  CREATE SCHEMA attendance;
  CREATE SCHEMA leave;
  CREATE SCHEMA notification;
EOF

# 3. Run migrations per service
cd services/auth-service
migrate -path migrations/ -database "postgres://user:pass@localhost/hris" up
cd ../..

# Repeat for each service...

# 4. Start NATS (with JetStream)
nats-server -js

# 5. Start Redis
redis-server

# 6. Build services
go build ./services/auth-service/cmd/server/
go build ./services/employee-service/cmd/server/
# ... repeat for other services

# 7. Start services in separate terminals
export DATABASE_URL=postgres://user:pass@localhost/hris/auth
export GRPC_PORT=50051
export NATS_URL=nats://localhost:4222
./auth-service

# Repeat in separate terminals for other services...
```

## Testing Services Locally

### Test Auth Service Login

```bash
grpcurl -plaintext \
  -d '{"email":"admin@example.com","password":"password123"}' \
  localhost:50051 \
  hris.auth.v1.AuthService/Login
```

### Test Employee Service Get Employee

```bash
grpcurl -plaintext \
  -d '{"tenant_id":"<tenant-uuid>","employee_id":"<emp-uuid>"}' \
  localhost:50052 \
  hris.workforce.v1.EmployeeService/GetEmployee
```

### Test Notification Service (SMTP)

```bash
# Send notification
grpcurl -plaintext \
  -d '{
    "tenant_id":"<tenant-uuid>",
    "recipient_id":"<user-uuid>",
    "template_key":"leave.requested",
    "variables":{"leave_type":"ANNUAL","days":"5"},
    "channels":["NOTIFICATION_CHANNEL_EMAIL"]
  }' \
  localhost:50055 \
  hris.notification.v1.NotificationService/SendNotification
```

## Development Workflow

### Run Unit Tests

```bash
go test ./services/auth-service/internal/...
go test ./services/employee-service/internal/...
# ... repeat for all services
```

### Run Integration Tests

```bash
# Requires PostgreSQL + NATS running
go test -tags integration ./services/auth-service/internal/...
go test -tags integration ./services/employee-service/internal/...
```

### Format & Lint Code

```bash
go fmt ./services/...
go vet ./services/...
```

### View Logs

```bash
# Docker
docker-compose -f docker-compose.production.yml logs -f auth-service

# Local binary
# Logs go to stdout
```

## Stopping Services

```bash
# Docker
docker-compose -f docker-compose.production.yml down

# Manual
# Press Ctrl+C in each terminal
```

## Troubleshooting

**Q: "PostgreSQL connection refused"**  
A: Ensure PostgreSQL is running and database/schemas exist. Check DATABASE_URL.

**Q: "NATS connection refused"**  
A: Ensure NATS server is running with JetStream enabled: `nats-server -js`

**Q: "Port already in use"**  
A: Change port in environment: `export GRPC_PORT=50061` or kill existing process.

**Q: "gRPC Health Check fails"**  
A: Check service logs for startup errors. Service may not have completed initialization.

---

See `staging.md` and `production.md` for deployment to staging/production environments.
