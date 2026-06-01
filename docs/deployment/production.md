# Production Deployment Guide

## Prerequisites

- Kubernetes 1.24+ OR managed Docker environment
- PostgreSQL 16+ (managed service recommended: AWS RDS, Azure Database)
- NATS 2.10+ with JetStream enabled
- Redis 7+ (managed: AWS ElastiCache, Azure Cache)
- SSL/TLS certificates

## Infrastructure Requirements

### Database (PostgreSQL)

- Version: 16 or higher
- Replication: Multi-AZ for HA
- Backup: Daily snapshots + PITR
- Monitoring: CloudWatch/Prometheus
- RLS: Enabled (enforced by application)

**Connection Pool (Per Service):**
- Min connections: 5
- Max connections: 20
- Connection timeout: 30s
- Idle timeout: 900s

### Message Broker (NATS)

- Version: 2.10+
- Cluster: 3+ nodes for HA
- JetStream: Enabled with persistent storage
- Storage: Minimum 50GB
- Memory: 2GB+ per node

**Subjects:**
```
hris.identity.user.* (auth events)
hris.workforce.employee.* (employee events)
hris.operations.leave.* (leave events)
hris.operations.checkin.* (attendance events)
```

### Cache (Redis)

- Version: 7+
- Mode: Standalone or Cluster
- Persistence: RDB snapshots + AOF
- Memory: 1GB+ (adjust based on load)
- TTL: Enforce with policies

## Kubernetes Deployment

### Deployment Manifest Template

```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: hris
spec:
  replicas: 3  # HA
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      containers:
      - name: auth-service
        image: registry.example.com/hris/auth-service:1.0.0
        ports:
        - containerPort: 50051
          name: grpc
        env:
        - name: GRPC_PORT
          value: "50051"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: postgres-creds
              key: url
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: "http://otel-collector:4317"
        livenessProbe:
          exec:
            command:
            - /bin/grpc_health_probe
            - -addr=:50051
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          exec:
            command:
            - /bin/grpc_health_probe
            - -addr=:50051
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: auth-service
  namespace: hris
spec:
  selector:
    app: auth-service
  ports:
  - port: 50051
    targetPort: 50051
    name: grpc
  type: ClusterIP
```

## Secrets Management

### Required Secrets

```bash
# Database credentials
kubectl create secret generic postgres-creds \
  --from-literal=url=postgres://user:pass@host:5432/hris \
  -n hris

# Auth service JWT keys
kubectl create secret generic jwt-keys \
  --from-literal=private-key=$(cat private.key | base64) \
  --from-literal=public-key=$(cat public.key | base64) \
  -n hris

# SMTP credentials (notification-service)
kubectl create secret generic smtp-credentials \
  --from-literal=username=your-email@gmail.com \
  --from-literal=password=app-password \
  -n hris
```

## Health Checks & Monitoring

### gRPC Health Probes

All services expose `grpc.health.v1.Health/Check`:

```bash
# Manual check
grpcurl -plaintext <service-ip>:50051 grpc.health.v1.Health/Check

# Kubernetes probe
livenessProbe:
  exec:
    command:
    - /bin/grpc_health_probe
    - -addr=:50051
```

### Metrics Collection

Export to Prometheus/Datadog:

```yaml
env:
- name: OTEL_EXPORTER_OTLP_ENDPOINT
  value: "http://prometheus:4317"
```

Metrics to monitor:
- `rpc_requests_total` (by endpoint)
- `rpc_duration_seconds` (p50, p95, p99)
- `db_query_duration_seconds`
- `nats_messages_consumed_total` (consumer lag)

## Deployment Process

### Step 1: Pre-Deployment Checks

```bash
# Verify all dependencies healthy
kubectl get svc postgres redis nats -n hris

# Run smoke tests
go test -tags integration ./services/...

# Check migrations are reversible
migrate -path migrations/ -database $DB_URL version
migrate -path migrations/ -database $DB_URL down
migrate -path migrations/ -database $DB_URL up
```

### Step 2: Blue-Green Deployment

```bash
# 1. Deploy v1.1.0 alongside v1.0.0
kubectl set image deployment/auth-service \
  auth-service=registry.example.com/hris/auth-service:1.1.0 \
  -n hris

# 2. Wait for new pods to be healthy
kubectl rollout status deployment/auth-service -n hris

# 3. Switch traffic (automatic via Kubernetes)
# or manual via load balancer

# 4. Monitor for errors (30+ minutes)
kubectl logs -f deployment/auth-service -n hris

# 5. If issues, rollback
kubectl rollout undo deployment/auth-service -n hris
```

### Step 3: Database Migrations

```bash
# Run in single pod (not parallel)
kubectl run -it --image=migrate:latest migrate-runner \
  --command -- migrate \
  -path /migrations \
  -database $DATABASE_URL \
  up
```

## Rollback Procedures

### If Deployment Fails

```bash
# 1. Stop new deployment
kubectl patch deployment auth-service \
  -p '{"spec":{"replicas":0}}' -n hris

# 2. Check logs for error
kubectl logs deployment/auth-service -n hris --previous

# 3. Fix issue
# ...

# 4. Restore replicas
kubectl patch deployment auth-service \
  -p '{"spec":{"replicas":3}}' -n hris
```

### If Database Migration Fails

```bash
# 1. Stop all services
kubectl scale deployment -l app=hris-service --replicas=0 -n hris

# 2. Rollback migrations
kubectl run migrate-runner \
  --image=migrate:latest \
  --command -- migrate \
  -path /migrations \
  -database $DATABASE_URL \
  down 1

# 3. Fix schema issues
# ...

# 4. Re-run migration
kubectl run migrate-runner \
  --image=migrate:latest \
  --command -- migrate \
  -path /migrations \
  -database $DATABASE_URL \
  up

# 5. Restart services
kubectl scale deployment -l app=hris-service --replicas=3 -n hris
```

## Monitoring & Alerts

### Key Metrics to Monitor

- Service latency (target: p99 < 500ms)
- Error rate (target: < 0.1%)
- Database connection pool utilization
- NATS consumer lag (target: < 1s)
- Cache hit rate (target: > 70%)
- Disk usage (target: < 80%)

### Alert Rules (Example)

```yaml
- alert: HighRPCLatency
  expr: rpc_duration_seconds{quantile="0.99"} > 0.5
  for: 5m

- alert: ServiceErrorRate
  expr: rate(rpc_requests_total{status="error"}[5m]) > 0.001
  for: 5m

- alert: NATSConsumerLag
  expr: nats_consumer_lag_bytes > 1000000
  for: 10m
```

---

See `local.md` for local development setup.
