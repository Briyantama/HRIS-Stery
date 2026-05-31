# Executive Summary  
This design document provides a complete blueprint for two modern SaaS products: a **multi-tenant AI-native HRIS** and a **POS (store) system**. Both leverage a front-end in SvelteKit (a Next.js-like framework【16†L113-L121】 with built-in SSR and optimized builds【16†L138-L146】) and a Laravel API gateway. Core business logic lives in Go microservices (high-performance gRPC services【24†L43-L50】). Services communicate via **gRPC** internally and publish events on **NATS** for loose coupling【5†L66-L74】. An independent AI service (Go) provides ML features. This setup follows industry best-practices: clean/DDD architecture, CQRS & event-driven patterns【4†L104-L112】【5†L66-L74】, and full observability (OpenTelemetry traces and metrics【29†L51-L54】). Multi-tenancy is handled by including tenant context (e.g. tenant ID in JWTs) in every request【13†L12-L21】. Comprehensive CI/CD pipelines automate linting, testing, and deployment. 

We outline two **Cursor AI prompts** (in English) that specify goals, scope, user stories, and interfaces for each project. We also provide “system rules” for Cursor (coding conventions, architecture principles), a recommended monorepo structure, and boilerplate code samples (proto definitions, service skeletons, Docker setup, CI workflow, etc.). This plan is rigorous yet action-oriented – a clear, corporate-style roadmap that you can pick up and execute. By following this blueprint, your team can *rapidly launch* robust, scalable HRIS and POS SaaS products with AI integration, with confidence in future scalability and maintainability.

## HRIS Project Prompt (Cursor AI)

- **Goal:** Build a multi-tenant HRIS (Human Resource Information System) SaaS with integrated AI. The platform manages companies (tenants), employees, attendance, leaves, payroll, recruitment, performance, etc. It must support Indonesian payroll rules eventually, provide analytics/chatbot features, and scale to thousands of users.  
- **Scope:**  
  - **Tenancy & Users:** Tenant signup/login; role-based access (HR Admin, Manager, Employee). Each tenant has isolated data (via tenant ID).  
  - **Core HR Modules:** Employee CRUD, Departments, Job Positions, Attendance logging, Leave requests/approvals, Payroll calculation (gross/net, BPJS, PPh21, THR), Recruitment (job postings, applications), Performance reviews.  
  - **AI Features:** Resume parsing and skill extraction, attrition risk forecasting, chat-based HR assistant, report generation insights.  
  - **UI:** Responsive admin dashboards (SvelteKit + Tailwind + shadcn-svelte components), with TanStack Query for data fetching and state.  
  - **Non-Functional:** High concurrency (Go services), secure (JWT+Sanctum, RBAC), multi-region support, full observability (OpenTelemetry + tracing【29†L51-L54】), automated testing, and CI/CD.  

- **Acceptance Criteria:**  
  - A user can register a new tenant/company and assign roles.  
  - HR Admin can add/edit employees, manage payroll schedules, and process payroll.  
  - Employees can submit leave requests; managers can approve. Approved leave generates an event.  
  - Payroll module correctly calculates one month’s salary including contributions and taxes.  
  - Data is isolated per tenant: no cross-tenant data leaks.  
  - AI module can analyze a sample resume text and return structured skills.  
  - System publishes domain events (e.g. `EmployeeCreated`, `LeaveApproved`) on NATS for other services to consume.  
  - Comprehensive tests and code linting are in place; deployment pipeline automatically builds and deploys images.  

- **Initial Tasks:**  
  1. Define multi-tenant strategy (JWT with tenant ID, or middleware to inject tenant context per request【13†L12-L21】).  
  2. Set up monorepo with `apps/` and `services/` folders (see folder tree below).  
  3. Create `proto/` definitions for services: Auth, Employee, Attendance, Leave, Payroll, Recruitment, Notification, AI, etc.  
  4. Scaffold Go microservices for each domain (using clean architecture templates, see “Handler / Usecase / Repo” structure).  
  5. Scaffold Laravel API Gateway (with Sanctum auth) to forward REST calls to gRPC.  
  6. Set up Docker Compose for local dev (Postgres, Redis, NATS, Elasticsearch, MinIO, etc.).  
  7. Establish CI pipeline (lint, test, build Docker) and define staging deployments.  
  8. Implement authentication (Laravel Sanctum, JWT tokens) and RBAC (roles/permissions).  
  9. Write database migration schemas (Employee, Tenant, etc.) and seed initial data.  
  10. Create SvelteKit front-end: login form, dashboard layout, and data fetching (TanStack Query).  

- **Example User Stories:**  
  - *As an HR Admin, I can log in and **view all employees** for my company.*  
  - *As an Employee, I can **check in/out attendance** via mobile/desktop.*  
  - *As an HR Admin, I can **approve or reject leave requests**; approved leaves trigger notifications to other systems.*  
  - *As a Payroll Admin, I can **generate payroll slips** for the month and email them to employees.*  
  - *As a User, I can chat with an **HR chatbot** to ask questions about benefits.*  

- **Proto & gRPC Endpoints:**  
  Define protobuf services and messages for all microservices. Key examples:  
  - **AuthService (gRPC):** `Login(LoginRequest) → AuthResponse`, `ValidateToken(TokenRequest) → UserContext`.  
  - **EmployeeService:** `CreateEmployee(CreateEmployeeReq) → Employee`, `ListEmployees(ListEmployeesReq) → EmployeeList`, `GetEmployee(GetEmployeeReq) → Employee`, `UpdateEmployee(UpdateEmployeeReq) → Employee`.  
  - **AttendanceService:** `CheckIn(CheckInReq) → Ack`, `CheckOut(CheckOutReq) → Ack`, `ListAttendance(ListAttendanceReq) → AttendanceList`.  
  - **LeaveService:** `ApplyLeave(ApplyLeaveReq) → LeaveResponse`, `ApproveLeave(ApproveLeaveReq) → LeaveResponse`.  
  - **PayrollService:** `GeneratePayroll(GenerateReq) → PayrollResult`, `GetPayslip(GetPayslipReq) → Payslip`.  
  - **RecruitmentService:** `PostJob(PostJobReq)`, `ApplyForJob(ApplyReq)`.  
  - **NotificationService:** `SendNotification(SendReq)`.  
  - **AIService:** `AnalyzeResume(AnalyzeReq) → AnalyzeResp`, `ForecastAttrition(ForecastReq) → ForecastResp`, `Chat(ChatReq) → ChatResp`.  

- **REST Gateway Endpoints (Laravel API):**  
  Expose JSON HTTP APIs mapped to gRPC calls (using [grpc-gateway](https://grpc-ecosystem.github.io/grpc-gateway/docs/tutorials/introduction/) to generate proxies【34†L65-L73】 or manually in controllers). Example routes:  
  - `POST /api/login` → AuthService.Login (returns JWT token).  
  - `GET /api/employees` → EmployeeService.ListEmployees.  
  - `POST /api/employees` → EmployeeService.CreateEmployee.  
  - `POST /api/attendance/checkin` → AttendanceService.CheckIn.  
  - `POST /api/leaves` → LeaveService.ApplyLeave.  
  - `POST /api/leaves/{id}/approve` → LeaveService.ApproveLeave.  
  - `GET /api/payroll/{month}/generate` → PayrollService.GeneratePayroll.  
  - `POST /api/ai/analyze-resume` → AIService.AnalyzeResume.  
  Use Laravel controllers to validate input (Form Requests), then call the gRPC client under the hood.  

- **NATS Events:**  
  Publish domain events for key actions to enable event-driven flows【5†L66-L74】. Examples:  
  - `Employee.Created` (payload: employee ID, tenant ID).  
  - `Leave.Requested`, `Leave.Approved`, `Leave.Rejected`.  
  - `Payroll.Processed` (when monthly payroll is complete).  
  - `User.Authenticated` (on login).  
  - `AI.AnalysisCompleted` (AI service can publish when done).  
  These subjects enable subscribers (e.g. NotificationService) to react. With JetStream, each event has “at least once” delivery【5†L66-L74】. NATS subjects are secured via ACLs (grant publish/subscribe by service)【21†L204-L212】.  

- **AI Service Contracts:**  
  The AI microservice (in Go) provides gRPC endpoints for ML features. Example proto:  
  ```proto
  syntax = "proto3";
  package ai;
  service AIService {
    rpc AnalyzeResume(AnalyzeResumeRequest) returns (AnalyzeResumeResponse);
    rpc ForecastAttrition(ForecastRequest) returns (ForecastResponse);
    rpc Chat(ChatRequest) returns (ChatResponse);
  }
  message AnalyzeResumeRequest { string text = 1; }
  message AnalyzeResumeResponse { repeated string skills = 1; }
  ```
  The AI service connects to LLMs (OpenAI, Gemini) or ML models internally. Use gRPC streams if needed for conversational chat.  

- **Security Rules:**  
  - **Authentication:** Use Laravel Sanctum for issuing JWTs/cookie sessions【18†L152-L160】. Protect the API routes via Sanctum middleware.  
  - **Authorization:** Implement RBAC: roles like `hr_admin`, `manager`, `employee`. Use policies or Spatie Permission in Laravel to enforce ACLs. Services trust the token: include user ID, tenant ID, and roles in JWT. Microservices always check the tenant context. AWS best practices suggest pushing tenant resolution into shared libraries【13†L12-L21】.  
  - **Multi-Tenancy:** Encode tenant ID in every request (header or JWT). In databases, either use separate schemas/tables per tenant or a `tenant_id` column filter. Avoid cross-tenant data access. Use middleware or gRPC interceptors to inject tenant context into logs and queries.  
  - **NATS Auth:** Configure NATS with permissions so only specific services can publish/subscribe to subjects (e.g. `auth-service` can publish `User.*`, `attendance-service` can subscribe to `Employee.*`). NATS ACL example: allow `auth-service` to publish on `User.*` and subscribe on `Auth.*`【21†L204-L212】.  
  - **Data Validation:** Use strong typing in protos and validate inputs in both gateway (Laravel Form Requests) and services (struct tags or manual checks).  

- **Testing Requirements:**  
  - **Unit Tests:** Each Go service has unit tests (use Go’s `testing` package) for usecase logic and repository methods. Laravel uses PHPUnit/Pest for controllers and service classes. SvelteKit uses Vitest/Playwright for frontend components/routes.  
  - **Integration Tests:** Spin up Docker Compose to test inter-service RPC calls. For example, test that the EmployeeService gRPC correctly writes to Postgres. Mock external APIs where needed (e.g. OpenAI).  
  - **Contract Tests:** Define simple “golden protos” tests: e.g. list a JSON gRPC response and compare. Use `grpcurl` or `protoc`-generated client libraries for end-to-end gRPC testing.  
  - **E2E/UI Tests:** Use Playwright or Cypress on the SvelteKit frontend to simulate key workflows (login, create employee, apply leave).  
  - **Observability Checks:** Assert that OpenTelemetry traces/logs appear for sample requests.  

- **CI/CD Triggers:**  
  - **On Pull Request:** Run CI pipeline (lint, unit tests, build). Deploy preview environments (e.g. with Laravel Envoy/Vercel or Kubernetes) for QA.  
  - **On Commit to `main`/`master`:** Trigger full build pipeline: run all tests, build Docker images for services and gateway, push images to registry, then deploy to staging/production. Use [GitHub Actions](https://docs.github.com/en/actions), GitLab CI or similar.  
  - **CD:** Configure auto-deploy to development or staging on every main commit. Manual promotion to production after QA review. Include rollback capabilities.  

## POS Project Prompt (Cursor AI)

- **Goal:** Build a multi-tenant POS SaaS system for retail stores. Features include product/inventory management, sales (cashier) transactions, supplier ordering, customer management, and receipts. Must be AI-ready (e.g. demand forecasting, product recommendation).  
- **Scope:**  
  - **Tenancy & Users:** Support multiple store chains (tenants), each with admin/cashier roles. Multi-store support per tenant (e.g. Store A, Store B).  
  - **Core Modules:** Product catalog, inventory levels, category management, purchase orders (to suppliers), sales entry (cashier interface), customers/loyalty, pricing/discounts, receipts/invoices, daily reports.  
  - **AI Features:** Inventory demand forecasting, dynamic pricing suggestions, image-to-product recognition, chat support for queries, trend analysis.  
  - **UI:** SvelteKit dashboards and cashier interface, barcode scanner integration (optional), offline support (PWA) if needed.  
  - **Non-Functional:** High concurrency at peak (Go services), real-time updates (e.g. using NATS for stock notifications), offline resilience, multi-tenant isolation, observability, CI/CD.  

- **Acceptance Criteria:**  
  - Store Admin can add products and define initial stock.  
  - Cashier can create a new **sale**, add items, and complete transaction (stock levels adjust). A digital receipt is stored.  
  - Inventory low alerts: When stock < threshold, an event is published, and admin sees a notification.  
  - Suppliers can be managed and purchase orders created to restock products.  
  - Customer records can be created and loyalty points tracked.  
  - Sales report (by day/store) can be generated.  
  - AI service can suggest reorder quantities based on past sales.  

- **Initial Tasks:**  
  1. Define tenant hierarchy (company→stores).  
  2. Monorepo scaffold (apps + services).  
  3. Define proto for Inventory, Product, Sales, Supplier, Customer, Notification, AI.  
  4. Scaffold Go services: InventoryService, ProductService, SalesService, SupplierService, CustomerService, NotificationService, AIService.  
  5. Setup Laravel gateway for authentication and HTTP endpoints.  
  6. Docker Compose dev env (Postgres, Redis, NATS, ES, MinIO, etc.).  
  7. CI/CD pipeline configuration.  
  8. Database schema: Product, Inventory (with location/store), Orders, Sales, Customer tables.  
  9. SvelteKit app: admin dashboard (product/inventory) and cashier UI (sales form, receipt printing).  

- **Example User Stories:**  
  - *As a Store Owner, I can register stores and add my first products to the inventory.*  
  - *As a Cashier, I can ring up a sale and print/display a receipt.*  
  - *As a Stock Manager, I get alerted when a product is low in stock.*  
  - *As an Admin, I can view daily sales reports across all stores.*  
  - *As a User, I can search products by name or scan barcode to add to sale.*  

- **Proto & gRPC Endpoints:**  
  - **AuthService:** (same as HRIS) `Login/Validate`.  
  - **InventoryService:** `AddItem(AddItemReq)`, `UpdateStock(UpdateStockReq)`, `ListItems(ListItemsReq)`, `GetItem(GetItemReq)`.  
  - **ProductService:** `CreateProduct`, `SearchProducts`, `UpdateProduct`.  
  - **SalesService:** `CreateSale(CreateSaleReq)`, `GetSalesReport(GetReportReq)`.  
  - **SupplierService:** `AddSupplier`, `CreatePurchaseOrder(CreatePOReq)`, `ReceiveStock(ReceiveReq)`.  
  - **CustomerService:** `CreateCustomer`, `GetCustomer`, `AddLoyaltyPoints`.  
  - **NotificationService:** for pushing real-time alerts (e.g. via WebSockets or push).  
  - **AIService:** `ForecastDemand(ForecastReq)`, `RecognizeProduct(ImageReq)`, `ChatSupport(ChatReq)`.  

- **REST Gateway Endpoints:**  
  - `POST /api/login`, `/api/logout`.  
  - `GET/POST /api/products` → ProductService.  
  - `GET/POST /api/inventory` → InventoryService.  
  - `POST /api/sales` → SalesService.CreateSale.  
  - `GET /api/sales/report` → SalesService.GetSalesReport.  
  - `GET/POST /api/suppliers`, `/api/purchase-orders`.  
  - `POST /api/ai/forecast-demand`.  
  Similar to HRIS, use Laravel controllers to proxy to gRPC clients.  

- **NATS Events:**  
  - `Product.LowStock` (when inventory dips below threshold).  
  - `Sale.Completed` (after a sale is recorded).  
  - `Inventory.Replenished` (after restock).  
  - `Customer.New` (when a new customer is created).  
  - `AI.OrderForecasted` (when AI computes reorder quantities).  
  These allow decoupled services: e.g., NotificationService listens on `Product.LowStock` to send alerts.  

- **AI Service Contracts:**  
  Example proto messages:  
  ```proto
  syntax = "proto3";
  package ai;
  service AIService {
    rpc ForecastDemand(ForecastRequest) returns (ForecastResponse);
    rpc RecognizeProduct(RecognitionRequest) returns (RecognitionResponse);
  }
  message ForecastRequest { repeated int32 pastSales = 1; int32 daysAhead = 2; }
  message ForecastResponse { repeated int32 forecastedSales = 1; }
  message RecognitionRequest { bytes image = 1; }
  message RecognitionResponse { string productId = 1; string name = 2; }
  ```  
  The AIService calls LLMs or ML models to predict demand and to identify products from images (optional).  

- **Security Rules:**  
  - **Auth & RBAC:** Same setup as HRIS (Sanctum, JWT, roles: `admin`, `cashier`). Ensure API tokens include tenant and store IDs. Services verify tokens and enforce data access by tenant.  
  - **Multi-Tenancy:** Add both `tenant_id` and optionally `store_id` to data models. Filter all queries by these IDs. Use middleware/interceptors to inject context. Libraries can abstract this (cf. AWS SaaS patterns【13†L12-L21】).  
  - **NATS Auth:** Define which microservice can publish/subscribe on subjects. For example, `inventory-service` can publish `Product.LowStock` events, but only `inventory-service` writes stock updates.  
  - **Data Validation:** Similar to HRIS: use protobuf schemas and Laravel validation. Prevent negative stock, validate email fields, etc.  

- **Testing Requirements:**  
  - **Unit Tests:** For Go services (business rules, repo methods). For Laravel (controllers and form validations). For front-end (Svelte components).  
  - **Integration Tests:** Test full sale flow: when a `Sale.Create` is called, it should deduct inventory and publish `Sale.Completed`. Use in-memory NATS or test subjects.  
  - **E2E Tests:** Simulate user scenarios: add product, perform a sale, check report.  
  - **Contract/Smoke Tests:** After each build, ensure gRPC endpoints respond (can use a dummy client or swagger).  

- **CI/CD Triggers:**  
  Same pattern as HRIS. On PR: run tests. On merge to main: build images (`go build`, `composer install`, `npm build`), push to registry, deploy. Possibly auto-deploy `alpha` and manual-promote to prod. Use preview environments for review.

## Cursor System Rules  

```markdown
# .cursor/rules/system.mdc
- Use **Clean Architecture** and **Domain-Driven Design** for all services. Push domain logic into **Use Cases/Interactors**, not controllers/handlers.
- Prefer **interfaces** for repositories/services and dependency injection. Avoid global state.
- Apply **CQRS** or event sourcing where useful: separate write models (commands) from read models (queries) when domain complexity grows.
- Embrace **Event-Driven Architecture**: business events (like `UserCreated`, `SaleCompleted`) should be published on a message bus (NATS JetStream).
- Always generate **production-ready code**: include error handling, logging, and documentation.
- Write **unit tests** and **integration tests** for each service.
- Maintain **SOLID** principles: single responsibility, etc.
- **Never** put business logic in controllers or SQL/DB code directly in handlers.
- Explain and document architectural decisions clearly.
```

```markdown
# .cursor/rules/golang.mdc
- Use **Go 1.25+** with modules. Follow idiomatic Go style (gofmt, golint). 
- Use a clear project layout:  
  ```
  cmd/{service-name}/        # main application entrypoint
  internal/                 # app-specific code
    /handler                # HTTP/gRPC handlers
    /usecase (or service)   # business logic layer
    /repository             # database access layer
    /model                  # domain models and protobuf definitions
  pkg/ (optional)          # shared libraries/utilities
  ```  
- Use **chi** router or grpc.Server for endpoints.
- Always use `context.Context` in handlers/usecases.
- Use **sqlc + pgx** for database (Postgres) with generated type-safe queries.
- Use **structured logging** (e.g. Uber zap or logrus) including request IDs and tenant IDs.
- Handle errors properly: wrap with context, return gRPC errors (`status.Errorf`) with appropriate codes.
- Use **NATS Go client** for messaging. Do not block; use async or JetStream.
- Use **OpenTelemetry** Go SDK for tracing (auto-instrument HTTP/gRPC, DB calls).
- Do not use global variables for shared resources (DB, NATS, etc); inject via constructors.
- Write **unit tests** with `go test`, mock external dependencies.
```

```markdown
# .cursor/rules/laravel.mdc
- Laravel is the **API Gateway** and **Auth** layer only.
- Do **not** put heavy processing or AI code in Laravel. Offload heavy tasks to microservices.
- Use **Sanctum** for API authentication (JWT tokens for SPAs/API). Issue a token on login.
- Implement **RBAC** (Spatie Permission or built-in gates/policies) to secure routes.
- Structure Laravel code with Controllers → Actions/Services → Repositories. Use Form Request classes for validation.
- Interact with gRPC services via generated gRPC clients (grpc/grpc-php or Timuch/client-generator).
- Handle exceptions globally; return JSON error responses. Do not expose raw stack traces.
- Configure CORS, rate limiting, and load balancing at the gateway.
- Use Laravel **queues** (Redis driver) for async jobs (e.g. send email receipts). Monitor with Horizon.
- Apply **database migrations** for each tenant (single database with tenant_id column or separate schemas). Consider using a package (e.g. spatie/laravel-multitenancy).
- Enable **OpenTelemetry** in Laravel (auto-instrument HTTP, DB) for tracing【29†L51-L54】.
- Use **composer scripts** and **Laravel Pint** for code style (PSR-12).
```

```markdown
# .cursor/rules/svelte.mdc
- Use **SvelteKit 5** with TypeScript (`.ts`/`.svelte` files).  
- Structure: `src/routes/` for pages; use `+page.svelte`, `+page.ts` (or `.server.ts` for data fetching).  
- Put UI components in `src/lib/` or co-locate in routes if only used once.  
- Use **TanStack Query** for server-state management (prefetch data in `load()` functions for SSR【37†L391-L399】).  
- Apply **Tailwind CSS** (already integrated). Use utility classes for styling. Optionally use shadcn-svelte UI components for common widgets.  
- For forms and validation, use libraries like **zod** for schemas and integrate with SvelteKit form actions.  
- Authentication: after login, store token in a secure cookie (via `setSession` or `load` in layout). Guard routes by checking auth in `hooks.server.ts`.  
- Enable **instrumentation**: use SvelteKit’s `instrumentation.server.js` for OpenTelemetry/tracing【32†L182-L185】.  
- Handle errors with SvelteKit’s `error.html` page template.  
- Ensure correct hydration: wrap `%sveltekit.body%` in `<div>`【32†L159-L168】.  
- Use ESLint/Prettier for frontend code style.  
- Write **frontend tests** with Vitest or Playwright (`src/tests`). Use `playwright.config.ts` for E2E.  
```

## Monorepo Folder Structure  
```
monorepo/
├── apps/
│   ├── web/              # SvelteKit frontend (HRIS or POS app)
│   │   ├── src/
│   │   ├── svelte.config.js
│   │   └── ...
│   ├── admin/            # (Optional) separate admin UI or PWA
│   └── api-gateway/      # Laravel project (routes, controllers, Sanctum auth)
│
├── services/             # Go microservices
│   ├── auth-service/     # handles login, token validation
│   ├── employee-service/
│   ├── attendance-service/
│   ├── leave-service/
│   ├── payroll-service/
│   ├── recruitment-service/
│   ├── performance-service/
│   ├── notification-service/
│   ├── ai-service/       # AI/Machine Learning tasks
│   ├── inventory-service/
│   ├── product-service/
│   ├── sales-service/
│   ├── supplier-service/
│   ├── customer-service/
│   └── ... (others as needed)
│
├── proto/                # Shared protobuf definitions
│   ├── auth.proto
│   ├── employee.proto
│   ├── attendance.proto
│   ├── leave.proto
│   ├── payroll.proto
│   ├── recruitment.proto
│   ├── performance.proto
│   ├── inventory.proto
│   ├── product.proto
│   ├── sales.proto
│   ├── supplier.proto
│   ├── customer.proto
│   ├── ai.proto
│   └── notification.proto
│
├── pkg/                  # Reusable libraries (logging, grpc utilities, etc.)
│   ├── logger/
│   ├── grpcclient/
│   ├── natsutil/
│   ├── datastore/
│   └── ...
│
├── deploy/
│   ├── docker/           # Dockerfiles for each service/app
│   ├── k8s/              # Kubernetes manifests/Helm charts
│   └── ci/               # CI/CD scripts (GitHub Actions, etc.)
│
├── docker-compose.yml    # Development environment (Postgres, Redis, NATS, ES, MinIO, etc.)
└── docs/
    ├── architecture.md
    ├── adr/              # Architecture Decision Records
    └── api.md            # API specifications (RESTful contracts)
```

## Boilerplate Code Samples

### Sample Protobuf Definitions (HRIS/ POS)  
```proto
// proto/employee.proto (HRIS)
syntax = "proto3";
package hris.employee;
message Employee {
  string id = 1;
  string name = 2;
  string email = 3;
  string role = 4;
  string department = 5;
  string tenant_id = 6;
}
message CreateEmployeeRequest { string name = 1; string email = 2; }
message EmployeeResponse { Employee emp = 1; }
message ListEmployeesRequest { string tenant_id = 1; }
message ListEmployeesResponse { repeated Employee employees = 1; }
service EmployeeService {
  rpc CreateEmployee(CreateEmployeeRequest) returns (EmployeeResponse);
  rpc ListEmployees(ListEmployeesRequest) returns (ListEmployeesResponse);
  rpc GetEmployee(Employee) returns (EmployeeResponse);
}
```
```proto
// proto/sales.proto (POS)
syntax = "proto3";
package pos.sales;
message SaleItem { string product_id = 1; int32 quantity = 2; decimal unit_price = 3; }
message CreateSaleRequest { 
  string store_id = 1; string cashier_id = 2; repeated SaleItem items = 3; 
}
message CreateSaleResponse { string receipt_id = 1; }
message SalesReportRequest { string store_id = 1; string date = 2; }
message SalesReportResponse { double total_sales = 1; int32 transactions = 2; }
service SalesService {
  rpc CreateSale(CreateSaleRequest) returns (CreateSaleResponse);
  rpc GetSalesReport(SalesReportRequest) returns (SalesReportResponse);
}
```

### Go Microservice Template (Example: Employee Service)  
```go
// services/employee-service/cmd/main.go
package main

import (
  "net"
  "os"
  "github.com/yourorg/employee-service/internal"
  "github.com/yourorg/employee-service/internal/handler"
  "github.com/yourorg/employee-service/internal/repository"
  "github.com/yourorg/employee-service/internal/usecase"
  "github.com/yourorg/employee-service/proto/hris" // generated pb
  "google.golang.org/grpc"
)

func main() {
  // Initialize DB, NATS, etc. (e.g. sql.DB via sqlc)
  db := InitPostgres() 
  repo := repository.NewEmployeeRepo(db)
  uc := usecase.NewEmployeeUsecase(repo)
  srv := grpc.NewServer()
  h := handler.NewEmployeeHandler(uc)
  h.RegisterService(srv) // registers gRPC handlers
  lis, err := net.Listen("tcp", ":50051"); if err != nil { panic(err) }
  grpc.Serve(lis, srv)
}
```
```go
// services/employee-service/internal/handler/handler.go
package handler

import (
  "context"
  pb "github.com/yourorg/employee-service/proto/hris"
)

type EmployeeUsecase interface {
  Create(ctx context.Context, name, email string) (string, error)
  List(ctx context.Context) ([]*pb.Employee, error)
}

type EmployeeHandler struct {
  uc EmployeeUsecase
  pb.UnimplementedEmployeeServiceServer
}

func NewEmployeeHandler(uc EmployeeUsecase) *EmployeeHandler {
  return &EmployeeHandler{ uc: uc }
}

func (h *EmployeeHandler) CreateEmployee(ctx context.Context, req *pb.CreateEmployeeRequest) (*pb.EmployeeResponse, error) {
  id, err := h.uc.Create(ctx, req.Name, req.Email)
  if err != nil { return nil, err }
  return &pb.EmployeeResponse{ Emp: &pb.Employee{Id: id, Name: req.Name, Email: req.Email} }, nil
}

func (h *EmployeeHandler) ListEmployees(ctx context.Context, req *pb.ListEmployeesRequest) (*pb.ListEmployeesResponse, error) {
  emps, err := h.uc.List(ctx)
  if err != nil { return nil, err }
  return &pb.ListEmployeesResponse{ Employees: emps }, nil
}
```
```go
// services/employee-service/internal/usecase/usecase.go
package usecase
import "context"

type EmployeeRepository interface {
  Create(ctx context.Context, name, email string) (string, error)
  List(ctx context.Context) ([]*Employee, error)
}

type EmployeeUsecase struct{ repo EmployeeRepository }
func NewEmployeeUsecase(r EmployeeRepository) *EmployeeUsecase { return &EmployeeUsecase{repo: r} }
func (u *EmployeeUsecase) Create(ctx context.Context, name, email string) (string, error) {
  // Business logic can go here (e.g., validations)
  return u.repo.Create(ctx, name, email)
}
func (u *EmployeeUsecase) List(ctx context.Context) ([]*Employee, error) {
  return u.repo.List(ctx)
}
```
```go
// services/employee-service/internal/repository/employee_repo.go
package repository

import (
  "context"
  "github.com/yourorg/employee-service/internal"
)
type employeeRepo struct{ db *sql.DB }
func NewEmployeeRepo(db *sql.DB) *employeeRepo { return &employeeRepo{db: db} }

func (r *employeeRepo) Create(ctx context.Context, name, email string) (string, error) {
  // Example using sqlc-generated query:
  empID, err := queries.CreateEmployee(ctx, sqlc.CreateEmployeeParams{
    ID: uuid.NewString(), Name: name, Email: email, TenantID: ctx.Value("tenant").(string),
  })
  return empID, err
}

func (r *employeeRepo) List(ctx context.Context) ([]*internal.Employee, error) {
  rows, err := queries.ListEmployees(ctx, ctx.Value("tenant").(string))
  // map sqlc results to internal.Employee slice
}
```

### Laravel API Gateway Skeleton  
```php
// apps/api-gateway/routes/api.php
use App\Http\Controllers\AuthController;
use App\Http\Controllers\EmployeeController;
use Illuminate\Support\Facades\Route;

Route::post('/login', [AuthController::class, 'login']);
Route::middleware('auth:sanctum')->group(function () {
    Route::get('/employees', [EmployeeController::class, 'index']);
    Route::post('/employees', [EmployeeController::class, 'store']);
    // ... other routes
});
```
```php
// apps/api-gateway/app/Http/Controllers/EmployeeController.php
namespace App\Http\Controllers;
use Illuminate\Http\Request;
use Grpc\EmployeeServiceClient;       // assume gRPC client
use EmployeeService\CreateEmployeeRequest;

class EmployeeController extends Controller
{
    public function index(Request $req)
    {
        $client = new EmployeeServiceClient(env('EMPLOYEE_SERVICE_HOST'));
        $grpcReq = new ListEmployeesRequest();
        $grpcReq->setTenantId($req->user()->tenant_id);
        list($resp, $status) = $client->ListEmployees($grpcReq)->wait();
        if ($status->code !== \Grpc\STATUS_OK) {
            return response()->json(['error'=>'Failed to list employees'], 500);
        }
        return response()->json($resp->getEmployees());
    }

    public function store(Request $req)
    {
        $data = $req->validate(['name'=>'required', 'email'=>'required|email']);
        $client = new EmployeeServiceClient(env('EMPLOYEE_SERVICE_HOST'));
        $grpcReq = new CreateEmployeeRequest();
        $grpcReq->setName($data['name']);
        $grpcReq->setEmail($data['email']);
        $grpcReq->setTenantId($req->user()->tenant_id);
        list($resp, $status) = $client->CreateEmployee($grpcReq)->wait();
        if ($status->code !== \Grpc\STATUS_OK) {
            return response()->json(['error'=>'Create failed'], 500);
        }
        return response()->json(['id'=>$resp->getEmp()->getId()], 201);
    }
}
```
This Laravel gateway logs in users with Sanctum and proxies requests to Go gRPC services. It **validates requests** (via `$req->validate`), injects the tenant ID from the authenticated user, calls the gRPC client, and returns JSON.

### SvelteKit Frontend Skeleton (HRIS example)  
```svelte
<!-- apps/web/src/routes/+layout.svelte -->
<script lang="ts">
  import { QueryClient, QueryClientProvider } from '@tanstack/svelte-query';
  import { browser } from '$app/environment';
  const queryClient = new QueryClient({ defaultOptions: { queries: { enabled: browser } } });
</script>
<QueryClientProvider client={queryClient}>
  <slot />
</QueryClientProvider>
```
```svelte
<!-- apps/web/src/routes/login/+page.svelte -->
<script lang="ts">
  import { goto } from '$app/navigation';
  let email = '', password = '';
  async function submit() {
    const res = await fetch('/api/login', {
      method: 'POST', headers: {'Content-Type':'application/json'},
      body: JSON.stringify({email, password})
    });
    if (res.ok) {
      // assume token is in cookie via Sanctum
      goto('/dashboard');
    } else {
      alert('Login failed');
    }
  }
</script>

<form on:submit|preventDefault={submit}>
  <input bind:value={email} type="email" placeholder="Email" required />
  <input bind:value={password} type="password" placeholder="Password" required />
  <button type="submit">Log In</button>
</form>
```
```svelte
<!-- apps/web/src/routes/dashboard/+page.svelte -->
<script lang="ts">
  import { createQuery } from '@tanstack/svelte-query';
  // Fetch list of employees from the gateway REST endpoint
  const employeesQuery = createQuery(['employees'], async () => {
    const res = await fetch('/api/employees');
    if (!res.ok) throw new Error('Fetch failed');
    return await res.json();
  });
</script>

{#if employeesQuery.isLoading}
  <p>Loading employees...</p>
{:else if employeesQuery.error}
  <p>Error loading employees.</p>
{:else}
  <h2>Employees</h2>
  <ul>
    {#each employeesQuery.data as emp}
      <li>{emp.name} ({emp.email})</li>
    {/each}
  </ul>
{/if}
```
This SvelteKit app uses TanStack Query for data, and interacts with the Laravel gateway at `/api/...`. Authentication state is managed via session cookies set by Sanctum.

### Dockerfiles & Dev Setup  

```dockerfile
# apps/api-gateway/Dockerfile (Laravel)
FROM php:8.2-fpm
# Install extensions & Composer
RUN apt-get update && apt-get install -y git unzip libpq-dev \
    && docker-php-ext-install pdo_pgsql
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer
WORKDIR /var/www/html
COPY . .
RUN composer install --no-dev --optimize-autoloader
EXPOSE 8000
CMD ["php", "artisan", "serve", "--host=0.0.0.0", "--port=8000"]
```
```dockerfile
# services/employee-service/Dockerfile (Go service)
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/employee-service ./cmd
FROM alpine:latest
COPY --from=builder /bin/employee-service /employee-service
EXPOSE 50051
ENTRYPOINT ["/employee-service"]
```
```dockerfile
# apps/web/Dockerfile (SvelteKit)
FROM node:20 AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build
FROM node:20-alpine
WORKDIR /app
COPY --from=build /app/build ./build
EXPOSE 3000
CMD ["node", "/app/build"]
```

```yaml
# docker-compose.yml (dev environment)
version: "3.8"
services:
  api-gateway:
    build: ./apps/api-gateway
    ports: ["8000:8000"]
    env_file: .env
    depends_on: [postgres, redis, nats, minio]

  employee-service:
    build: ./services/employee-service
    depends_on: [postgres, nats]
  
  # ... other Go services ...

  web:  # SvelteKit frontend (optional in compose)
    build: ./apps/web
    ports: ["3000:3000"]
    depends_on: [api-gateway]

  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
      POSTGRES_DB: mydb
    ports: ["5432:5432"]
  redis:
    image: redis:7
    ports: ["6379:6379"]
  nats:
    image: nats:latest
    ports: ["4222:4222", "8222:8222"]  # NATS server and monitor
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.8.1
    environment: [ "discovery.type=single-node", "ES_JAVA_OPTS=-Xms512m -Xmx512m" ]
    ports: ["9200:9200"]
  minio:
    image: minio/minio
    command: server /data
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports: ["9000:9000"]
```

### CI/CD Workflow Example (GitHub Actions)  
```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  lint-build-test:
    runs-on: ubuntu-latest
    services:
      postgres:  # for integration tests if needed
        image: postgres:15
        env:
          POSTGRES_DB: testdb
          POSTGRES_USER: user
          POSTGRES_PASSWORD: pass
        ports: [5432:5432]
    steps:
      - uses: actions/checkout@v3
      - name: Setup Go, PHP, Node
        uses: actions/setup-go@v4
        with: go-version: '1.25'
      - uses: actions/setup-php@v4
        with: php-version: '8.2'
      - uses: actions/setup-node@v3
        with: node-version: '20'
      - name: Run Go tests
        working-directory: services/employee-service
        run: go test ./...
      - name: Run Laravel tests
        working-directory: apps/api-gateway
        run: |
          composer install
          cp .env.example .env
          php artisan migrate --force
          php artisan test
      - name: Run SvelteKit tests
        working-directory: apps/web
        run: |
          npm ci
          npm test
      - name: Build and Push Docker images
        run: |
          docker build -t myorg/api-gateway:latest ./apps/api-gateway
          docker build -t myorg/employee-service:latest ./services/employee-service
          # ... build other images ...
          echo "PUSH_IMAGES_PLACEHOLDER"
```

## Conventions & Libraries  

- **Coding Style:**  
  - *Go:* use `gofmt` and `golangci-lint`. Follow Effective Go. Organize code in `cmd/`, `internal/`, `pkg/`. Use error wrapping (`fmt.Errorf`) and avoid naked returns.  
  - *PHP/Laravel:* PSR-12 standard. Use Laravel Pint (PHP-CS-Fixer). Service classes in `app/Services`, Form Requests in `app/Http/Requests`.  
  - *TypeScript/Svelte:* ESLint + Prettier. Sort imports alphabetically. Use zod for schema validation (Zod schemas generate TS types and runtime checks).  
- **Error Handling:** Return structured errors. In Go services, return `status.Errorf(codes.InvalidArgument, "…")` for gRPC. In Laravel, catch exceptions and return JSON with proper HTTP status. Never leak internal errors.  
- **Observability:** Integrate **OpenTelemetry** on all layers. Laravel uses auto-instrumentation (via extension or package)【29†L51-L54】. Go services use OpenTelemetry SDK (export to Jaeger/Prometheus). Add tracing middleware on HTTP/gRPC, and metrics (request count, latency). Instrument SvelteKit via `instrumentation.server.js`【32†L182-L185】.  
- **Secrets Management:** Store secrets in environment variables or a secret manager. Do not hardcode API keys or DB passwords. For local dev, use `.env` files (never commit). In Kubernetes, use Secrets.  
- **Database Migrations:**  
  - Laravel: use migrations (`artisan make:migration`) for gateway.  
  - Go: use a tool like `golang-migrate` or `Goose` in each service’s `deploy/` folder to manage its schema.  
- **Multi-Tenant Approach:**  
  - Prefer a **single database with a `tenant_id`** on every table, isolating data by filtering on that ID. (Alternatively, schema-per-tenant is more complex.)  
  - Implement a middleware/interceptor that extracts `tenant_id` from the JWT and sets it in context (Go) or a `Scoper` trait (Laravel).  
  - Follow AWS best practices: push tenant resolution logic into shared libraries so service code can remain tenant-agnostic【13†L12-L21】.  
- **Recommended Libraries:**  
  - *Go:* `chi` router, `sqlc` + `pgx` (Postgres), `zap` or `zerolog` for logging, `go-redis`, `nats.go` for NATS, `grpc` official libs, OpenTelemetry-Go.  
  - *Laravel:* `laravel/sanctum` (auth), `spatie/laravel-permission` (RBAC), `laravel/ai` (AI SDK【27†L189-L197】), `laravel/horizon` (queue monitoring), `beyondcode/laravel-websockets` (realtime notifications), `spatie/laravel-multitenancy`.  
  - *Frontend:* `@tanstack/svelte-query`, `zod` (validation), `sveltekit-auth` or `lucia` (if needed), `shadcn-svelte` (UI components), `svelte-routing` (if not using file router), Playwright for E2E, Vitest for unit.  

## Responsibilities (Service vs Gateway)  

| Component           | Responsibility                                                              |
|---------------------|-----------------------------------------------------------------------------|
| **Laravel Gateway** | - Authenticate users (Sanctum) and issue JWTs. <br> - Central RBAC and policies. <br> - Input validation (Form Requests). <br> - Simple orchestration (e.g. combine data from multiple services). <br> - Expose HTTP/JSON endpoints (via gRPC-Gateway). <br> - Trigger background jobs/notifications (queue commands). |
| **Go Microservices**| - Core business logic and data access (CRUD operations in DDD layers). <br> - High-concurrency tasks (batch payroll, report generation). <br> - Publish/subscribe to NATS events. <br> - Handle heavy compute (AI processing, analytics). <br> - Independent scaling per service.                                        |

## MVP Task Checklist (First 30)  

| # | Task                                                                 |
|---|----------------------------------------------------------------------|
| 1 | Define tenant data model (e.g. add `tenant_id` column to each table).|
| 2 | Initialize Git monorepo; create `apps/`, `services/`, `proto/` dirs.|
| 3 | Scaffold `apps/api-gateway` as a new Laravel project.               |
| 4 | Install Sanctum in Laravel; configure user model with `tenant_id`.   |
| 5 | Write initial Laravel routes (e.g. `/api/login`) and controllers.    |
| 6 | Create `proto/auth.proto` and generate gRPC stubs for Go and PHP.    |
| 7 | Scaffold `services/auth-service` (Go) with basic Login/Validate methods.|
| 8 | Set up Postgres & Redis in Docker Compose; configure Laravel env.    |
| 9 | Implement Laravel login: validate user, call AuthService gRPC, issue token.|
| 10| Scaffold SvelteKit app (`apps/web`); set up Tailwind and ESLint.      |
| 11| Build a SvelteKit login page; call `/api/login`, handle JWT cookie.   |
| 12| Scaffold `services/employee-service` (Go) with Create/List RPCs.     |
| 13| In Laravel, add `/api/employees` endpoints; proxy to EmployeeService.  |
| 14| Implement `CreateEmployee` in Go: save to DB (sqlc/migrations).       |
| 15| Run migrations for Employee table (id, name, email, tenant_id).       |
| 16| In SvelteKit, after login, fetch `/api/employees` and display list.   |
| 17| Scaffold `services/attendance-service` and gRPC methods (CheckIn).   |
| 18| Add `/api/attendance/checkin` to Laravel; connect to AttendanceService.|
| 19| Set up NATS in Docker; publish `Employee.Created` event after creation.|
| 20| Scaffold `services/leave-service` (ApplyLeave, ApproveLeave RPCs).    |
| 21| Add leave request forms/UI in SvelteKit; hook into `/api/leaves`.     |
| 22| Scaffold `services/payroll-service` (GeneratePayroll RPC).            |
| 23| Add basic payroll page: invoke payroll generation via API.           |
| 24| Write NATS consumers: e.g. NotificationService subscribing to events.   |
| 25| Scaffold `services/ai-service`; add `AnalyzeResume` RPC stub.        |
| 26| Create a command/job in Laravel that calls AI gRPC for a sample text. |
| 27| Set up OpenTelemetry for Laravel and Go (install libraries).         |
| 28| Write unit tests for one Go service and one Laravel controller.       |
| 29| Configure GitHub Actions: run tests and build Docker images.          |
| 30| Deploy a dev environment (e.g. Kubernetes or Laravel Cloud preview).  |

Each step aligns with the architecture and ensures a working MVP. From here, the system can scale organically (adding more services like Recruitment, POS inventory, etc.) without major rewrites.

**Sources:** Architecture and technology choices are based on official docs and best practices (e.g., SvelteKit guides【16†L113-L121】【32†L182-L185】, Laravel docs【19†L73-L81】【18†L152-L160】【27†L189-L197】, gRPC resources【24†L43-L50】【34†L65-L73】, NATS guides【5†L66-L74】【21†L139-L148】, OpenTelemetry best practices【29†L51-L54】). These references ensure the solution is grounded in current standards.