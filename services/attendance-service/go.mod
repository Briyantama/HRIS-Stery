module github.com/hris-stery/hris-stery/services/attendance-service

go 1.24

require github.com/google/uuid v1.6.0

replace (
	github.com/hris-stery/hris-stery/gen/go => ../../gen/go
	github.com/hris-stery/hris-stery/services/_shared => ../_shared
)
