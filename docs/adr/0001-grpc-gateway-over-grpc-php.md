# ADR-0001: Use grpc-gateway Instead of gRPC-PHP for Laravel-to-Service Communication

**Status:** Accepted  
**Date:** 2026-05-30  
**Deciders:** Architecture Team  
**Supersedes:** —

---

## Context

The system architecture requires Laravel (API Gateway) to call Go microservices. Go services communicate with each other using native gRPC. The question is how Laravel should reach those services.

The research document initially suggested using `grpc/grpc-php` (the official PHP gRPC client). After evaluating this option against alternatives, the grpc-gateway approach is superior for this project.

### Option A: grpc/grpc-php (native PHP gRPC)

PHP gRPC requires:
- The `grpc` C extension compiled into the PHP runtime
- The `protoc-gen-php` plugin for code generation
- A non-trivial Docker build (multi-stage with extension compilation)
- Runtime dependency on a native extension that is difficult to debug in containers

Known problems:
- `grpc/grpc-php` lags behind the Go gRPC SDK in feature support
- Extension compilation adds 3–5 minutes to Docker build times
- Debugging gRPC connection issues in PHP is significantly harder than HTTP
- The PHP gRPC ecosystem has fewer maintained examples and community support
- Extension version pinning creates upgrade friction

### Option B: grpc-gateway (selected)

`grpc-gateway` is a protoc plugin that reads gRPC service definitions and generates a reverse-proxy server that translates HTTP/JSON calls into gRPC calls. The generated gateway runs as part of each Go service.

Laravel calls Go services via **HTTP/JSON** using Guzzle. Go services remain gRPC-native for service-to-service communication.

```
Laravel (Guzzle HTTP client)
  → HTTP/JSON (grpc-gateway reverse proxy on :808x)
    → gRPC (internal, Go-to-Go)
```

### Option C: Envoy sidecar proxy

A third option is to run Envoy as a sidecar to translate gRPC ↔ HTTP/JSON at the network layer. This is valid but adds operational complexity (Envoy config, sidecar container management) that is not warranted at this scale.

---

## Decision

**Use grpc-gateway.** Each Go service exposes two ports:
- `:5005x` — native gRPC (for service-to-service communication)
- `:808x` — HTTP/JSON (grpc-gateway, for Laravel and external tooling)

HTTP annotations are added to all proto service definitions. The grpc-gateway binary is generated from these annotations via `buf generate`.

Laravel communicates with Go services exclusively via the HTTP/JSON port using typed Service classes backed by Guzzle.

---

## Proto Annotation Pattern

```proto
import "google/api/annotations.proto";

service AuthService {
  rpc Login(LoginRequest) returns (LoginResponse) {
    option (google.api.http) = {
      post: "/v1/auth/login"
      body: "*"
    };
  }
}
```

---

## Consequences

### Positive
- No native PHP extensions required. Laravel Dockerfile remains a standard PHP-FPM image.
- HTTP/JSON is debuggable with standard tools (curl, Postman, browser DevTools).
- grpc-gateway generated code is type-safe and kept in sync with protos via `buf generate`.
- Go services can be called by any HTTP client — useful for testing, admin tooling, and future SDKs.
- Service-to-service communication remains native gRPC (full performance, streaming, metadata).

### Negative
- grpc-gateway adds a thin HTTP layer on each Go service. Negligible latency overhead (<1ms).
- HTTP/JSON loses gRPC streaming for Laravel-initiated calls. Acceptable: Laravel does not consume streaming RPCs. The `ai-service` Chat stream is consumed directly by SvelteKit via SSE, not through Laravel.
- Proto annotations must be maintained alongside the service definitions. `buf lint` enforces this.

### Neutral
- Service-to-service gRPC calls are unaffected. This decision only changes how Laravel reaches Go services.

---

## Compliance

- All proto files must include grpc-gateway HTTP annotations for every RPC.
- `buf generate` must be run after any proto change.
- Laravel Service classes (`app/Services/`) must be typed wrappers over Guzzle — no raw `Http::post()` calls in controllers.
