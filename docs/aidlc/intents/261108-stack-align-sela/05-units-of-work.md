# 05 - Units of work

| Unit | Content | Gate | Depends on | Bolt |
| --- | --- | --- | --- | --- |
| U1 | Go module, copies of SELA `httpx`, `health`, `db/migrate.go`, `testdb`, `config.Load` | D1 | - | B1 |
| U2 | goose migrations, `0001_bootstrap.sql`, one file per old up/down pair | D1 | U1 | B1, B2 |
| U3 | `internal/events` publisher and consumer on Dragonfly Streams | D3 | U1 | B1 |
| U4 | `internal/authn` (and POS `internal/ratelimit`) | D2 | U1 | B1 |
| U5 | Service ports to `services/<name>` with `http.go` | D2 | U1-U4 | B2 |
| U6 | Wiring in `cmd/api` | D2 | U5 | B3 |
| U7 | Frontend toolchain and API contract readers | D4 | U5 | B4 |
| U8 | Compose, Makefile, `scripts/check.sh`, CI | D1 | U6 | B4 |
| U9 | Deletions of the old tree | D1 | U6 | B5 |
| U10 | ADRs, `CLAUDE.md`, `.claude/`, FSD stack rows | D1-D4 | U6 | B4 |

## Route table (old path to new path)

Planned routes, decided 2026-11-08 (roles the old gateway never defined were set by the orchestrator; see `audit.md`). Health: `GET /healthz`, public.

| Service | RPC | Old path | New method + path | Public | Roles | Paginated | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| auth | Login | `POST auth/login` | `POST /api/v1/auth/login` | yes | - | no | 200 |
| auth | RefreshToken | `POST auth/refresh` | `POST /api/v1/auth/refresh` | yes | - | no | 200 |
| auth | RevokeToken | `POST auth/logout` (gateway) | `POST /api/v1/auth/logout` | no | any authenticated | no | 200 |
| auth | ValidateToken | - | none | no | no route (gateway-only) | no | 200 |
| auth | GetPermissions | - | `GET /api/v1/auth/permissions/{user_id}` | no | self or hr_admin | no | 200 |
| auth | RegisterTenant | - | `POST /api/v1/auth/register` | yes | - | no | 201 |
| employee | CreateEmployee | - | `POST /api/v1/employees` | no | hr_admin | no | 201 |
| employee | GetEmployee | - | `GET /api/v1/employees/{id}` | no | any authenticated | no | 200 |
| employee | UpdateEmployee | - | `PATCH /api/v1/employees/{id}` | no | hr_admin | no | 200 |
| employee | TerminateEmployee | - | `POST /api/v1/employees/{id}/terminate` | no | hr_admin | no | 200 |
| employee | ListEmployees | - | `GET /api/v1/employees` | no | any authenticated | yes | 200 |
| employee | GetOrgChart | - | `GET /api/v1/employees/org-chart` | no | any authenticated | no | 200 |
| employee | CreateDepartment | - | `POST /api/v1/departments` | no | hr_admin | no | 201 |
| employee | ListDepartments | - | `GET /api/v1/departments` | no | any authenticated | no | 200 |
| employee | CreatePosition | - | `POST /api/v1/positions` | no | hr_admin | no | 201 |
| employee | ListPositions | - | `GET /api/v1/positions` | no | any authenticated | no | 200 |
| leave | ApplyLeave | `POST v1/leaves` | `POST /api/v1/leaves` | no | all authenticated | no | 201 |
| leave | ListLeaveRequests | `GET v1/leaves` | `GET /api/v1/leaves` | no | all authenticated | yes | 200 |
| leave | GetLeaveRequest | `GET v1/leaves/{id}` | `GET /api/v1/leaves/{id}` | no | all authenticated | no | 200 |
| leave | ApproveLeave | `POST v1/leaves/{id}/approve` | `POST /api/v1/leaves/{id}/approve` | no | manager, hr_admin | no | 200 |
| leave | RejectLeave | `POST v1/leaves/{id}/reject` | `POST /api/v1/leaves/{id}/reject` | no | manager, hr_admin | no | 200 |
| leave | CancelLeave | `POST v1/leaves/{id}/cancel` | `POST /api/v1/leaves/{id}/cancel` | no | all authenticated | no | 200 |
| leave | GetLeaveBalance | `GET v1/leave-balance/{employeeId}` | `GET /api/v1/leave-balance/{employee_id}` | no | all authenticated | no | 200 |
| leave | ListLeaveTypes | `GET v1/leave-types` | `GET /api/v1/leave-types` | no | all authenticated | no | 200 |
| attendance | CheckIn | - | `POST /api/v1/attendance/check-in` | no | any authenticated | no | 200 |
| attendance | CheckOut | - | `POST /api/v1/attendance/check-out` | no | any authenticated | no | 200 |
| attendance | GetAttendanceSummary | - | `GET /api/v1/attendance/summary/{employee_id}` | no | any authenticated | no | 200 |
| attendance | ListAttendance | - | `GET /api/v1/attendance` | no | manager, hr_admin | yes (cursor: next_page_token, total_count) | 200 |
| attendance | OverrideAttendance | - | `POST /api/v1/attendance/override` | no | hr_admin | no | 200 |
| audit | Record | - | `POST /api/v1/audit/entries` | no | hr_admin | no | 201 |
| audit | Query | - | `GET /api/v1/audit/entries` | no | hr_admin | yes (limit/offset, total_count) | 200 |
| audit | GetEntry | - | `GET /api/v1/audit/entries/{entry_id}` | no | hr_admin | no | 200 |
| notification | SendNotification | - | `POST /api/v1/notifications` (proto `/send` verb dropped) | no | hr_admin | no | 201 |
| notification | GetDeliveryStatus | - | `GET /api/v1/notifications/{notification_id}` | no | hr_admin | no | 200 |
| notification | ListNotifications | - | `GET /api/v1/notifications` | no | any authenticated | yes | 200 |
| notification | MarkRead | - | `PATCH /api/v1/notifications` (body: notification_ids; proto verb dropped) | no | any authenticated | no | 200 |
| notification | UpdateChannelConfig | - | `PUT /api/v1/notification-channels/{channel}` | no | hr_admin | no | 200 |
| document | UploadDocument | - | `POST /api/v1/documents` (proto `/upload` dropped) | no | any authenticated | no | 201 |
| document | ListDocuments | - | `GET /api/v1/documents` | no | any authenticated | yes | 200 |
| document | DownloadDocument | - | `GET /api/v1/documents/{file_id}/download` | no | any authenticated | no | 200 |
| document | GetDocumentMetadata | - | `GET /api/v1/documents/{file_id}/metadata` | no | any authenticated | no | 200 |
| document | DeleteDocument | - | `DELETE /api/v1/documents/{file_id}` | no | hr_admin | no | 200 |
| document | GetPresignedUrl | - | `POST /api/v1/documents/{file_id}/presigned-url` | no | any authenticated | no | 200 |
| document | RollbackDocumentVersion | - | `POST /api/v1/documents/{file_id}/rollback` | no | hr_admin | no | 200 |
