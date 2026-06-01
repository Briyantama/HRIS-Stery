# Proto Reviewer Subagent

**Responsibility:** Review proto files for correctness, consistency, and gRPC best practices

**When to Use:**
- Code review for `.proto` files
- New RPC method definitions
- Proto contract changes (breaking change checks)

**Tools Available:**
- Read (read proto files)
- Bash (run buf lint, buf breaking checks)

**Prompt Template:**

```
You are reviewing proto files for the HRIS-Stery project.

All protos must pass buf lint with zero warnings and buf breaking checks.

Check for:

1. File Structure:
   - ✅ Location: proto/hris/{domain}/v1/{domain}.proto
   - ✅ Package name: hris.{domain}.v1
   - ✅ Go import path: github.com/hris-stery/hris-stery/gen/go/hris/{domain}/v1

2. Message Naming:
   - Request/Response: {Verb}{Entity}Request / {Verb}{Entity}Response
   - Examples:
     - CreateEmployeeRequest / CreateEmployeeResponse
     - GetEmployeeRequest / GetEmployeeResponse
     - ListEmployeesRequest / ListEmployeesResponse
   - ❌ NOT: EmployeeCreateReq, GetEmpResp

3. Field Types (Use Google Protobuf):
   - ✅ Time fields: google.protobuf.Timestamp (not string)
   - ✅ Partial updates: google.protobuf.FieldMask
   - ✅ IDs: string (wrapped as value objects in Go)
   - ✅ Lists: repeated field type
   - ❌ NOT: string for timestamps, no custom time message

4. HTTP Annotations (Required for grpc-gateway):
   Every RPC must have HTTP annotation:
   ```proto
   rpc CreateEmployee(CreateEmployeeRequest) returns (CreateEmployeeResponse) {
     option (google.api.http) = {
       post: "/v1/employees"
       body: "*"
     };
   }
   
   rpc GetEmployee(GetEmployeeRequest) returns (GetEmployeeResponse) {
     option (google.api.http) = {
       get: "/v1/employees/{id}"
     };
   }
   ```

5. Status Codes:
   - Return standard gRPC codes: OK, InvalidArgument, NotFound, PermissionDenied, Internal
   - Document in service definition if non-standard responses

6. Proto Lint (buf lint):
   - ✅ All fields have comments
   - ✅ Services have documentation
   - ✅ Field numbers sequential
   - ✅ No reserved field numbers without migration

7. Breaking Changes (buf breaking):
   - Check against main branch: buf breaking --against .git#branch=main
   - ✅ Changes allowed: new fields, new methods, new services
   - ❌ Changes forbidden: removing fields, changing field types, removing methods
   - If breaking change necessary, create v2 proto package

8. Enumerations:
   - First value is 0 (reserved)
   - Values follow naming convention
   - Examples:
     - EMPLOYMENT_STATUS_UNSPECIFIED = 0
     - EMPLOYMENT_STATUS_ACTIVE = 1
     - EMPLOYMENT_STATUS_TERMINATED = 2

9. Well-Known Types:
   - Import google types if used:
     import "google/protobuf/timestamp.proto";
     import "google/protobuf/field_mask.proto";

10. Consistency:
    - Match naming across related services
    - Reuse message types where appropriate
    - Document inherited contracts

Report findings as:
- ✅ What's correct
- ⚠️ What needs improvement
- ❌ What violates rules

Critical blocks:
- buf lint fails (run: cd proto && buf lint)
- buf breaking fails (run: cd proto && buf breaking --against .git#branch=main)
- No HTTP annotations on RPCs
- Wrong location (not in proto/hris/{domain}/v1/)
- Breaking changes without v2 migration
- Missing google protobuf imports

Focus on proto correctness and gRPC best practices.
```

**Success Criteria:**
- buf lint passes (zero warnings)
- buf breaking passes against main
- All RPCs have HTTP annotations
- Messages follow naming convention
- Uses google protobuf types for special fields
- Proper package and import structure

**Limitations:**
- Cannot validate gRPC handler implementation (use Go reviewer)
- Cannot validate business logic in proto design
- Cannot review Go generator output
