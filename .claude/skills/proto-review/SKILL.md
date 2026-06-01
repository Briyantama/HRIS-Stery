# Proto Review Skill

**Trigger:** Before implementing gRPC handlers for any service

**Purpose:** Ensure proto contracts are defined before handler code is written

## Instructions

1. Identify the service being modified (e.g., `employee-service`)
2. Extract the domain from the service name (e.g., `employee` from `employee-service`)
3. Check if proto file exists: `proto/hris/{domain}/v1/{domain}.proto`
4. If proto exists:
   - Read it and identify the RPC methods
   - Confirm all handler methods in the request have corresponding proto definitions
   - Check for HTTP annotations (required for grpc-gateway)
5. If proto does NOT exist:
   - STOP — Create proto file first
   - Use existing proto files as templates (e.g., `proto/hris/auth/v1/auth.proto`)
   - Run `buf lint` and `buf breaking` checks

## Proto Checklist

- [ ] Proto file exists in `proto/hris/{domain}/v1/`
- [ ] All RPC methods have request/response messages defined
- [ ] All messages follow naming: `{Verb}{Entity}Request` / `{Verb}{Entity}Response`
- [ ] HTTP annotations present for grpc-gateway
- [ ] `buf lint` passes (zero warnings)
- [ ] `buf breaking` passes against main branch
- [ ] Google protobuf types used (Timestamp, FieldMask, etc.)

## Commands

```bash
# Check if proto exists
ls proto/hris/{domain}/v1/{domain}.proto

# Lint proto file
cd proto && buf lint

# Check breaking changes
cd proto && buf breaking --against .git#branch=main
```

## Examples

**Trigger:** "Implement CreateEmployee RPC handler"
→ Check `proto/hris/employee/v1/employee.proto` for `CreateEmployeeRequest` / `CreateEmployeeResponse`

**Trigger:** "Add UpdateEmployee endpoint"
→ Proto must define UpdateEmployeeRequest and UpdateEmployeeResponse first

**Trigger:** "Add new service"
→ Create `proto/hris/{domain}/v1/{domain}.proto` before implementing any Go code
