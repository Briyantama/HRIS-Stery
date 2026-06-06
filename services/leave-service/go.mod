module github.com/hris-stery/hris-stery/services/leave-service

go 1.25.0

require (
	github.com/google/uuid v1.6.0
	github.com/hris-stery/hris-stery/gen/go v0.0.0-00010101000000-000000000000
	github.com/hris-stery/hris-stery/services/_shared v0.0.0-00010101000000-000000000000
	github.com/jackc/pgx/v5 v5.9.2
	github.com/nats-io/nats.go v1.52.0
	go.uber.org/zap v1.28.0
	google.golang.org/grpc v1.81.1
)

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.18.6 // indirect
	github.com/nats-io/nkeys v0.4.15 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.52.0 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace (
	github.com/hris-stery/hris-stery => ../..
	github.com/hris-stery/hris-stery/gen/go => ../../gen/go
	github.com/hris-stery/hris-stery/services/_shared => ../_shared
)
