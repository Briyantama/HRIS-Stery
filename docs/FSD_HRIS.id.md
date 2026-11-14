# FSD HRIS-Stery

*Functional Specification Document, Sistem Manajemen SDM (HRIS) multi-tenant. Menjelaskan perilaku yang sudah ada di kode (status + bukti) dan arsitektur target yang diselaraskan ke stack standar Stery Suite (SELA).*

Catatan status (2026-11-14): bukti as-is pada bagian 2-4 disesuaikan dengan ADR-0005/0006 (satu modul Go, sesi BFF); status FR tidak berubah kecuali disebut.

| | |
|---|---|
| **Nama Dokumen** | FSD HRIS-Stery |
| **Versi** | 1.0 (Draft) |
| **Tanggal** | 6 November 2026 |
| **Disusun oleh** | Tim Produk & Engineering |
| **Status** | Draft untuk ditinjau |
| **Dokumen Terkait** | `HRIS-Stery/docs/BRD_HRIS.id.md`, `HRIS-Stery/docs/PRD_HRIS.id.md`, `stery-suite/docs/00_Analisis_Dampak_Risiko.id.md`, `stery-suite/docs/BRD_Stery_Suite.id.md`, `stery-suite/docs/PRD_Stery_Suite.id.md`, `stery-suite/docs/FSD_Stery_Suite.id.md`, `SELA/CLAUDE.md`, `SELA/docs/FSD_Sela.md`, `HRIS-Stery/docs/adr/0001..0004` |
| **Perubahan dari vX** | Perubahan dari vX: Dokumen pertama |

---

## 1. Pendahuluan

### 1.1 Tujuan dokumen

Dokumen ini adalah spesifikasi fungsional dan teknis HRIS-Stery (kode produk `HRS`). Ia melakukan dua hal yang sengaja dipisahkan:

1. **Kondisi saat ini (as-is):** perilaku fungsional yang benar-benar ada di kode, tiap butir diberi status dan bukti jalur berkas. Kebenaran = kode.
2. **Kondisi target (to-be):** arsitektur HRIS setelah diselaraskan ke stack standar Stery Suite yang bersumber dari SELA (STACK-SPEC pada 2.1) dan keputusan D-HRS-01..12 (BRD 9.3). Selisih as-is ke to-be dicatat pada tabel delta 2.7.

Dokumen ini tidak mengubah kode apa pun. Penyelarasan stack adalah alur kerja yang dijadwalkan pada roadmap (BRD 12, PRD 11), bukan bagian dari penulisan dokumen ini.

### 1.2 Cakupan dan dokumen terkait

| Dokumen | Peran | Rujukan dari FSD ini |
|---|---|---|
| `BRD_HRIS.id.md` | Kebutuhan bisnis BR-HRS-01..22, keputusan D-HRS, pertanyaan terbuka Q-HRS, risiko RISK-HRS | Matriks 11.2 memetakan setiap BR ke FR |
| `PRD_HRIS.id.md` | User story US-HRS, persona, NFR-HRS, roadmap R0-R5 | Setiap FR di bagian 4 menjadi turunan dari fitur MoSCoW PRD 4 |
| `00_Analisis_Dampak_Risiko.id.md` | Analisis dampak dan risiko lintas produk | Bagian 2.7 dan 11.3 |
| `FSD_Stery_Suite.id.md` | Kontrak platform (port, SessionService, EntitlementService, IntegrationService) | Bagian 2.3, 2.5, 7.2 |
| `SELA/CLAUDE.md`, `SELA/docs/FSD_Sela.md` | Sumber STACK-SPEC | Bagian 2.1 |
| `docs/adr/0001..0004` | Keputusan arsitektur HRIS yang ada | Bagian 2.7 dan 4.09-4.12 |

Dalam cakupan: modul 01-08 (AUTH, EMP, ATT, LEV, NTF, AUD, DOC, AI) dan persyaratan keamanan lintas modul (`FR-HRS-SEC.*`). Modul 09-12 (PAY, TAX, REC, PRF) **Ditunda** (ADR-0004) dan hanya memiliki bagian pembenaran penundaan, tanpa FR.

### 1.3 Konvensi

- **Status** (kosakata tetap; padanan istilah baku `Done / Partial / Not started / Pending Verification / Deferred`): `Selesai` (Done; perilaku terverifikasi benar di kode), `Sebagian` (Partial; ada tetapi cacat atau tidak lengkap, cacat diuraikan), `Belum` (Not started; tidak ada di kode), `Pending Verification` (kegagalan disimpulkan dari pembacaan statis, belum dijalankan), `Ditunda` (Deferred; ditunda resmi, ADR-0004).
- **Bukti:** jalur berkas relatif terhadap repo `HRIS-Stery` (tanpa awalan `HRIS-Stery/`), dalam kolom berjudul `Bukti`. Nomor baris, bila perlu, ditulis di luar tanda kutip balik. Jalur target (belum ada) hanya muncul di tabel delta 2.7 dan di teks, tidak pernah di kolom `Bukti`. **Pemetaan jalur (2026-11-14):** banyak baris masih mengutip tata letak sebelum konsolidasi (`services/<name>-service/...`, `services/_shared/...`, `apps/api-gateway/...`, `proto/...`); sejak ADR-0005 kode berada di `apps/api/services/<name>/...` (domain, application, infrastructure, `http.go`), kode bersama di repo saudara `stery-kit`, dan gateway Laravel serta `proto/` sudah tidak ada. Tabel 2.2-2.5 dan baris yang disebut pada catatan status di atas sudah disesuaikan; jalur tata letak lama yang tersisa adalah riwayat sampai barisnya diverifikasi ulang.
- **Prioritas:** `M` (Must), `S` (Should), `C` (Could).
- **ID:** `FR-HRS-<MM>.<n>` (MM = nomor modul katalog), `FR-HRS-SEC.<n>` untuk keamanan lintas modul. BR, D, Q, RISK didefinisikan di BRD; US di PRD; NFR-HRS direferensikan pada bagian 9. SELA dirujuk dengan ID aslinya (mis. `SELA FR-05.1`).
- **Nilai bisnis yang tidak ditemukan di repo** (harga, SLA, kuota, ambang) ditulis `tbd` dengan rujukan Q-HRS. Dokumen ini tidak memuat angka `Rp` apa pun.
- **Estimasi** (usaha, kesiapan) ditandai `[INFERENCE]`.

### 1.4 Asal bukti

Seluruh klaim "saat ini" berasal dari pembacaan statis kode dan berkas konfigurasi repo `HRIS-Stery` (cabang `development`, commit terakhir 10 Juni 2026). **Tidak ada build, uji, linter, atau migrasi yang dijalankan** untuk menyusun dokumen ini; karena itu klaim bahwa suatu langkah "gagal" ditandai `Pending Verification` kecuali kegagalannya terlihat langsung dari teks (mis. sintaks SQL yang tidak valid). Beberapa klaim awal hasil pindai asisten ternyata keliru setelah dibaca ulang; koreksinya dicatat pada 11.3.

---

## 2. Arsitektur dan Stack

### 2.1 Stack (STACK-SPEC, sumber SELA)

Stack target seluruh produk Stery Suite adalah stack SELA (keputusan pemilik; D-HRS-11). Tabel berikut disalin dari STACK-SPEC. Kolom `Sumber (SELA)` menunjuk berkas SELA yang menjadi acuan.

| # | Lapisan | Standar | Sumber (SELA) |
|---|---|---|---|
| 1 | Tata letak repo | per produk: `apps/api`, `apps/web`, `deploy/`, `scripts/`, `docs/`, `Makefile`, `CLAUDE.md`, `AGENTS.md` | `SELA/CLAUDE.md` (bagian "Layout"); `SELA/docs/FSD_Sela.md` 2.6 |
| 2 | Bahasa | Go >= 1.25.1, satu modul per produk `github.com/<owner>/<product>` (handle pemilik `Briyantama` seperti SELA, belum terverifikasi untuk repo baru) | `SELA/CLAUDE.md`; `SELA/apps/api/go.mod` |
| 3 | Dekomposisi | modul = paket Go `services/<name>` dalam satu proses; inti `Service` + repositori Postgres + adaptor HTTP + adaptor gRPC opsional; batas lewat antarmuka Go dan kontrak proto sehingga dapat dipisah kelak | `SELA/CLAUDE.md` ("Layout", "Conventions"); `SELA/docs/FSD_Sela.md` 2.2 |
| 4 | Transport | REST `/api/v1` pada `net/http` ServeMux stdlib, amplop `{success,data,error}` lewat `kit/httpx`; gRPC (`api/proto/v1`, buf, `gen/go` ter-commit, `scripts/proto-gen.sh --check`) untuk klien mesin/internal dengan interseptor deny-by-default; tanpa tingkat gateway; hanya Caddy sebagai reverse proxy | `SELA/docs/FSD_Sela.md` 2.3, 6 |
| 5 | Basis data | PostgreSQL 18 (`uuidv7()`), pgx/v5 SQL mentah (tanpa sqlc, tanpa ORM), goose v3.26.0 migrasi tertanam + `cmd/migrate up\|down`, `testdb.NewMigrated` (kloning templat) | `SELA/CLAUDE.md` ("Conventions"); `SELA/docs/FSD_Sela.md` 5 |
| 6 | Isolasi tenant | sesuai D8 (HRS, POS, platform): RLS ENABLE+FORCE, `tenant_isolation` dengan `USING` dan `WITH CHECK`, `current_setting('app.tenant_id', true)::uuid`, diset per transaksi oleh `kit/tenantdb.WithTenant`; peran aplikasi NOBYPASSRLS, peran migrasi terpisah | Tambahan suite (SELA tidak memiliki konsep tenant); pola `SELA/CLAUDE.md` ("Invariants") |
| 7 | Penyimpan state | Dragonfly lewat go-redis/v9: sesi (hanya hash), batas laju, kuota, antrean, skrip Lua | `SELA/docs/FSD_Sela.md` 2.4 |
| 8 | Penyimpanan objek | RustFS (S3) lewat aws-sdk-go-v2 di balik `objstore.Store`; hanya URL bertanda tangan; pemindaian ClamAV `clamd` fail-closed di jalur unggah | `SELA/docs/FSD_Sela.md` 2.1, 8 |
| 9 | Autentikasi | sesi buram di cookie HttpOnly Secure, hash di Dragonfly; tabel RBAC (roles, permissions, actions, role_actions) + `RequirePermission(code)`; tanpa JWT, tanpa token di penyimpanan peramban | `SELA/docs/FSD_Sela.md` 3, 8 |
| 10 | Asinkron | worker ticker dalam proses + outbox Postgres + konsumen idempoten; tanpa NATS/Kafka | `SELA/docs/FSD_Sela.md` 2.5 |
| 11 | Observabilitas | `log/slog`, `/healthz`; lihat D11 (OpenTelemetry/Prometheus adalah keputusan pemilik `tbd`, Q-HRS-08) | `SELA/CLAUDE.md` |
| 12 | Konfigurasi | variabel lingkungan, validasi gagal-cepat, rahasia tidak pernah dicatat, `.env.example` | `SELA/CLAUDE.md` ("Conventions") |
| 13 | PII dan audit | `kit/piicrypt` (AES-256-GCM, AAD per tujuan, indeks buta, rotasi kunci), `kit/audit` `audit_logs` append-only dengan trigger DB | `SELA/docs/FSD_Sela.md` 8 |
| 14 | Pembayaran | `payment.Provider`, Midtrans QRIS (Xendit sudah dihapus), webhook bertanda tangan. **Tidak dipakai langsung oleh HRIS:** penagihan suite ada di platform (D5); HRIS hanya membaca entitlement | `SELA/docs/FSD_Sela.md` 2.1 |
| 15 | Web | SvelteKit 2 + Svelte 5 + Vite 8 + TypeScript 6, `ssr=false`, `adapter-static`, npm, Node >= 20.19; `apiFetch` + cermin amplop; i18n `{id,en}` dengan uji paritas; Vitest 4 + Testing Library; Playwright (kanal Chrome) | `SELA/CLAUDE.md`; `SELA/docs/FSD_Sela.md` 2.1, 10 |
| 16 | Deploy | Caddy; web statis; `deploy/docker-compose.yml` untuk Dragonfly, RustFS, profil ClamAV; port standalone bawaan API 8080, gRPC 9090, web 5173 | `SELA/docs/FSD_Sela.md` 2.1 |
| 17 | Gerbang mutu | `scripts/check.sh` (drift proto, gofmt, vet, golangci-lint, `go test -cover`, svelte-check, vitest coverage, e2e tumpukan nyata; `CHECK_STAGES`), GitHub Actions job go/web/e2e, lantai cakupan 80%, `-race` di CI | `SELA/CLAUDE.md` |
| 18 | Proses | TDD RED/GREEN, `docs/tracking/tracker.json` + `scripts/build_tracker.py`, FSD diperbarui bersama kode, artefak `docs/aidlc/` | `SELA/CLAUDE.md` |
| 19 | Lokal | Bahasa Indonesia bawaan, Inggris didukung, IDR, WIB | `SELA/CLAUDE.md` |

Catatan: pada mode `standalone` (bawaan) HRIS memakai stack di atas **dengan** pengguna, sesi, dan tenant miliknya sendiri; mode `suite` hanya mengganti adaptor identitas, entitlement, notifikasi, audit, dan peristiwa (2.3).

Status implementasi 2026-11-08 (ADR-0005, ADR-0006). Baris 1-3, 5, 11, 12 dan 17 sudah terpenuhi: satu modul `apps/api` dengan `cmd/api`, `cmd/migrate` dan `services/<name>`, migrasi goose di `apps/api/internal/db/migrations`, slog, Postgres 18, `scripts/check.sh` dan CI. Kode bersama berasal dari repo saudara `stery-kit` (`httpx`, `health`, `objstore`, `migrate`, `testdb`, `pgtenant`, `events`, `authn`, `ratelimit`; npm `@stery/kit`), bukan folder `kit/` di repo suite. Bagian yang implementasinya berbeda dari tabel ini:

| Baris | Seperti dibangun | Alasan |
|---|---|---|
| 4 | Hanya REST; tidak ada adaptor gRPC, `api/proto`, maupun `gen/go` | tidak ada klien mesin (ADR-0005, opsi C ditolak) |
| 7, 10 | Dragonfly stream `hris:events` dengan empat consumer group dan idempotensi `processed_events`; tanpa outbox Postgres; peristiwa dipublikasikan setelah commit | ADR-0006 |
| 9 | JWT RS256 diverifikasi di API (`authn.Authenticate`, peran per rute), refresh token dan daftar cabut di Dragonfly; web menyimpan access token di cookie HttpOnly (`hris_session`) di belakang proxy `/api` SvelteKit, tidak pernah di penyimpanan browser | ADR-0005; sesi buram tetap menjadi target (FR-HRS-SEC.4) |
| 15 | SvelteKit 2.63 + Svelte 5 + Vite 8 + TypeScript 6 dengan rute server (aksi login, proxy `/api`, `adapter-auto`), bukan `ssr=false` + `adapter-static` | BFF cookie HttpOnly membutuhkan server |
| 16 | `deploy/docker-compose.yml` (Dragonfly, RustFS, profil `api` opsional) dan `docker-compose.production.yml`; port API 8000 | port dipertahankan dari gateway sebelumnya |

### 2.2 Dekomposisi modul

Kondisi saat ini (2026-11-14, ADR-0005): satu modul Go `apps/api` (`cmd/api`, `cmd/migrate`) dengan tujuh paket di `apps/api/services/<name>`, satu proses, satu biner. Sebelum 2026-11-08 ini adalah tujuh layanan mandiri dengan port gRPC sendiri (50051-50056, audit bentrok dengan leave di 50054) ditambah gateway Laravel; port tersebut sudah tidak ada. Modul `ai` belum memiliki paket.

| Modul | Skema | Paket (saat ini) | Status | Bukti |
|---|---|---|---|---|
| 01 auth | `auth` | `services/auth` | Sebagian | `apps/api/services/auth/http.go` |
| 02 employee | `employee` | `services/employee` | Sebagian | `apps/api/services/employee/http.go` |
| 03 attendance | `attendance` | `services/attendance` | Sebagian | `apps/api/services/attendance/http.go` |
| 04 leave | `leave` | `services/leave` | Sebagian | `apps/api/services/leave/http.go` |
| 05 notification | `notification` | `services/notification` | Sebagian | `apps/api/services/notification/http.go` |
| 06 audit | `audit` | `services/audit` | Sebagian | `apps/api/services/audit/http.go` |
| 07 document | `document` | `services/document` | Sebagian | `apps/api/services/document/http.go` |
| 08 ai | (tidak ada) | (tidak ada) | Belum | `apps/api/cmd/api/main.go` (tanpa pengkabelan `ai`) |

Hal yang diamati:
- Bentrok port gRPC antara leave dan audit sudah hilang: tidak ada server gRPC (ADR-0005, opsi C ditolak); semua modul dipasang pada satu `http.Server` di `apps/api/cmd/api/routes.go`.
- `ai` tidak lagi memiliki paket maupun kontrak proto di repo.
- `services/_shared` digantikan repo saudara `stery-kit` (httpx, health, objstore, migrate, testdb, pgtenant, events, authn, ratelimit) dan `apps/api/internal/{config,db,testdb,ttlcache}`.
- Struktur internal tiap modul masih mengikuti domain/application/infrastructure ditambah adaptor `http.go`. Bagian domain dan aplikasi inilah yang dipertahankan pada penyelarasan (2.7).

### 2.3 Komponen antarmuka

**Saat ini (as-is):**

| Komponen | Perilaku | Status | Bukti |
|---|---|---|---|
| Web SvelteKit (BFF) | Aksi formulir login mengirim POST ke API Go `/api/v1/auth/login` dan menyimpan access token di cookie HttpOnly `hris_session`; `/api/<path>` adalah proxy BFF yang menyertakan bearer token; klien API di browser tidak mengirim header Authorization | Sebagian | `apps/web/src/routes/(auth)/login/+page.server.ts`, `apps/web/src/routes/api/[...path]/+server.ts`, `apps/web/src/lib/server/session.ts` |
| API HTTP Go | Satu `net/http` ServeMux di `cmd/api` yang memasang semua modul di `/api/v1`; peran ditegakkan per rute lewat `authn`; gateway Laravel 11 (Sanctum, spatie/permission, `BaseGrpcGatewayClient`) dan jalur grpc-gateway dihapus oleh ADR-0005 | Sebagian | `apps/api/cmd/api/routes.go`, `apps/api/cmd/api/main.go` |
| Pembatasan laju | Tidak diterapkan: tidak ada limiter yang dipasang di `cmd/api` (middleware `throttle` Laravel hilang bersama gateway) | Belum | `apps/api/cmd/api/routes.go` |
| Panggilan antarmodul | Antarmuka Go dalam proses (`deps.go` per modul); tanpa gRPC, mTLS tidak diperlukan | Sebagian | `apps/api/services/leave/deps.go` |
| HTTP metrik | tidak ada endpoint Prometheus/metrik di `apps/api`; observabilitas memakai `log/slog` dan `/healthz` | Belum | `apps/api/cmd/api/routes.go` |

**Target (to-be), per D-HRS-12 dan STACK-SPEC 4, 9:**

| Komponen target | Perilaku |
|---|---|
| `apps/web` (SPA) | SvelteKit `ssr=false`, `adapter-static`; memanggil REST `/api/v1` dengan cookie sesi; tanpa token di penyimpanan peramban |
| REST `/api/v1` | Adaptor HTTP di setiap modul pada ServeMux stdlib; amplop `{success,data,error}`; izin eksplisit per rute (`RequirePermission`); tenant dan aktor dari sesi |
| gRPC `api/proto/v1` | Kontrak mesin/internal dan klien platform; interseptor deny-by-default; token layanan HMAC |
| Port suite (mode `suite`) | `Identity`, `Tenant`, `Entitlement`, `Notification`, `Audit`, `Events` dengan dua adaptor (standalone, suite); rincian antarmuka Go di `FSD_Stery_Suite.id.md` 2.4 |

Adaptor `standalone` memakai tabel pengguna, sesi, peran, dan audit milik HRIS. Adaptor `suite` memanggil `SessionService/Introspect` (cache lokal 30 s), `EntitlementService/Check` (+ snapshot bertanda tangan, toleransi `tbd`, Q-HRS-11), dan `IntegrationService/Publish` platform (D-HRS-03..05).

### 2.4 Penggunaan Dragonfly

**Saat ini:** Dragonfly (protokol Redis, `go-redis/v9`) menyimpan refresh token, daftar cabut, dan stream peristiwa `hris:events` (ADR-0006). Cache department/position di employee dan cache pencarian di leave adalah peta dalam proses dengan TTL, bukan Dragonfly.

| Penggunaan | Status | Bukti |
|---|---|---|
| Refresh token dan daftar cabut di Dragonfly | Sebagian | `apps/api/services/auth/infrastructure/token_service.go`, `apps/api/services/auth/application/commands/refresh_token.go` |
| Cache departemen dalam proses (TTL) | Sebagian | `apps/api/services/employee/infrastructure/cache/department_cache.go` |
| Stream peristiwa `hris:events` | Sebagian | `apps/api/cmd/api/main.go` |

**Target:** satu instans Dragonfly per produk pada profil suite (port 6381 untuk HRIS, 6379 pada profil standalone). Kunci dan kegunaannya:

| Kunci (target) | Isi | TTL | Modul |
|---|---|---|---|
| `sess:<sha256(session_id)>` | metadata sesi (user_id, tenant_id, role, expires_at); hanya hash, tanpa nilai sesi mentah | `tbd` (Q-HRS-11) | 01 AUTH |
| `rl:login:<ip>` dan `rl:login:<acct>` | penghitung batas laju login | `tbd` | 01 AUTH, SEC |
| `lock:<user_id>` | penguncian akun setelah gagal berulang | `tbd` | 01 AUTH |
| `rl:api:<tenant>:<route-class>` | batas laju endpoint sensitif | `tbd` | SEC |
| `ai:quota:<tenant>:<yyyymm>` | kuota pemakaian AI per tenant | `tbd` (Q-HRS-09) | 08 AI |
| `ent:snap:<org>` | snapshot entitlement bertanda tangan (hanya mode suite) | 24 jam bawaan D4, `tbd` | 02 EMP |

### 2.5 Event dan outbox

**Saat ini:** NATS JetStream dihapus oleh ADR-0006; peristiwa kini mengalir lewat Dragonfly stream `hris:events` (`apps/api/cmd/api/main.go`) dengan publisher di `apps/api/services/<name>/infrastructure/` dan consumer di `apps/api/services/notification/consumers.go` dan `apps/api/services/audit/consumers.go`. Tabel peristiwa di bawah adalah as-is sebelum konsolidasi (NATS, ADR-0003); subjek dan cacatnya adalah riwayat sampai diverifikasi ulang. Amplop dasar: `event_id`, `event_type`, `schema_version` (=1), `tenant_id`, `actor_id`, `occurred_at`, `payload`.

| Subjek / jenis peristiwa (as-is) | Produsen | Konsumen yang ada | Status | Bukti |
|---|---|---|---|---|
| `hris.identity.user.registered` | auth | employee (membuat kerangka karyawan), audit, notification | Sebagian: payload tidak memuat `tenant_id` sehingga konsumen employee tidak dapat membuat rekaman; galat di-ack lalu hilang | `services/auth-service/internal/infrastructure/nats_publisher.go`, `services/employee-service/internal/infrastructure/nats_consumer.go` |
| `hris.identity.user.login_succeeded`, `login_failed` | auth | audit (jenis tidak dikenal: peringatan lalu ack) | Sebagian | `services/audit-service/internal/infrastructure/nats/identity_consumer.go` |
| `hris.identity.user.password_changed` | (dideklarasikan, tidak pernah dipancarkan) | - | Belum | `services/auth-service/internal/domain/user.go` |
| `hris.identity.tenant.created` | auth | - | Sebagian | `services/auth-service/internal/application/commands/register_tenant.go` |
| `hris.workforce.employee.created` | employee | leave (membuat saldo), notification, audit | Sebagian | `services/leave-service/internal/infrastructure/nats/employee_created_consumer.go` |
| `hris.workforce.employee.terminated` | employee | notification, audit | Sebagian | `services/employee-service/internal/infrastructure/nats_publisher.go` |
| `hris.workforce.employee.updated` | (dideklarasikan, tidak pernah dipancarkan karena `UpdateEmployee` belum ada) | - | Belum | `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| `hris.attendance.record.checked_in\|checked_out\|overridden` | attendance | tidak ada: audit hanya berlangganan `hris.operations.>` dan memetakan `hris.operations.attendance.marked\|corrected` | Sebagian: subjek tidak cocok dengan konsumen | `services/attendance-service/internal/domain/events.go`, `services/audit-service/internal/infrastructure/nats/operations_consumer.go` |
| `hris.operations.leave.requested\|approved\|rejected\|cancelled` | leave | notification (`leave_event_consumer`), audit | Sebagian: payload `{}` karena field privat tanpa `MarshalJSON` | `services/leave-service/internal/infrastructure/nats/event_publisher.go`, `services/notification-service/internal/application/consumers/leave_event_consumer.go` |
| `hris.operations.leave.balance_updated` | (dideklarasikan, tidak pernah dipublikasikan) | - | Belum | `services/leave-service/internal/domain/events.go` |
| `hris.notification.*` | (tidak ada produsen) | audit (konsumen ada) | Belum | `services/audit-service/internal/infrastructure/nats/notification_consumer.go` |

Cacat lintas peristiwa yang terbaca dari kode: (a) payload attendance dan leave kosong sehingga konsumen tidak mendapat data bisnis; (b) subjek attendance tidak sama dengan yang diharapkan konsumen audit; (c) `CLAUDE.md` bagian 6 mensyaratkan `event_id` UUID v7 sedangkan pembuat id ada pada tiap publisher dan jenisnya belum diverifikasi (`Pending Verification`); tidak ada outbox; (d) konsumen mencatat idempotensi di tabel `processed_events` per skema (pola yang dipertahankan).

**Target (D-HRS-05):** tanpa NATS. Per produk: peristiwa domain dalam proses + outbox transaksional Postgres (`FOR UPDATE SKIP LOCKED`) + konsumen idempoten (`processed_events`, ada di semua skema as-is). Lintas produk: outbox -> `IntegrationService/Publish` platform -> platform mengirim ke endpoint `Deliver` produk tujuan; opt-in per org lewat fitur entitlement `integration.<a>_<b>`; pengiriman at-least-once.

Amplop target: `{event_id UUIDv7, event_type, org_id, store_id?, actor_id, occurred_at, schema_version, payload}`. Untuk HRIS `store_id` tidak dipakai; `org_id` = `tenant_id` HRIS pada mode standalone.

Katalog peristiwa target (nama final mengikuti satu skema `hrs.<agregat>.<kejadian>`; penamaan `tbd` sampai ada keputusan pemilik):

| Peristiwa target | Produsen | Konsumen dalam produk | Konsumen lintas produk (hipotesis S5) |
|---|---|---|---|
| `hrs.employee.created` | 02 EMP | 04 LEV (saldo), 05 NTF, 06 AUD | INT-1 (POS) |
| `hrs.employee.updated` | 02 EMP | 06 AUD | - |
| `hrs.employee.terminated` | 02 EMP | 04 LEV, 05 NTF, 06 AUD | INT-1: nonaktifkan pengguna POS |
| `hrs.attendance.checked_in` / `checked_out` / `overridden` | 03 ATT | 06 AUD | - |
| `hrs.leave.requested` / `approved` / `rejected` / `cancelled` | 04 LEV | 03 ATT (status ON_LEAVE), 05 NTF, 06 AUD | - |
| `hrs.identity.tenant_registered` / `user_registered` | 01 AUTH | 02 EMP (kerangka karyawan), 05 NTF, 06 AUD | - |
| `hrs.gathering.scheduled` (hipotesis) | 02 EMP | - | INT-2 (SELA) |

INT-1 dan INT-2 adalah hipotesis (BR-HRS-21); validasi permintaan pasar dulu.

### 2.6 Struktur kode dan kontrak

Struktur repo target `HRIS-Stery` (jalur target, belum ada; baris tabel ini bukan `Bukti`):

| Jalur target | Isi |
|---|---|
| `apps/api/cmd/api` | titik masuk server (REST + gRPC + worker) |
| `apps/api/cmd/migrate` | `up\|down` goose |
| `apps/api/services/<name>` | modul `auth`, `employee`, `attendance`, `leave`, `notification`, `audit`, `document`, `ai` (domain, application, repositori pgx, adaptor HTTP, adaptor gRPC) |
| `apps/api/internal/db/migrations` | migrasi goose tertanam (mengganti `services/*/migrations`) |
| `api/proto/v1`, `gen/go` | kontrak proto ter-commit; `scripts/proto-gen.sh --check` |
| `apps/web` | SPA SvelteKit |
| `kit/` (repo `stery-suite`) | `httpx`, `config`, `db`, `testdb`, `audit`, `piicrypt`, `objstore`, `health`, `tenantdb`, `session`, `outbox` (D-HRS-01) |
| `deploy/`, `scripts/`, `docs/` | `docker-compose.yml`, `scripts/check.sh`, dokumen |

Kontrak yang dijaga lintas penyelarasan: (1) nama tabel dan kolom bertenant tetap dengan `tenant_id` sehingga isi data dapat dimigrasi; (2) kontrak proto `hris.*.v1` dipertahankan sebagai dasar `api/proto/v1` dengan perbaikan pada RPC yang bermasalah (bagian 6); (3) invarian RLS (D8) tidak pernah dilonggarkan selama migrasi.

### 2.7 Delta penyelarasan as-is -> to-be

Kolom reuse: domain dan application dipertahankan, infrastructure dan interfaces ditulis ulang. Estimasi usaha S (hari), M (minggu), L (beberapa minggu); seluruhnya `[INFERENCE]`.

| as-is -> to-be | modul terdampak | reuse (domain/application dipertahankan, infrastructure/interfaces ditulis ulang) | effort S/M/L [INFERENCE] |
|---|---|---|---|
| Gateway Laravel 11 dihapus -> adaptor HTTP Go `/api/v1` per modul (D-HRS-12) | semua modul; rute leave dan auth yang ada di `apps/api-gateway` | Kontrak proto dan validasi domain dipertahankan; controller PHP dan klien gRPC PHP dibuang; peta rute dari anotasi proto (bagian 6) | L |
| SSR SvelteKit (adapter-auto, cookie diset di server) -> SPA `ssr=false` + `adapter-static`, `apiFetch`, i18n `{id,en}` (D-HRS-07) | 4 halaman web | Komponen UI `src/lib/components/ui/*`, skema zod cuti dipertahankan; `+page.server.ts` dan `hooks.server.ts` ditulis ulang | M |
| JWT RS256 (15 menit/7 hari, hardcoded) -> sesi buram HttpOnly, hash di Dragonfly (D3) | 01 AUTH, semua interseptor, web | Entitas `User`, aturan password, bcrypt, RBAC domain dipertahankan; penandatangan JWT dan klaim dibuang | M |
| NATS JetStream -> outbox Postgres + konsumen idempoten (D-HRS-05) | 01, 02, 03, 04, 05, 06 | Definisi peristiwa domain dan `processed_events` dipertahankan; `nats_*` ditulis ulang; payload peristiwa dilengkapi | M |
| golang-migrate -> goose v3.26.0 tertanam, `cmd/migrate`, CI up->down->up | semua skema | Isi DDL dipertahankan setelah diperbaiki (absensi 001 tidak valid); format file diubah | M |
| sqlc (hanya auth-service) + pgx -> pgx SQL mentah | 01 AUTH | Antarmuka repositori di domain dipertahankan; kode sqlc dibuang | S |
| zap -> `log/slog` JSON | semua | Pemanggilan log dipetakan; `services/_shared/observability/logging.go` dibuang | S |
| MinIO -> RustFS lewat `objstore.Store` (presign-only, ClamAV fail-closed) | 07 DOC | Kebutuhan bisnis dokumen dipertahankan; `storage/minio_client.go`, `presigner.go` ditulis ulang | S |
| Redis 7 -> Dragonfly (go-redis/v9) | 01 AUTH, SEC | Antarmuka token store dipertahankan | S |
| Layanan mikro (7 proses, 7 port gRPC) -> modul `services/<name>` dalam satu proses `apps/api` | semua | Seluruh domain/application; `main.go` per layanan diganti satu `cmd/api`; gRPC antarmodul diganti panggilan antarmuka Go | L |
| RLS per layanan lewat `postgres/rls.go` -> RLS ENABLE+FORCE+`WITH CHECK` di semua tabel bertenant via `kit/tenantdb.WithTenant`; peran NOBYPASSRLS (D8) | semua skema | Kebijakan RLS yang ada diperbaiki (FORCE, `WITH CHECK`); `rls.go` diganti | M |
| OTel/Prometheus (`_shared/observability`, tidak pernah diinisialisasi) -> baseline `slog` + `/healthz`; OTel opsional (D11, Q-HRS-08) | semua | Tidak ada yang dipertahankan; kode OTel tidak aktif sekarang | S |
| Kredensial dev tertulis di berkas terlacak -> `.env.example` + validasi gagal-cepat | deploy | `deploy/docker/*` ditulis ulang tanpa rahasia | S |
| Kolom NIK/NPWP/birth_date `TEXT` polos -> `kit/piicrypt` (AES-256-GCM + indeks buta) | 02 EMP | Skema kolom tetap, ditambah kolom blind index; data lama dienkripsi saat migrasi | M |
| Cuti/kehadiran tanpa `employee_number`, enum UPPERCASE vs lowercase, kolom `description` tidak ada -> perbaikan skema dan SQL | 02, 03, 04 | Domain dipertahankan; SQL repositori ditulis ulang | M |
| Dokumen: RPC presign/rollback/download kosong -> implementasi nyata + uji | 07 DOC | Domain `File`/`FileVersion` dipertahankan | M |
| Mode `suite`: adaptor port `Identity`, `Tenant`, `Entitlement`, `Notification`, `Audit`, `Events` + manifest produk | 01, 05, 06, SEC | Antarmuka diletakkan di application; adaptor baru | M |

---

## 3. Aktor dan Model Hak Akses

### 3.1 Aktor

| Aktor | Deskripsi | Status as-is | Bukti |
|---|---|---|---|
| Admin HR (`hr_admin`) | mengelola data induk, melihat audit, menyetujui cuti | Sebagian: peran ada, penegakan tidak ada | `services/auth-service/internal/domain/permission.go` |
| Manajer (`manager`) | melihat karyawan dan kehadiran bawahan, menyetujui cuti | Sebagian: tanpa hierarki atasan-bawahan yang ditegakkan | `services/auth-service/internal/domain/permission.go` |
| Karyawan (`employee`) | mengajukan cuti, melihat data sendiri | Sebagian | `services/auth-service/internal/domain/permission.go` |
| Auditor | membaca jejak audit | Belum: tidak ada peran bawaan; hanya `audit:read` milik `hr_admin` | `services/auth-service/internal/domain/permission.go` |
| Admin platform (suite) | mengelola org/tenant, entitlement | Belum (ada di platform `stery-suite`) | `proto/hris/auth/v1/auth.proto` |
| Sistem/worker | konsumen peristiwa, job terjadwal | Sebagian | `services/leave-service/internal/infrastructure/nats/employee_created_consumer.go` |

### 3.2 Izin bawaan dan matriks peran (as-is)

Sebelas izin bawaan dibuat oleh `RegisterTenant`: `employee:read|write|delete`, `attendance:read|write`, `leave:read|write|approve`, `notification:read`, `audit:read`, `settings:write`. Tiga peran sistem dibuat per tenant.

| Izin | `hr_admin` | `manager` | `employee` |
|---|---|---|---|
| `employee:read` | ya | ya | ya |
| `employee:write` | ya | - | - |
| `employee:delete` | ya | - | - |
| `attendance:read` | ya | ya | - |
| `attendance:write` | ya | - | - |
| `leave:read` | ya | ya | ya |
| `leave:write` | - | - | ya |
| `leave:approve` | ya | ya | - |
| `notification:read` | ya | ya | ya |
| `audit:read` | ya | - | - |
| `settings:write` | ya | - | - |

Temuan (Bukti `services/auth-service/internal/domain/permission.go`, baris 33-58):
- `hr_admin` tidak memiliki `leave:write` (tidak dapat mengajukan cuti atas nama sendiri bila perannya hanya itu).
- `employee` tidak memiliki `attendance:*`: dengan RBAC bawaan karyawan tidak dapat melakukan check-in. Ini cacat desain yang harus dikoreksi (FR-HRS-03.1, FR-HRS-01.6).
- Tidak ada izin untuk dokumen dan AI (`document:*`, `ai:*`); penambahan ada pada FR-HRS-07.* dan FR-HRS-08.*.
- Persetujuan cuti di leave-service membaca header `x-user-roles` yang dikirim pemanggil (`services/leave-service/internal/interfaces/grpc/context_guards.go`, baris 92-140); header tersebut dapat dipalsukan karena tidak ada interseptor autentikasi.

### 3.3 Model hak akses target

- Tenant dan aktor selalu berasal dari sesi (cookie buram), tidak pernah dari badan permintaan (FR-HRS-SEC.4, SEC.5).
- Tabel RBAC (roles, permissions, actions, role_actions) + `RequirePermission(code)` (STACK-SPEC 9). Izin target menambah `document:read|write|delete`, `ai:use`, `user:manage`.
- Cakupan data: `employee` hanya baris sendiri (`employee_id` dari sesi); `manager` tim langsung berdasarkan `manager_id`; `hr_admin` seluruh tenant. Aturan per modul ada di bagian 4.
- Deny-by-default: rute/RPC tanpa deklarasi izin ditolak (FR-HRS-SEC.5).

---

## 4. Spesifikasi Fungsional per Modul

Tiap modul menyatakan: ringkasan kondisi saat ini, tabel FR (`ID | Requirement | Prioritas | Status | Bukti`), lalu aturan bisnis, mesin status (bila ada), dan peristiwa yang dihasilkan/dikonsumsi. Kolom Requirement menyatakan perilaku **target**; kolom Status dan Bukti menyatakan kondisi **saat ini**.

### 4.01 AUTH: identitas, registrasi tenant, RBAC

Layanan `auth-service` memiliki enam RPC dan keenamnya memiliki handler berisi (Bukti `services/auth-service/internal/interfaces/grpc/auth_service.go`). Skema `auth`: `tenants`, `users`, `roles`, `permissions`, `role_permissions`, `user_roles`, `login_attempts`, `refresh_tokens` (dua terakhir tidak dipakai kode), `processed_events`. Kata sandi di-hash bcrypt cost bawaan (10).

| ID | Requirement | Prioritas | Status | Bukti |
|---|---|---|---|---|
| FR-HRS-01.1 | Tenant mendaftar sendiri (nama, slug unik, admin pertama). Dalam satu transaksi atomik dibuat tenant, tiga peran sistem, izin bawaan, dan pengguna admin berperan `hr_admin`; peristiwa `tenant_registered` dan `user_registered` dipancarkan lewat outbox | M | Sebagian: handler dan aturan ada, tetapi pembuatan tidak atomik dan INSERT/UPDATE `auth.tenants` hanya diberikan ke peran `hris_admin` sehingga gagal di bawah DSN bawaan `hris_app` (`Pending Verification`, belum dijalankan) | `services/auth-service/internal/application/commands/register_tenant.go` |
| FR-HRS-01.2 | Login dengan email + kata sandi; sukses menerbitkan sesi buram (cookie HttpOnly Secure) berisi tenant, pengguna, peran; kredensial salah tidak membedakan "email tidak ada" dan "sandi salah" | M | Sebagian: handler `Login` ada dan menerbitkan JWT; aksi login web memanggil API Go dan menyimpan token di cookie HttpOnly `hris_session` (tiruan ke `localhost:3000` sudah teratasi); sesi berupa JWT, bukan sesi buram | `apps/api/services/auth/application/commands/login.go`, `apps/web/src/routes/(auth)/login/+page.server.ts` |
| FR-HRS-01.3 | Kebijakan kata sandi: 8-128 karakter dengan huruf besar, kecil, angka, simbol; hash bcrypt (skema hash dipertahankan, D-HRS-06 dan D10); slug tenant 3-50 karakter `^[a-z0-9-]+$` | M | Selesai | `services/auth-service/internal/domain/user.go` |
| FR-HRS-01.4 | Siklus hidup sesi: terbit, validasi (introspeksi), perpanjang, cabut/logout; kedaluwarsa dengan TTL `tbd` (Q-HRS-11); hanya hash sesi disimpan di Dragonfly. As-is: JWT RS256 15 menit + refresh 7 hari (hardcoded), refresh dan daftar cabut di Redis | M | Sebagian: refresh dan cabut ada tetapi tanpa deteksi pemakaian ulang keluarga token; TTL hardcoded di `cmd/server/main.go` baris 81 | `services/auth-service/internal/application/commands/refresh_token.go`, `services/auth-service/internal/application/commands/revoke_token.go`, `services/auth-service/internal/infrastructure/jwt_signer.go` |
| FR-HRS-01.5 | Penguncian akun setelah percobaan gagal berulang (ambang dan durasi `tbd`) dan pembatasan laju login per alamat dan per akun | M | Belum: tabel `login_attempts` ada di skema tetapi tidak dipakai; tidak ada kode lockout | `services/auth-service/migrations/001_create_auth_schema.up.sql` |
| FR-HRS-01.6 | RBAC: peran, izin, aksi; `RequirePermission(code)` di setiap rute; izin bawaan dikoreksi (karyawan memperoleh `attendance:write` dan `attendance:read` miliknya; `hr_admin` memperoleh `leave:write`; izin dokumen dan AI ditambahkan) | M | Sebagian: 11 izin dan 3 peran dibuat, tetapi tidak ada yang menegakkan; kekeliruan matriks pada 3.2 | `services/auth-service/internal/domain/permission.go` |
| FR-HRS-01.7 | Manajemen pengguna oleh admin: buat, nonaktifkan, ganti peran, reset akses; pengguna terhubung ke karyawan (`employee_id`) | S | Belum: proto auth tidak memiliki RPC pengguna | `proto/hris/auth/v1/auth.proto` |
| FR-HRS-01.8 | Verifikasi email dan lupa/reset kata sandi lewat tautan sekali pakai berumur pendek | S | Belum: tidak ada RPC maupun peristiwa `password_changed` yang dipancarkan | `services/auth-service/internal/domain/user.go` |
| FR-HRS-01.9 | Alamat IP klien dan agen pengguna tercatat pada peristiwa login/audit | S | Sebagian: `extractIPAddr` selalu mengembalikan `0.0.0.0` (baris 196-199) | `services/auth-service/internal/interfaces/grpc/auth_service.go` |
| FR-HRS-01.10 | Mode `suite`: identitas didelegasikan ke platform lewat port `Identity` (adaptor `suite` memanggil `SessionService/Introspect`); mode `standalone` tidak berubah (D-HRS-02, D-HRS-03) | M | Belum | `proto/hris/auth/v1/auth.proto` |

**Aturan bisnis**
- Satu email unik per tenant (target); slug tenant unik global.
- Pengguna `deactivated` tidak dapat login dan semua sesinya dicabut.
- Peran sistem tidak dapat dihapus; peran kustom `tbd` (Q-HRS-06 mencakup kebijakan, belum ada keputusan).
- Tenant dan aktor pada operasi tidak pernah diambil dari badan permintaan (as-is: diambil dari badan; lihat 8.2).

**Mesin status sesi (target)**

| Dari | Peristiwa | Ke |
|---|---|---|
| (tidak ada) | login sukses | `active` |
| `active` | perpanjangan (sliding, TTL `tbd`) | `active` |
| `active` | logout / pencabutan admin / pengguna dinonaktifkan | `revoked` |
| `active` | TTL habis | `expired` |

**Peristiwa.** Dihasilkan: `identity.tenant_registered`, `identity.user_registered`, `user.login_succeeded`, `user.login_failed` (as-is `hris.identity.*`, 2.5). Dikonsumsi: tidak ada.

### 4.02 EMP: karyawan, departemen, jabatan, bagan organisasi

Layanan `employee-service` memiliki 10 RPC; 8 memiliki handler, `UpdateEmployee` dan `GetOrgChart` tidak ada di `internal/interfaces/grpc/employee_service.go`. Skema `employee`: `departments`, `positions`, `employees`, `processed_events`; enum PG huruf kecil (`employment_status`, `contract_type`) sementara domain Go menulis huruf besar. Kolom sensitif: `national_id` (NIK), `tax_id` (NPWP), `birth_date` (baris 87-91 migrasi) bertipe polos.

| ID | Requirement | Prioritas | Status | Bukti |
|---|---|---|---|---|
| FR-HRS-02.1 | Membuat karyawan: `employee_number` otomatis atau manual unik per tenant, data pribadi, departemen, jabatan, `manager_id`, jenis kontrak, tanggal masuk; menolak melampaui `hrs.employees.max` (FR-HRS-02.11) | M | Sebagian: handler ada, tetapi INSERT tidak mengisi `employee_number` (NOT NULL), domain tidak memiliki `employeeNumber`, `ContractType` diabaikan, dan enum domain huruf besar ditulis ke enum PG huruf kecil (`Pending Verification`, belum dijalankan) | `services/employee-service/internal/application/commands/create_employee.go`, `services/employee-service/internal/infrastructure/postgres/employee_repository.go` |
| FR-HRS-02.2 | Daftar dan detail karyawan dengan filter (status, departemen, pencarian nama, nomor), paginasi dengan total benar | M | Sebagian: `ListEmployees` mengembalikan Total = panjang halaman dan offset selalu 0; `GetEmployee` ada | `services/employee-service/internal/application/queries/list_employees.go`, `services/employee-service/internal/application/queries/get_employee.go` |
| FR-HRS-02.3 | Mengubah data karyawan (data pribadi, departemen, jabatan, manajer, status) dengan jejak audit dan peristiwa `employee.updated` | M | Belum: RPC `UpdateEmployee` (PATCH `/v1/employees/{id}`) tidak memiliki handler | `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| FR-HRS-02.4 | Terminasi karyawan (tanggal, alasan) menutup status, menonaktifkan pengguna terkait, memicu penutupan saldo/kehadiran, dan memancarkan `employee.terminated` | M | Sebagian: terminasi dan peristiwa ada; tanpa kaskade ke pengguna/cuti/kehadiran | `services/employee-service/internal/application/commands/terminate_employee.go` |
| FR-HRS-02.5 | Departemen berhierarki (`parent_id`), kepala departemen, daftar | M | Sebagian: buat dan daftar ada; `parent_id` tidak dipakai; cache dalam proses TTL 24 jam | `services/employee-service/internal/application/commands/create_department.go`, `services/employee-service/internal/infrastructure/cache/department_cache.go` |
| FR-HRS-02.6 | Jabatan dengan level (JUNIOR/SENIOR/LEAD/MANAGER, selaras dengan domain), daftar | M | Sebagian: INSERT/SELECT memakai kolom `description` yang tidak ada di skema; default kolom `level` `'staff'` tidak termasuk enum domain (`Pending Verification`) | `services/employee-service/internal/infrastructure/postgres/position_repository.go`, `services/employee-service/internal/domain/position.go` |
| FR-HRS-02.7 | Bagan organisasi (pohon departemen dan manajer-bawahan) untuk tenant, termasuk filter departemen | M | Belum: RPC `GetOrgChart` (GET `/v1/employees/org-chart`) tanpa handler | `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| FR-HRS-02.8 | NIK, NPWP, tanggal lahir dienkripsi dengan `kit/piicrypt` (AES-256-GCM, indeks buta untuk pencarian NIK/nomor); tampilan dimasking untuk peran tanpa kebutuhan; akses tercatat di audit | M | Belum: kolom `TEXT` polos (baris 89-91) | `services/employee-service/migrations/001_create_employee_schema.up.sql` |
| FR-HRS-02.9 | Impor karyawan massal (CSV) dengan pratinjau galat dan batas kuota | C | Belum | `proto/hris/employee/v1/employee.proto` |
| FR-HRS-02.10 | Peristiwa `user_registered` membuat kerangka karyawan untuk admin pertama tenant | S | Sebagian: konsumen ada tetapi tidak dapat bekerja (payload tanpa `tenant_id`, id departemen/jabatan acak, galat di-ack lalu hilang) | `services/employee-service/internal/infrastructure/nats_consumer.go` |
| FR-HRS-02.11 | Batas `hrs.employees.max` per entitlement (mode suite) atau konfigurasi tenant (standalone); nilai `tbd` (Q-HRS-02); pelampauan hanya menolak pembuatan karyawan baru, data lama tetap | M | Belum | `services/employee-service/internal/application/commands/create_employee.go` |

**Aturan bisnis**
- `employee_number` unik per tenant (`UNIQUE (tenant_id, employee_number)` sudah ada di skema).
- Karyawan terminasi tidak dihapus fisik; status `terminated` dengan `termination_date`.
- `manager_id` merujuk karyawan lain di tenant yang sama (FK diri sendiri, `ON DELETE SET NULL`); siklus manajer ditolak (target).
- Penghapusan data pribadi mengikuti kebijakan retensi UU PDP (8.3).

**Mesin status karyawan** (enum PG `employment_status`: `active`, `inactive`, `terminated`, `on_leave`; domain Go memakai nilai huruf besar, `services/employee-service/internal/domain/employee.go`):

| Dari | Peristiwa | Ke |
|---|---|---|
| (baru) | buat | `active` |
| `active` | nonaktifkan | `inactive` |
| `inactive` | aktifkan | `active` |
| `active` | cuti panjang `tbd` (kebijakan Q-HRS-06) | `on_leave` |
| `active`, `inactive`, `on_leave` | terminasi | `terminated` (final) |

**Peristiwa.** Dihasilkan: `employee.created`, `employee.terminated`, `employee.updated` (belum dipancarkan). Dikonsumsi: `identity.user_registered` (FR-HRS-02.10).

### 4.03 ATT: kehadiran

Layanan `attendance-service` memiliki lima RPC dengan handler. Skema `attendance` memiliki satu tabel `attendance_records` (kolom: `employee_id`, `date`, `check_in_at`, `check_out_at`, koordinat, `status`, `work_duration_mins`, `notes`, `created_by`, `is_override`). **Migrasi 001 tidak valid di PostgreSQL** (baris 25 `CONSTRAINT ... UNIQUE (...) WHERE NOT is_override` bukan sintaks sah; baris 45-46 `GRANT ... TO app` memakai peran `app`, bukan `hris_app`); akibatnya layanan tidak dapat berjalan pada DB segar sebelum migrasi diperbaiki.

| ID | Requirement | Prioritas | Status | Bukti |
|---|---|---|---|---|
| FR-HRS-03.1 | Check-in harian oleh karyawan: satu catatan aktif per karyawan per hari; menyimpan waktu server (UTC) dan koordinat opsional; ditolak bila sudah check-in | M | Sebagian: command ada; bergantung pada migrasi 001 yang tidak valid; RBAC bawaan karyawan tidak memiliki `attendance:*` | `services/attendance-service/internal/application/commands/check_in.go`, `services/attendance-service/migrations/001_create_attendance_schema.up.sql` |
| FR-HRS-03.2 | Check-out: memerlukan check-in aktif; menghitung `work_duration_mins`; waktu keluar setelah masuk (constraint `attendance_check_in_before_out`) | M | Sebagian | `services/attendance-service/internal/application/commands/check_out.go` |
| FR-HRS-03.3 | Koreksi oleh HR/atasan dengan alasan wajib dan jejak audit (`is_override`, `created_by`); dapat membuat catatan baru bila belum ada | M | Sebagian: override gagal untuk catatan baru | `services/attendance-service/internal/application/commands/override_attendance.go` |
| FR-HRS-03.4 | Daftar kehadiran dengan filter rentang tanggal, karyawan, status; paginasi | M | Sebagian: `ListAttendance` mengabaikan filter tanggal | `services/attendance-service/internal/interfaces/grpc/attendance_service.go` |
| FR-HRS-03.5 | Ringkasan kehadiran per karyawan per periode (hadir, terlambat, absen, cuti, total jam) dan ringkasan harian untuk HR | M | Sebagian: `GetAttendanceSummary` mengabaikan `employee_id` dan `period`; ringkasan harian ada | `services/attendance-service/internal/application/queries/get_daily_summary.go` |
| FR-HRS-03.6 | Aturan kerja per tenant: jam kerja, ambang terlambat, hari kerja, geofence opsional; status `LATE`/`HALF_DAY` dihitung otomatis; nilai bawaan `tbd` (Q-HRS-05) | S | Belum: status `LATE` dan `HALF_DAY` ada di enum tetapi tidak pernah dihitung | `services/attendance-service/internal/domain/attendance_record.go` |
| FR-HRS-03.7 | Cuti yang disetujui menandai hari terkait `ON_LEAVE` (konsumen peristiwa `leave.approved`) | S | Belum | `services/attendance-service/internal/domain/attendance_record.go` |
| FR-HRS-03.8 | Tanggal dan batas hari dihitung pada zona waktu tenant (bawaan WIB); waktu disimpan UTC | M | Belum: kolom `date` tanpa zona; zona tenant tidak ada | `services/attendance-service/migrations/001_create_attendance_schema.up.sql` |

**Aturan bisnis**
- Unik satu catatan non-override per `(tenant_id, employee_id, date)`; di PostgreSQL dinyatakan sebagai indeks unik parsial (`CREATE UNIQUE INDEX ... WHERE NOT is_override`), bukan `CONSTRAINT`.
- Check-out tidak boleh sebelum check-in. Koreksi tidak menimpa data asli tanpa jejak audit.
- Migrasi 002 membuat indeks `idx_attendance_employee_date` tanpa `IF NOT EXISTS`, nama yang sama dengan indeks di migrasi 001 (baris 29): migrasi 002 diperkirakan gagal dengan galat "relation already exists" bahkan setelah 001 diperbaiki (`Pending Verification`, belum dijalankan). Keduanya harus diperbaiki bersama.

**Mesin status kehadiran harian**

| Dari | Peristiwa | Ke |
|---|---|---|
| (tidak ada) | check-in | `PRESENT` (atau `LATE` bila lewat ambang, target) |
| `PRESENT` | check-out dengan durasi di bawah batas hari penuh | `HALF_DAY` (target) |
| (tidak ada) | cuti disetujui | `ON_LEAVE` (target) |
| (tidak ada) | akhir hari tanpa check-in | `ABSENT` (job penutup hari, target) |

**Peristiwa.** Dihasilkan: `attendance.checked_in`, `checked_out`, `overridden` (as-is subjek `hris.attendance.record.*`, tidak cocok dengan konsumen audit; payload kosong). Dikonsumsi: `leave.approved` (target, FR-HRS-03.7).

### 4.04 LEV: cuti

Layanan `leave-service` memiliki delapan RPC dengan handler. Skema `leave`: `leave_types` (kode, nama, `max_days_per_year`, `requires_document`, `is_paid`), `leave_requests` (status `PENDING|APPROVED|REJECTED|CANCELLED`, `days_count INTEGER CHECK (> 0)`, `approved_by_id`, `rejection_reason`), `leave_balances` (`entitled_days`, `used_days`, `pending_days`, unik per tahun), `processed_events`. Kebijakan RLS memakai `current_setting('app.tenant_id')` tanpa `missing_ok`, tanpa `WITH CHECK` dan tanpa `FORCE`; migrasi memberi hak ke peran `app`.

| ID | Requirement | Prioritas | Status | Bukti |
|---|---|---|---|---|
| FR-HRS-04.1 | Jenis cuti per tenant dapat dibuat/diubah/dinonaktifkan oleh HR; seed bawaan sesuai kebijakan `tbd` (Q-HRS-06) | M | Sebagian: `ListLeaveTypes` ada; tidak ada pembuat jenis cuti dan tidak ada seed, sehingga daftar kosong dan tidak ada saldo | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| FR-HRS-04.2 | Saldo tahunan per karyawan per jenis dibuat saat `employee.created` (idempoten) dan dapat disesuaikan HR | M | Sebagian: konsumen ada tetapi `EmployeeValidator` di `main.go` adalah no-op dan tanpa jenis cuti tidak ada saldo yang terbentuk | `services/leave-service/internal/infrastructure/nats/employee_created_consumer.go` |
| FR-HRS-04.3 | Mengajukan cuti: jenis, rentang tanggal, alasan; sistem menghitung `days_count` hari kerja; menolak saldo kurang, tanggal tumpang tindih dengan pengajuan aktif, dan rentang terbalik | M | Sebagian: saldo ditegakkan di kode aplikasi (tanpa kunci baris); tumpang tindih tidak dicek; `ApplyLeave` tidak mengisi `DaysCount` sehingga INSERT berpotensi melanggar `CHECK (days_count > 0)` (`Pending Verification`) | `services/leave-service/internal/application/commands/create_leave_request.go`, `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| FR-HRS-04.4 | Setujui atau tolak oleh atasan langsung atau HR dengan alasan penolakan; aktor persetujuan diambil dari sesi; pemohon tidak boleh menyetujui pengajuannya sendiri; keputusan hanya dari `PENDING` | M | Sebagian: handler ada; `approver_id` diambil dari badan permintaan, tanpa pencegahan persetujuan diri, peran dibaca dari header `x-user-roles` | `services/leave-service/internal/interfaces/grpc/context_guards.go`, `services/leave-service/internal/domain/leave_request.go` |
| FR-HRS-04.5 | Pembatalan oleh pemohon (dan HR) memulihkan `pending_days`/`used_days`; aturan status mana yang boleh dibatalkan ditetapkan di domain | M | Sebagian: `Cancel` mengizinkan `PENDING` padahal `ErrCannotCancelPending` dideklarasikan (`leave_request.go` baris 159-172); pemulihan saldo belum terverifikasi | `services/leave-service/internal/domain/leave_request.go`, `services/leave-service/internal/domain/errors.go` |
| FR-HRS-04.6 | Daftar dan detail pengajuan dengan filter status/karyawan/periode; karyawan hanya melihat miliknya, manajer timnya, HR semuanya | M | Sebagian: filter status dikirim sebagai `LEAVE_STATUS_*` yang tidak cocok dengan nilai DB (`PENDING|...`) | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| FR-HRS-04.7 | Melihat saldo per jenis untuk tahun berjalan | M | Sebagian: `GetLeaveBalance` ada; pengajuan yang melintasi tahun dihitung seluruhnya pada tahun mulai | `services/leave-service/internal/domain/leave_balance.go` |
| FR-HRS-04.8 | Kalender hari libur nasional/tenant yang dikecualikan dari hitungan hari kerja | C | Belum | `services/leave-service/migrations/001_create_leave_schema.up.sql` |
| FR-HRS-04.9 | Penyesuaian saldo manual oleh HR (tambah/kurang) dengan alasan dan audit; carry-over tahunan sesuai kebijakan `tbd` | C | Belum | `services/leave-service/internal/domain/leave_balance.go` |

**Aturan bisnis**
- Saldo tersedia = `entitled_days - used_days - pending_days` dan tidak boleh negatif (CHECK DB sudah ada pada tiga kolom).
- Pengajuan `PENDING` menambah `pending_days`; `APPROVED` memindahkan ke `used_days`; `REJECTED`/`CANCELLED` memulihkan.
- Pengajuan yang berakhir sebelum mulai ditolak; jenis `requires_document` memerlukan lampiran dokumen (target, terkait 4.07).

**Mesin status cuti**

| Dari | Peristiwa | Ke | Efek saldo |
|---|---|---|---|
| (baru) | ajukan | `PENDING` | `pending_days += days_count` |
| `PENDING` | setujui | `APPROVED` | `pending_days -= n; used_days += n` |
| `PENDING` | tolak | `REJECTED` | `pending_days -= n` |
| `PENDING` | batalkan | `CANCELLED` | `pending_days -= n` |
| `APPROVED` | batalkan (sebelum tanggal mulai, target) | `CANCELLED` | `used_days -= n` |

**Peristiwa.** Dihasilkan: `leave.requested`, `approved`, `rejected`, `cancelled` (payload as-is kosong). Dikonsumsi: `employee.created` (membuat saldo), `employee.terminated` (target: menutup saldo).

### 4.05 NTF: notifikasi

Layanan `notification-service` memiliki lima RPC dengan handler. Skema `notification`: `notifications` (kanal `EMAIL|IN_APP`, `template_key`, `variables` JSONB, status `PENDING|SENT|FAILED|READ`), `notification_channel_configs` (per tenant: `channels TEXT[]`, `email_config` JSONB), `processed_events`. Adaptor email dan in-app ada di kode tetapi **tidak dipasang**: `cmd/server/main.go` baris 112-132 mengisi `_ = emailAdapter` dan `_ = inAppAdapter`; enam templat HTML ada di `templates/` tetapi tidak pernah dimuat. Migrasi memberi hak ke peran `app`; RLS ENABLE tanpa FORCE, `WITH CHECK` ada lewat kebijakan terpisah.

| ID | Requirement | Prioritas | Status | Bukti |
|---|---|---|---|---|
| FR-HRS-05.1 | Notifikasi tersimpan per penerima (`recipient_id`) dengan kanal, templat, variabel, status, dan waktu kirim/baca; dibuat dalam transaksi yang sama dengan peristiwa pemicunya (outbox) | M | Sebagian: tabel dan repositori ada; penyimpanan berjalan lewat konsumen/handler | `services/notification-service/migrations/001_create_notification_schema.up.sql` |
| FR-HRS-05.2 | Pengiriman email lewat SMTP dengan templat berlokal `{id,en}`, percobaan ulang terbatas, dan pencatatan `SENT`/`FAILED` | M | Sebagian: adaptor email ada tetapi tidak dipasang; templat tidak dimuat; `MarkSent`/`MarkFailed` tidak punya pemanggil sehingga status tidak pernah berubah dari `PENDING` | `services/notification-service/internal/infrastructure/channels/email_adapter.go`, `services/notification-service/internal/infrastructure/template_engine.go` |
| FR-HRS-05.3 | Konsumsi peristiwa: pengajuan cuti (ke atasan/HR), keputusan cuti (ke pemohon), karyawan baru, pengguna terdaftar; konsumsi idempoten | M | Sebagian: konsumen leave (dengan uji), auth, dan employee ada; payload leave kosong sehingga templat tidak dapat diisi | `services/notification-service/internal/application/consumers/leave_event_consumer.go` |
| FR-HRS-05.4 | Daftar notifikasi pengguna (paginasi, jumlah belum dibaca) dan tanda sudah dibaca (satu/semua) | M | Sebagian: `MarkRead` hanya menerima status `SENT` sehingga selalu gagal diam-diam; `unread_count` selalu 0; `next_page_token` berisi literal `"next"` | `services/notification-service/internal/application/commands/mark_notification_read.go`, `services/notification-service/internal/interfaces/grpc/notification_service.go` |
| FR-HRS-05.5 | Konfigurasi kanal per tenant (kanal aktif, SMTP); rahasia SMTP disimpan terenkripsi, tidak dalam JSONB polos | S | Sebagian: `UpdateChannelConfig` menimpa daftar kanal; `smtp_password` tersimpan polos di JSONB | `services/notification-service/internal/interfaces/grpc/notification_service.go` |
| FR-HRS-05.6 | Kanal tambahan (WhatsApp, push) | C | Belum (penyedia `tbd`, Q-HRS-13) | `proto/hris/notification/v1/notification.proto` |
| FR-HRS-05.7 | Preferensi dan opt-out per pengguna per jenis notifikasi; notifikasi keamanan wajib tidak dapat dimatikan | C | Belum | `proto/hris/notification/v1/notification.proto` |

**Aturan bisnis**
- Kegagalan pengiriman tidak membatalkan transaksi bisnis; notifikasi masuk antrean dan dicoba ulang hingga batas `tbd`.
- Konten notifikasi tidak memuat PII kelas 2 (NIK, NPWP, tanggal lahir).
- Penerima ditentukan dari data karyawan/atasan; pada mode suite pengiriman lewat port `Notification` (hub platform).

**Mesin status notifikasi**

| Dari | Peristiwa | Ke |
|---|---|---|
| (baru) | dibuat | `PENDING` |
| `PENDING` | terkirim | `SENT` |
| `PENDING` | gagal permanen | `FAILED` |
| `SENT` atau (IN_APP) `PENDING` | dibaca | `READ` |

**Peristiwa.** Dikonsumsi: `identity.user_registered`, `employee.created|terminated`, `leave.requested|approved|rejected|cancelled`. Dihasilkan: `notification.*` (belum ada produsen; audit memiliki konsumen).

### 4.06 AUD: audit

Layanan `audit-service` memiliki tiga RPC dengan handler. Skema `audit`: `entries` (dipartisi bulanan 2023-01 s.d. 2027-12, tanpa partisi bawaan; kolom `actor_id`, `action`, `resource_type`, `resource_id UUID`, `changes JSONB`, `success`), `processed_events`. RLS tidak diaktifkan pada `entries` (komentar migrasi: layanan audit membaca semua tenant); peran `hris_app` hanya diberi INSERT.

| ID | Requirement | Prioritas | Status | Bukti |
|---|---|---|---|---|
| FR-HRS-06.1 | Mencatat entri audit untuk tindakan sensitif dari modul lain secara sinkron (port `Audit.Record`) maupun lewat peristiwa | M | Sebagian: konsumen identity/workforce/operations/notification ada; `resource_id` bertipe UUID sementara konsumen mengirim string kosong, sehingga INSERT diperkirakan gagal (`Pending Verification`) | `services/audit-service/internal/infrastructure/nats/identity_consumer.go`, `services/audit-service/internal/infrastructure/postgres/audit_repository.go` |
| FR-HRS-06.2 | Query entri dengan filter aktor, aksi, jenis/ID sumber, rentang waktu; paginasi; hanya peran `audit:read`; selalu terbatas pada tenant pemanggil | M | Pending Verification: `hris_app` hanya memiliki INSERT pada `entries`, sehingga `Query` diperkirakan ditolak izin (`[INFERENCE]`) | `services/audit-service/internal/interfaces/grpc/audit_service.go`, `services/audit-service/migrations/001_create_audit_schema.up.sql` |
| FR-HRS-06.3 | Mengambil satu entri berdasarkan ID | S | Sebagian: `GetEntry` memakai query dengan `Limit:1` sehingga tidak mengembalikan entri yang diminta | `services/audit-service/internal/interfaces/grpc/audit_service.go` |
| FR-HRS-06.4 | Entri tidak dapat diubah atau dihapus (UPDATE/DELETE/TRUNCATE ditolak di tingkat DB, trigger `kit/audit`) | M | Sebagian: append-only hanya lewat pembatasan hak (`hris_app` INSERT); tidak ada trigger penolak | `services/audit-service/migrations/001_create_audit_schema.up.sql` |
| FR-HRS-06.5 | Retensi entri sesuai kebijakan (masa `tbd`, Q-HRS-04) dengan pemeliharaan partisi otomatis | S | Sebagian: partisi statis sampai 2027-12; entri di luar rentang akan gagal | `services/audit-service/migrations/001_create_audit_schema.up.sql` |
| FR-HRS-06.6 | Cakupan minimum: login (sukses/gagal), perubahan karyawan, koreksi absensi, keputusan cuti, akses/unduh dokumen rahasia, perubahan PII, perubahan peran | M | Sebagian: subjek attendance tidak cocok dengan konsumen; jenis `login_*` tidak dipetakan; akses dokumen dan PII tidak dicatat | `services/audit-service/internal/infrastructure/nats/operations_consumer.go` |
| FR-HRS-06.7 | Isolasi tenant pada entri audit: RLS ENABLE+FORCE dengan `tenant_id`; akses lintas tenant hanya untuk peran operator platform | M | Belum: RLS dimatikan pada `entries` | `services/audit-service/migrations/001_create_audit_schema.up.sql` |

**Aturan bisnis**
- Entri berisi siapa (`actor_id`), apa (`action`, `resource_type`, `resource_id`), kapan (`created_at`), hasil (`success`, `error_message`), dan perubahan sebelum/sesudah (`changes`) tanpa nilai PII kelas 2 (hanya penanda "diubah").
- Penulisan audit tidak boleh membatalkan transaksi bisnis sensitif yang sudah berhasil; kegagalan penulisan tercatat dan dicoba ulang (outbox).

**Peristiwa.** Dikonsumsi: seluruh `identity.*`, `workforce.*` (employee), `operations.*` (leave/attendance), `notification.*`. Dihasilkan: tidak ada.

### 4.07 DOC: dokumen karyawan

Layanan `document-service` memiliki tujuh RPC; empat memiliki handler berisi (`UploadDocument`, `ListDocuments`, `GetDocumentMetadata`, `DeleteDocument`), tiga kosong (`DownloadDocument`, `GetPresignedUrl`, `RollbackDocumentVersion`; `document_service.go` baris 215-221 untuk dua terakhir). Tidak ada berkas uji di layanan ini. Skema `document`: `files` (kolom `minio_bucket`, `minio_key`, `status`, `classification` bawaan `'PUBLIC'`, `retention_days`, `current_version`, `total_versions`), `file_versions` (tanpa `tenant_id` dan tanpa RLS), `processed_events`. `UploadDocument` selalu mengembalikan `Unauthenticated` karena kunci konteks `actor_id` tidak pernah diisi di mana pun (`internal/interfaces/grpc/mapping.go` baris 13 membacanya).

| ID | Requirement | Prioritas | Status | Bukti |
|---|---|---|---|---|
| FR-HRS-07.1 | Unggah dua langkah: klien meminta URL presign (batas ukuran/MIME divalidasi), mengunggah langsung ke RustFS, lalu konfirmasi; metadata dibuat dalam transaksi bersama audit | M | Sebagian: `UploadDocument` ada tetapi selalu `Unauthenticated`; `GetPresignedUrl` kosong; bucket (`<tenantID>-documents`) tidak pernah dibuat | `services/document-service/internal/application/commands/upload_document.go`, `services/document-service/internal/interfaces/grpc/document_service.go` |
| FR-HRS-07.2 | Pemindaian malware (ClamAV `clamd`) fail-closed: berkas berstatus `SCANNING` tidak dapat diunduh sampai `CLEAN`; `SCAN_FAILED` ditolak | M | Belum: enum status `SCANNED`/`SCAN_FAILED` ada tanpa pemindai | `services/document-service/internal/domain`, `services/document-service/migrations/001_create_document_schema.up.sql` |
| FR-HRS-07.3 | Unduh hanya lewat URL presign berumur pendek (TTL `tbd`; as-is 3600 detik hardcoded) dengan pemeriksaan izin dan klasifikasi serta catatan audit untuk dokumen rahasia | M | Belum: `DownloadDocument` placeholder | `services/document-service/internal/interfaces/grpc/document_service.go` |
| FR-HRS-07.4 | Daftar dan metadata dokumen per entitas (`entity_type`, `entity_id`), tidak menampilkan dokumen terhapus | M | Sebagian: list/metadata ada; dokumen yang dihapus tidak disaring | `services/document-service/internal/interfaces/grpc/document_service.go` |
| FR-HRS-07.5 | Hapus dokumen (soft-delete, audit) dan penghapusan fisik menurut retensi (`retention_days`) | S | Sebagian: `DeleteDocument` ada; tidak ada pekerja retensi | `services/document-service/internal/application/commands/delete_document.go` |
| FR-HRS-07.6 | Versi dokumen: unggah versi baru dan pulihkan versi lama (`RollbackDocumentVersion`) dengan uji | S | Belum: `CreateDocumentVersion` tidak dikabelkan ke RPC mana pun; rollback kosong | `services/document-service/internal/application/commands/create_document_version.go` |
| FR-HRS-07.7 | Klasifikasi dokumen (`PUBLIC`, `INTERNAL`, `CONFIDENTIAL`, `tbd` final) menentukan izin akses dan pencatatan audit; dokumen karyawan (kontrak, identitas) bawaan `CONFIDENTIAL` | S | Sebagian: kolom `classification` ada dengan bawaan `PUBLIC`; tidak ditegakkan | `services/document-service/migrations/001_create_document_schema.up.sql` |
| FR-HRS-07.8 | Batas ukuran, tipe MIME yang diizinkan, dan kuota penyimpanan per tenant (nilai `tbd`) | S | Belum: tidak ada validasi ukuran/MIME | `services/document-service/internal/application/commands/upload_document.go` |

**Aturan bisnis**
- Satu bucket dan satu kunci akses untuk HRIS di RustFS (D-HRS-06); objek diberi awalan `tenant_id/` dan tidak pernah publik.
- Aksi dokumen `CONFIDENTIAL` selalu diaudit (FR-HRS-06.6). Menghapus karyawan tidak otomatis menghapus dokumen; retensi menentukan.

**Mesin status dokumen**

| Dari | Peristiwa | Ke |
|---|---|---|
| (baru) | permintaan presign | `PENDING_UPLOAD` (target) |
| `PENDING_UPLOAD` | unggah dikonfirmasi | `SCANNING` |
| `SCANNING` | bersih | `CLEAN` (dapat diunduh) |
| `SCANNING` | terinfeksi/gagal | `SCAN_FAILED` (ditolak) |
| `CLEAN` | hapus | `DELETED` |

Nilai status as-is: `UPLOADED` (bawaan kolom), `SCANNED`, `SCAN_FAILED`; nama target di atas `[INFERENCE]` dan disesuaikan saat implementasi.

**Peristiwa.** Dihasilkan: `document.uploaded|deleted` (target). Dikonsumsi: `employee.terminated` (target, memicu penjadwalan retensi).

### 4.08 AI: asisten AI

`ai-service` hanya memiliki kontrak `proto/hris/ai/v1/ai.proto` (tiga RPC: `AnalyzeResume`, `ForecastAttrition`, `Chat` streaming). Direktori `cmd/server/` kosong dan layanan dikomentari di `go.work`. Seluruh modul berstatus **Belum**. Pemrosesan PII ke penyedia LLM pihak ketiga memerlukan kebijakan terlebih dahulu (BR-HRS-11, RISK-HRS-12).

| ID | Requirement | Prioritas | Status | Bukti |
|---|---|---|---|---|
| FR-HRS-08.1 | Analisis resume (pencocokan kandidat ke deskripsi jabatan) | C | Belum | `proto/hris/ai/v1/ai.proto` |
| FR-HRS-08.2 | Prediksi attrisi berbasis data kehadiran/cuti agregat (tanpa PII langsung) | C | Belum | `proto/hris/ai/v1/ai.proto` |
| FR-HRS-08.3 | Chat asisten HR (streaming) dengan penolakan topik di luar cakupan dan tanpa mengirim NIK/NPWP/tanggal lahir ke LLM | C | Belum | `proto/hris/ai/v1/ai.proto` |
| FR-HRS-08.4 | Kebijakan dan kuota: pencabutan persetujuan, kuota biaya per tenant (`tbd`), log pemakaian tanpa isi sensitif; fitur hanya aktif bila entitlement/tenant mengaktifkan | C | Belum | `go.work` |

**Aturan bisnis.** Tidak ada pengiriman PII kelas 2 ke pihak ketiga tanpa dasar pemrosesan dan perjanjian (Q-HRS-09, Q-HRS-14). Modul dimatikan secara bawaan.

**Peristiwa.** Dihasilkan/dikonsumsi: tidak ada (belum ditentukan).

### 4.09 PAY: payroll (Ditunda)

**Status: Ditunda** (ADR-0004, `docs/adr/0004-defer-payroll-to-phase2.md`). Tidak ada FR pada versi ini. Alasan: logika payroll melibatkan PPh21, BPJS, dan THR yang memiliki konsekuensi hukum dan keuangan; kesalahan perhitungan berdampak langsung pada karyawan dan perusahaan. Penundaan ke fase 2 adalah keputusan terdokumentasi (BR-HRS-13, D-HRS-10).

Prasyarat sebelum FR payroll dapat ditulis: (1) tinjauan hukum aturan PPh21, BPJS, THR yang berlaku (jadwal `tbd`, Q-HRS-10); (2) ADR aturan pajak (modul 10); (3) ADR imutabilitas `PayrollRun` (run yang sudah diterbitkan tidak berubah); (4) desain saga/penutupan periode; (5) tinjauan retensi dan UU PDP atas data gaji.

### 4.10 TAX: aturan pajak (Ditunda)

**Status: Ditunda** (ADR-0004). Tidak ada FR. Aturan pajak (tarif, PTKP, TER) harus berasal dari tinjauan hukum dan dipelihara sebagai data versi-berlaku, bukan konstanta kode. Prasyarat: tinjauan hukum (Q-HRS-10), ADR aturan pajak, dan strategi pembaruan regulasi.

### 4.11 REC: rekrutmen (Ditunda)

**Status: Ditunda** (ADR-0004, BR-HRS-14). Tidak ada FR. Prasyarat: desain domain kandidat dan alur lamaran, kebijakan retensi data kandidat (UU PDP), dan keputusan terkait modul AI (FR-HRS-08.1).

### 4.12 PRF: kinerja (Ditunda)

**Status: Ditunda** (ADR-0004, BR-HRS-14). Tidak ada FR. Prasyarat: desain siklus penilaian dan skala nilai, tinjauan hak akses data penilaian, dan kebijakan retensi.

---

### 4.SEC Persyaratan keamanan lintas modul

Persyaratan berikut berlaku lintas modul 4.01 sampai 4.08; kolom Requirement menyatakan perilaku **target**, Status dan Bukti menyatakan kondisi **saat ini**.

| ID | Requirement | Prioritas | Status | Bukti |
|---|---|---|---|---|
| FR-HRS-SEC.1 | Setiap tabel bertenant memakai RLS ENABLE+FORCE dengan `tenant_isolation` (`USING` dan `WITH CHECK`) dan `current_setting('app.tenant_id', true)::uuid`, diset per transaksi oleh `kit/tenantdb.WithTenant`; peran aplikasi NOBYPASSRLS; peran migrasi terpisah; `audit.entries` ikut | M | Sebagian: benar pada auth (users, roles, refresh_tokens), employee, document.files; belum pada attendance, leave, notification, audit | `services/_shared/postgres/rls.go`, `docs/adr/0002-postgres-rls-for-tenant-isolation.md` |
| FR-HRS-SEC.2 | Uji lintas tenant wajib di setiap modul (tenant A tidak dapat membaca/menulis data tenant B) dijalankan di CI terhadap Postgres nyata dengan peran aplikasi NOBYPASSRLS | M | Sebagian: berkas uji integrasi ada untuk auth dan employee; job `test-integration` CI hanya mencakup auth, employee, leave dan terhubung sebagai `hris_admin` yang dibuat sebagai `POSTGRES_USER` (bypass RLS, `[INFERENCE]`), sehingga kebijakan RLS tidak teruji; attendance, notification, audit, document tidak tercakup; kandungan uji lintas tenant belum diperiksa | `services/auth-service/internal/integration/auth_integration_test.go`, `services/employee-service/internal/integration/employee_integration_test.go`, `.github/workflows/ci-go.yml` |
| FR-HRS-SEC.3 | NIK, NPWP, tanggal lahir dienkripsi `kit/piicrypt` dengan indeks buta; data lama dienkripsi saat migrasi; kolom `TEXT` polos dihapus | M | Belum | `services/employee-service/migrations/001_create_employee_schema.up.sql` |
| FR-HRS-SEC.4 | Sesi buram: cookie HttpOnly Secure SameSite, hash di Dragonfly, tanpa JWT dan tanpa token di peramban; tenant dan aktor selalu dari sesi | M | Belum: JWT dan cookie `auth_user` terbaca klien | `services/auth-service/internal/infrastructure/jwt_signer.go`, `apps/web/src/lib/stores/auth.ts` |
| FR-HRS-SEC.5 | Deny-by-default: interseptor/middleware menolak setiap rute/RPC tanpa izin eksplisit; rute internal memakai token layanan HMAC | M | Belum | `services/_shared/server/grpc_interceptors.go` |
| FR-HRS-SEC.6 | Pembatasan laju login dan endpoint sensitif per alamat dan per akun dengan Dragonfly; ambang `tbd` | M | Belum | `apps/api/cmd/api/routes.go` |
| FR-HRS-SEC.7 | Tidak ada rahasia atau kredensial bawaan di kode maupun berkas terlacak; rahasia dari env dengan validasi gagal-cepat; `.env.example`; rahasia yang pernah dikomit dirotasi dan dipurge dari riwayat | M | Sebagian: konfigurasi dari env ada tetapi kredensial dev tertulis di berkas terlacak dan nilai bawaan ada di kode | `services/_shared/config/loader.go`, `deploy/docker/docker-compose.yml` |
| FR-HRS-SEC.8 | Hak subjek data: HR dapat mengekspor/mengoreksi/menghapus data satu karyawan atas permintaan, tercatat di audit; dasar pemrosesan dan persetujuan dicatat | S | Belum | `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| FR-HRS-SEC.9 | Header keamanan web (CSP, `X-Content-Type-Options`, HSTS, `Referrer-Policy`), CORS ketat, proteksi CSRF untuk cookie sesi | S | Belum | `apps/web/svelte.config.js` |
| FR-HRS-SEC.10 | Log `slog` tidak memuat PII kelas 2, kata sandi, token, atau kredensial; ID permintaan dikorelasikan | M | Belum: tidak ada kebijakan redaksi; `zap` dipakai | `services/_shared/observability/logging.go` |
| FR-HRS-SEC.11 | Retensi dan penghapusan per kelas data dijalankan oleh pekerja terjadwal dengan jejak audit; masa retensi `tbd` (Q-HRS-04) | S | Belum | `services/document-service/migrations/001_create_document_schema.up.sql` |

## 5. Model Data

Kondisi saat ini: satu skema Postgres per layanan di satu basis data `hris_db` (`deploy/docker/init-postgres.sql`) atau per layanan di `docker-compose.production.yml` (basis data per layanan tidak dibuat oleh skrip init; `Pending Verification`). Target: satu basis data `hris`, satu peran aplikasi `hris` NOBYPASSRLS, peran migrasi terpisah, tanpa FK lintas basis data (D-HRS-06).

Kelas PII: **PII-0** tidak ada; **PII-1** identitas langsung (nama, email, telepon); **PII-2** teregulasi/sensitif, wajib dienkripsi `kit/piicrypt` (NIK, NPWP, tanggal lahir); **SEC** rahasia teknis (hash sandi, kredensial SMTP). Pemetaan NIK/NPWP/data cuti ke kategori UU PDP final adalah keputusan hukum (Q-HRS-04); alasan cuti sakit mungkin data kesehatan `[INFERENCE]`.

| Entitas | Kolom kunci | RLS saat ini | Kelas PII | Bukti |
|---|---|---|---|---|
| `auth.tenants` | `id`, `name`, `slug`, status | Tidak ada RLS; INSERT/UPDATE hanya `hris_admin` | PII-0 | `services/auth-service/migrations/001_create_auth_schema.up.sql` |
| `auth.users` | `id`, `tenant_id`, `email`, `password_hash`, status | ENABLE+FORCE+WITH CHECK | PII-1, SEC (hash) | `services/auth-service/migrations/001_create_auth_schema.up.sql` |
| `auth.roles`, `auth.permissions`, `auth.role_permissions`, `auth.user_roles` | `tenant_id` (roles), kode izin | `role_permissions` dan `user_roles` tanpa RLS | PII-0 | `services/auth-service/migrations/001_create_auth_schema.up.sql` |
| `auth.login_attempts`, `auth.refresh_tokens` | percobaan login, token | `refresh_tokens` ENABLE+FORCE+WITH CHECK; keduanya tidak dipakai kode | PII-1 (IP) | `services/auth-service/migrations/001_create_auth_schema.up.sql` |
| `employee.departments` | `tenant_id`, `name`, `parent_id` (tidak dipakai) | ENABLE+FORCE+WITH CHECK | PII-0 | `services/employee-service/migrations/001_create_employee_schema.up.sql` |
| `employee.positions` | `tenant_id`, `name`, `level` | ENABLE+FORCE+WITH CHECK | PII-0 | `services/employee-service/migrations/001_create_employee_schema.up.sql` |
| `employee.employees` | `tenant_id`, `employee_number` (unik), `manager_id`, `national_id`, `tax_id`, `birth_date`, `phone`, `termination_date` | ENABLE+FORCE+WITH CHECK | PII-1; **PII-2: `national_id`, `tax_id`, `birth_date`** | `services/employee-service/migrations/001_create_employee_schema.up.sql` |
| `attendance.attendance_records` | `tenant_id`, `employee_id`, `date`, `check_in_at`, `check_out_at`, koordinat, `status`, `is_override` | ENABLE tanpa FORCE, kebijakan tanpa `WITH CHECK` | PII-1 (lokasi) | `services/attendance-service/migrations/001_create_attendance_schema.up.sql` |
| `leave.leave_types`, `leave_requests`, `leave_balances` | `tenant_id`, `employee_id`, `leave_type_id`, rentang tanggal, `days_count`, saldo | ENABLE tanpa FORCE, tanpa `WITH CHECK` | PII-1; `reason` mungkin kesehatan `[INFERENCE]` | `services/leave-service/migrations/001_create_leave_schema.up.sql` |
| `notification.notifications` | `tenant_id`, `recipient_id`, `channel`, `template_key`, `variables`, status | ENABLE tanpa FORCE; `WITH CHECK` lewat kebijakan terpisah | PII-1 | `services/notification-service/migrations/001_create_notification_schema.up.sql` |
| `notification.notification_channel_configs` | `tenant_id`, `channels`, `email_config` | ENABLE tanpa FORCE; `WITH CHECK` lewat kebijakan terpisah | SEC (kredensial SMTP polos) | `services/notification-service/migrations/001_create_notification_schema.up.sql` |
| `audit.entries` (partisi bulanan) | `tenant_id`, `actor_id`, `action`, `resource_type`, `resource_id`, `changes`, `created_at` | RLS dimatikan | PII-1 (aktor) | `services/audit-service/migrations/001_create_audit_schema.up.sql` |
| `document.files` | `tenant_id`, `entity_type`, `entity_id`, `minio_bucket`, `minio_key`, `classification`, `retention_days`, versi | ENABLE+FORCE+WITH CHECK | PII-2 (isi dokumen) | `services/document-service/migrations/001_create_document_schema.up.sql` |
| `document.file_versions` | `file_id`, `version_number`, `minio_key` | tanpa `tenant_id`, tanpa RLS | PII-2 (isi) | `services/document-service/migrations/001_create_document_schema.up.sql` |
| `*.processed_events` (auth, employee, leave, notification, audit, document) | `event_id`, `processed_at` | tanpa RLS (tabel idempotensi internal) | PII-0 | `services/leave-service/migrations/001_create_leave_schema.up.sql` |


Tabel target tambahan: `outbox` (per produk, kolom `event_id` UUIDv7, `event_type`, `payload`, `created_at`, `published_at`), `audit_logs` (`kit/audit`, append-only dengan trigger), `sessions` metadata (opsional; hash di Dragonfly), `entitlement_snapshot` (mode suite), `org_id`/`external_ref` pada `tenants` untuk tautan ke org platform. Semua tabel bertenant menyatakan `tenant_id` NOT NULL dan kebijakan `tenant_isolation` (D8).

Temuan skema yang harus dibereskan pada gerbang stabilisasi: (a) migrasi absensi 001 sintaks tidak valid; (b) indeks duplikat tanpa `IF NOT EXISTS`: `idx_attendance_employee_date` (absensi 001 baris 29 dan 002) serta `idx_leave_requests_employee_id`, `idx_leave_requests_status`, `idx_leave_requests_approved_by` (leave 001 dan 002), sehingga migrasi 002 kedua layanan diperkirakan gagal (`Pending Verification`); (c) peran `app` pada migrasi attendance, leave, notification sementara skrip init hanya membuat `hris_app` dan `hris_admin`; (d) RLS tanpa FORCE dan tanpa `WITH CHECK` pada attendance dan leave (notification tanpa FORCE, dengan `WITH CHECK` lewat kebijakan terpisah); (e) `current_setting('app.tenant_id')` tanpa `missing_ok` pada attendance, leave, dan notification akan melempar galat bila variabel belum diset.

---

## 6. Spesifikasi API

### 6.1 Ketentuan

- **Target:** REST `/api/v1` pada ServeMux stdlib, amplop `{success,data,error}` (`kit/httpx`), cookie sesi `stery_session` pada mode suite (nama cookie mode standalone `tbd`), izin eksplisit per rute (`RequirePermission`), tenant dan aktor dari sesi. gRPC `api/proto/v1` untuk klien mesin/platform dengan interseptor deny-by-default.
- **As-is:** 47 RPC pada delapan berkas proto (`proto/hris/*/v1/*.proto`); anotasi `google.api.http` memberi jalur `/v1/...`. Jalur HTTP tersebut **tidak disajikan oleh proses mana pun** (tidak ada `runtime.NewServeMux`/`ListenAndServe` di `services/**` di luar `*.pb.gw.go`; 11.3). Gateway Laravel hanya mengekspos rute `auth/*` dan cuti (2.3). Karena itu kolom status di bawah menilai **implementasi RPC**, sedangkan ketersediaan HTTP untuk semua rute adalah `Belum`.
- Rute target = rute anotasi proto dengan awalan `/api`. Rute statis didahulukan atas pola `{id}` oleh ServeMux (mis. `/api/v1/employees/org-chart` sebelum `/api/v1/employees/{id}`).
- Izin target memakai kode `resource:action` (3.3). `publik` = tanpa sesi, dengan batas laju.

### 6.2 Peta 47 RPC ke rute target

Status mengikuti 1.3: handler berisi tetapi cacat = `Sebagian`; handler kosong atau tidak ada = `Belum`.

| # | RPC | Rute target (`/api/v1`) | Izin target | Status | Bukti |
|---|---|---|---|---|---|
| 1 | `AuthService.Login` | `POST /api/v1/auth/login` | publik | Sebagian | `services/auth-service/internal/interfaces/grpc/auth_service.go` |
| 2 | `AuthService.RefreshToken` | dihapus pada target (sesi buram sliding); as-is `POST /v1/auth/refresh` | - | Sebagian | `services/auth-service/internal/application/commands/refresh_token.go` |
| 3 | `AuthService.ValidateToken` | gRPC `SessionService/Introspect` (internal); tanpa rute publik | token layanan | Sebagian | `services/auth-service/internal/interfaces/grpc/auth_service.go` |
| 4 | `AuthService.RevokeToken` | `POST /api/v1/auth/logout` | sesi sah | Sebagian | `services/auth-service/internal/application/commands/revoke_token.go` |
| 5 | `AuthService.GetPermissions` | `GET /api/v1/auth/me` (izin pengguna berjalan); `GET /api/v1/users/{id}/permissions` untuk admin | `user:manage` (admin) | Sebagian | `services/auth-service/internal/interfaces/grpc/auth_service.go` |
| 6 | `AuthService.RegisterTenant` | `POST /api/v1/auth/register` | publik | Sebagian | `services/auth-service/internal/application/commands/register_tenant.go` |
| 7 | `EmployeeService.CreateEmployee` | `POST /api/v1/employees` | `employee:write` | Sebagian | `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| 8 | `EmployeeService.GetEmployee` | `GET /api/v1/employees/{id}` | `employee:read` | Sebagian | `services/employee-service/internal/application/queries/get_employee.go` |
| 9 | `EmployeeService.UpdateEmployee` | `PATCH /api/v1/employees/{id}` | `employee:write` | Belum | `proto/hris/employee/v1/employee.proto` |
| 10 | `EmployeeService.TerminateEmployee` | `POST /api/v1/employees/{id}/terminate` | `employee:delete` | Sebagian | `services/employee-service/internal/application/commands/terminate_employee.go` |
| 11 | `EmployeeService.ListEmployees` | `GET /api/v1/employees` | `employee:read` | Sebagian | `services/employee-service/internal/application/queries/list_employees.go` |
| 12 | `EmployeeService.GetOrgChart` | `GET /api/v1/employees/org-chart` | `employee:read` | Belum | `proto/hris/employee/v1/employee.proto` |
| 13 | `EmployeeService.CreateDepartment` | `POST /api/v1/departments` | `employee:write` | Sebagian | `services/employee-service/internal/application/commands/create_department.go` |
| 14 | `EmployeeService.ListDepartments` | `GET /api/v1/departments` | `employee:read` | Sebagian | `services/employee-service/internal/application/queries/list_departments.go` |
| 15 | `EmployeeService.CreatePosition` | `POST /api/v1/positions` | `employee:write` | Sebagian | `services/employee-service/internal/application/commands/create_position.go` |
| 16 | `EmployeeService.ListPositions` | `GET /api/v1/positions` | `employee:read` | Sebagian | `services/employee-service/internal/application/queries/list_positions.go` |
| 17 | `AttendanceService.CheckIn` | `POST /api/v1/attendance/check-in` | `attendance:write` (diri sendiri) | Sebagian | `services/attendance-service/internal/application/commands/check_in.go` |
| 18 | `AttendanceService.CheckOut` | `POST /api/v1/attendance/check-out` | `attendance:write` (diri sendiri) | Sebagian | `services/attendance-service/internal/application/commands/check_out.go` |
| 19 | `AttendanceService.GetAttendanceSummary` | `GET /api/v1/attendance/summary/{employee_id}` | `attendance:read` | Sebagian | `services/attendance-service/internal/application/queries/get_daily_summary.go` |
| 20 | `AttendanceService.ListAttendance` | `GET /api/v1/attendance` | `attendance:read` | Sebagian | `services/attendance-service/internal/interfaces/grpc/attendance_service.go` |
| 21 | `AttendanceService.OverrideAttendance` | `POST /api/v1/attendance/override` | `attendance:write` (HR/atasan) | Sebagian | `services/attendance-service/internal/application/commands/override_attendance.go` |
| 22 | `LeaveService.ApplyLeave` | `POST /api/v1/leave-requests` | `leave:write` | Sebagian | `services/leave-service/internal/application/commands/create_leave_request.go` |
| 23 | `LeaveService.GetLeaveRequest` | `GET /api/v1/leave-requests/{id}` | `leave:read` | Sebagian | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| 24 | `LeaveService.ListLeaveRequests` | `GET /api/v1/leave-requests` | `leave:read` | Sebagian | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| 25 | `LeaveService.ApproveLeave` | `POST /api/v1/leave-requests/{id}/approve` | `leave:approve` | Sebagian | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| 26 | `LeaveService.RejectLeave` | `POST /api/v1/leave-requests/{id}/reject` | `leave:approve` | Sebagian | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| 27 | `LeaveService.CancelLeave` | `POST /api/v1/leave-requests/{id}/cancel` | `leave:write` | Sebagian | `services/leave-service/internal/domain/leave_request.go` |
| 28 | `LeaveService.GetLeaveBalance` | `GET /api/v1/leave-balance/{employee_id}` | `leave:read` | Sebagian | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| 29 | `LeaveService.ListLeaveTypes` | `GET /api/v1/leave-types` | `leave:read` | Sebagian | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| 30 | `NotificationService.SendNotification` | internal (gRPC/port `Notification`); tanpa rute publik | token layanan | Sebagian | `services/notification-service/internal/interfaces/grpc/notification_service.go` |
| 31 | `NotificationService.GetDeliveryStatus` | `GET /api/v1/notifications/{notification_id}` | `notification:read` | Sebagian | `services/notification-service/internal/interfaces/grpc/notification_service.go` |
| 32 | `NotificationService.ListNotifications` | `GET /api/v1/notifications` | `notification:read` | Sebagian | `services/notification-service/internal/interfaces/grpc/notification_service.go` |
| 33 | `NotificationService.MarkRead` | `POST /api/v1/notifications/mark-read` | `notification:read` | Sebagian | `services/notification-service/internal/application/commands/mark_notification_read.go` |
| 34 | `NotificationService.UpdateChannelConfig` | `PUT /api/v1/notifications/channel-config` | `settings:write` | Sebagian | `services/notification-service/internal/interfaces/grpc/notification_service.go` |
| 35 | `AuditService.Record` | internal (port `Audit.Record`); tanpa rute publik | token layanan | Sebagian | `services/audit-service/internal/interfaces/grpc/audit_service.go` |
| 36 | `AuditService.Query` | `GET /api/v1/audit/entries` | `audit:read` | Sebagian | `services/audit-service/internal/interfaces/grpc/audit_service.go` |
| 37 | `AuditService.GetEntry` | `GET /api/v1/audit/entries/{entry_id}` | `audit:read` | Sebagian | `services/audit-service/internal/interfaces/grpc/audit_service.go` |
| 38 | `DocumentService.UploadDocument` | `POST /api/v1/documents/upload` (target: minta presign unggah) | `document:write` | Sebagian | `services/document-service/internal/application/commands/upload_document.go` |
| 39 | `DocumentService.DownloadDocument` | `GET /api/v1/documents/{file_id}/download` | `document:read` | Belum | `services/document-service/internal/interfaces/grpc/document_service.go` |
| 40 | `DocumentService.ListDocuments` | `GET /api/v1/documents` | `document:read` | Sebagian | `services/document-service/internal/interfaces/grpc/document_service.go` |
| 41 | `DocumentService.GetDocumentMetadata` | `GET /api/v1/documents/{file_id}/metadata` | `document:read` | Sebagian | `services/document-service/internal/interfaces/grpc/document_service.go` |
| 42 | `DocumentService.DeleteDocument` | `DELETE /api/v1/documents/{file_id}` | `document:delete` | Sebagian | `services/document-service/internal/application/commands/delete_document.go` |
| 43 | `DocumentService.GetPresignedUrl` | `POST /api/v1/documents/{file_id}/presigned-url` | `document:read` | Belum | `services/document-service/internal/interfaces/grpc/document_service.go` |
| 44 | `DocumentService.RollbackDocumentVersion` | `POST /api/v1/documents/{file_id}/rollback` | `document:write` | Belum | `services/document-service/internal/interfaces/grpc/document_service.go` |
| 45 | `AIService.AnalyzeResume` | `POST /api/v1/ai/analyze-resume` | `ai:use` | Belum | `proto/hris/ai/v1/ai.proto` |
| 46 | `AIService.ForecastAttrition` | `POST /api/v1/ai/forecast-attrition` | `ai:use` | Belum | `proto/hris/ai/v1/ai.proto` |
| 47 | `AIService.Chat` | `POST /api/v1/ai/chat` (aliran SSE) | `ai:use` | Belum | `proto/hris/ai/v1/ai.proto` |

Hasil hitung: 47 RPC = auth 6, employee 10, attendance 5, leave 8, notification 5, audit 3, document 7, ai 3. Handler berisi: 39; kosong atau tidak ada: 8 (employee 2, document 3, ai 3). Ini membetulkan klaim awal "dokumen 5/7": yang berisi hanya 4 dari 7 (`DownloadDocument` juga placeholder).

### 6.3 Kontrak gRPC dan pemetaan galat

- Kontrak proto adalah dasar `api/proto/v1` target; perubahan yang diperlukan: tambah RPC manajemen pengguna dan izin dokumen/AI (FR-HRS-01.7, FR-HRS-07.*), hapus `RefreshToken`/`ValidateToken` token-JWT dan ganti dengan `SessionService` platform pada mode suite, tambah `X-Request-ID` sebagai metadata, dan bidang `org_id`.
- Pemetaan kode gRPC saat ini (jumlah pemakaian di `internal/interfaces/grpc/*.go`): attendance-service hanya memakai `codes.Internal` (5 kali), sehingga galat domain (mis. sudah check-in) sampai ke klien sebagai `Internal`; leave-service memakai paling beragam (`InvalidArgument`, `NotFound`, `PermissionDenied`, `Unauthenticated`, `FailedPrecondition`, `Internal`). Target: satu tabel pemetaan di `kit/httpx` dari galat domain ke status HTTP dan kode gRPC: validasi -> 400/`InvalidArgument`; tidak ditemukan -> 404/`NotFound`; tidak berizin -> 403/`PermissionDenied`; tanpa sesi -> 401/`Unauthenticated`; konflik status -> 409/`FailedPrecondition`; batas laju -> 429/`ResourceExhausted`; entitlement -> 402 atau 403 `tbd`.
- Amplop galat REST: `{"success":false,"error":{"code":"...","message":"..."}}`; `code` stabil, `message` terlokalisasi `{id,en}` oleh klien.

---

## 7. Alur Proses Kritis

### 7.1 Pengajuan dan persetujuan cuti

Alur target dengan titik putus as-is pada kolom kanan.

| Langkah | Alur target | Kondisi as-is | Status | Bukti |
|---|---|---|---|---|
| 1 | Karyawan membuka form cuti; klien memuat `GET /leave-types` dan `GET /leave-balance/{employee_id}` | daftar jenis cuti kosong (tidak ada pembuat/seed); halaman `/leaves/apply` ada | Sebagian | `apps/web/src/routes/(app)/leaves/apply/+page.svelte`, `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| 2 | `POST /leave-requests` dengan jenis, tanggal, alasan; `employee_id` dari sesi | `employee_id` dari badan permintaan; `ApplyLeave` tidak mengisi `days_count` (CHECK `> 0`) | Sebagian | `services/leave-service/internal/application/commands/create_leave_request.go` |
| 3 | Validasi: rentang valid, hari kerja dihitung, saldo cukup, tidak tumpang tindih; transaksi: sisipkan pengajuan `PENDING`, `pending_days += n`, tulis outbox `leave.requested` | saldo diperiksa di aplikasi tanpa kunci baris; tumpang tindih tidak dicek; payload peristiwa `{}` | Sebagian | `services/leave-service/internal/domain/leave_request.go`, `services/leave-service/internal/infrastructure/nats/event_publisher.go` |
| 4 | Konsumen: notifikasi ke atasan langsung/HR; audit | konsumen notification ada (dengan uji) tetapi email tidak terpasang dan payload kosong | Sebagian | `services/notification-service/internal/application/consumers/leave_event_consumer.go` |
| 5 | Atasan/HR membuka antrean (`GET /leave-requests?status=PENDING`) | filter status memakai `LEAVE_STATUS_*` yang tidak cocok dengan nilai DB | Sebagian | `services/leave-service/internal/interfaces/grpc/leave_service.go`, `apps/web/src/routes/(app)/admin/leaves/+page.svelte` |
| 6 | `POST /leave-requests/{id}/approve|reject`; aktor dari sesi, bukan pemohon sendiri, hanya dari `PENDING`; transaksi memperbarui saldo dan menulis outbox | `approver_id` dari badan; tidak ada pencegahan persetujuan diri; peran dari header `x-user-roles` | Sebagian | `services/leave-service/internal/interfaces/grpc/context_guards.go` |
| 7 | Konsumen: notifikasi ke pemohon; kehadiran hari terkait menjadi `ON_LEAVE` (FR-HRS-03.7); audit | kehadiran tidak mengonsumsi peristiwa cuti | Belum | `services/attendance-service/internal/domain/attendance_record.go` |
| 8 | Pembatalan oleh pemohon memulihkan saldo | `Cancel` mengizinkan `PENDING`; pemulihan saldo belum terverifikasi | Sebagian | `services/leave-service/internal/domain/leave_request.go` |

### 7.2 Rantai peristiwa onboarding

| Langkah | Alur target | Kondisi as-is | Status | Bukti |
|---|---|---|---|---|
| 1 | `POST /auth/register` (publik, dibatasi laju): satu transaksi membuat tenant, tiga peran, izin, admin `hr_admin`; outbox `tenant_registered` dan `user_registered` (muatan memuat `tenant_id`) | tidak atomik; INSERT `auth.tenants` hanya diizinkan untuk `hris_admin` (`Pending Verification`); payload `user_registered` tanpa `tenant_id` | Sebagian | `services/auth-service/internal/application/commands/register_tenant.go`, `services/auth-service/migrations/001_create_auth_schema.up.sql` |
| 2 | Employee mengonsumsi `user_registered`, membuat kerangka karyawan untuk admin (departemen/jabatan bawaan tenant), memancarkan `employee.created` | konsumen ada tetapi tidak dapat bekerja: `tenant_id` hilang, id departemen/jabatan diacak, galat di-ack lalu hilang | Sebagian | `services/employee-service/internal/infrastructure/nats_consumer.go` |
| 3 | Leave mengonsumsi `employee.created` dan membuat saldo tiap jenis cuti aktif, idempoten lewat `processed_events` | konsumen ada; `EmployeeValidator` no-op; tanpa jenis cuti tidak ada saldo | Sebagian | `services/leave-service/internal/infrastructure/nats/employee_created_consumer.go` |
| 4 | Notification mengirim email sambutan/aktivasi dan notifikasi in-app | adaptor email tidak terpasang | Sebagian | `services/notification-service/cmd/server/main.go` |
| 5 | Audit mencatat `tenant_registered`, `user_registered`, `employee.created` | konsumen ada; `resource_id` UUID vs string kosong | Pending Verification | `services/audit-service/internal/infrastructure/nats/identity_consumer.go` |

Pada mode suite, langkah 1 digantikan: org dan admin dibuat platform; HRIS menerima `Deliver` `org_created` dan membuat tenant HRIS yang tertaut `external_ref` (FSD suite 7.5).

### 7.3 Check-in dan check-out kehadiran

| Langkah | Alur target | Kondisi as-is | Status | Bukti |
|---|---|---|---|---|
| 1 | Karyawan membuka halaman kehadiran (SPA) | tidak ada halaman kehadiran di web | Belum | `apps/web/src/routes/(app)/dashboard/+page.svelte` |
| 2 | `POST /attendance/check-in` (opsional koordinat); `employee_id` dan `tenant_id` dari sesi; menolak bila sudah ada catatan aktif hari ini (zona tenant) | `CheckIn` ada; tenant dari badan; migrasi 001 tidak valid sehingga tidak dapat dieksekusi di DB segar | Sebagian | `services/attendance-service/internal/application/commands/check_in.go` |
| 3 | Status `PRESENT` atau `LATE` menurut aturan tenant (FR-HRS-03.6) | `LATE` tidak pernah dihitung | Belum | `services/attendance-service/internal/domain/attendance_record.go` |
| 4 | `POST /attendance/check-out`: harus ada check-in aktif; durasi dihitung; outbox `checked_out` | `CheckOut` ada | Sebagian | `services/attendance-service/internal/application/commands/check_out.go` |
| 5 | HR/atasan mengoreksi lewat `POST /attendance/override` dengan alasan; catatan baru bila tidak ada | override gagal untuk catatan baru | Sebagian | `services/attendance-service/internal/application/commands/override_attendance.go` |
| 6 | Peristiwa dikonsumsi audit | subjek `hris.attendance.record.*` tidak cocok dengan konsumen `hris.operations.>`; payload `{}` | Sebagian | `services/attendance-service/internal/domain/events.go` |

---

## 8. Keamanan, Privasi dan UU PDP

### 8.1 Klasifikasi data

| Kelas | Contoh | Perlakuan target |
|---|---|---|
| PII-2 (teregulasi/sensitif) | NIK (`national_id`), NPWP (`tax_id`), tanggal lahir, isi dokumen karyawan, alasan cuti sakit `[INFERENCE]` | enkripsi `kit/piicrypt` (AES-256-GCM, AAD per tujuan, indeks buta), tampilan dimasking, akses diaudit, tidak di log, tidak ke pihak ketiga tanpa dasar hukum (FR-HRS-SEC.3) |
| PII-1 (identitas langsung) | nama, email, telepon, koordinat absensi | akses sesuai peran, tidak di log secara polos |
| SEC (rahasia teknis) | hash kata sandi, kredensial SMTP, kunci enkripsi | tidak pernah dicatat; kredensial SMTP terenkripsi; kunci per lingkungan |
| PII-0 | data organisasi non-personal | tanpa perlakuan khusus |

### 8.2 Postur keamanan saat ini

| Temuan | Status | Bukti |
|---|---|---|
| Tidak ada autentikasi/otorisasi di interseptor gRPC mana pun; tenant dan aktor dipercaya dari badan permintaan atau metadata | Belum | `services/_shared/server/grpc_interceptors.go` |
| Alamat IP klien selalu `0.0.0.0` | Sebagian | `services/auth-service/internal/interfaces/grpc/auth_service.go` |
| Tidak ada penguncian akun atau pembatasan laju (`throttle` Laravel tidak pernah dipasang; tabel `login_attempts` tidak dipakai) | Belum | `apps/api-gateway/routes/api.php`, `services/auth-service/migrations/001_create_auth_schema.up.sql` |
| JWT RS256 dengan TTL hardcoded 15 menit/7 hari; tanpa deteksi pemakaian ulang keluarga refresh token | Sebagian | `services/auth-service/cmd/server/main.go`, `services/auth-service/internal/infrastructure/jwt_signer.go` |
| Gateway tidak mengekstrak token: `getBearerToken()` mengembalikan string kosong; `LeaveController` membaca peran dari atribut yang tidak pernah diisi | Belum | `apps/api-gateway/app/Services/BaseGrpcGatewayClient.php`, `apps/api-gateway/app/Http/Middleware/JwtValidation.php` |
| Hook SvelteKit mempercayai cookie `auth_user` (JSON peran) yang tidak HttpOnly; JWT tidak diverifikasi di server web | Sebagian | `apps/web/src/hooks.server.ts`, `apps/web/src/routes/(auth)/login/+page.server.ts` |
| Kredensial SMTP tersimpan polos di JSONB | Sebagian | `services/notification-service/migrations/001_create_notification_schema.up.sql` |
| Kredensial dev dan bawaan tertulis di berkas terlacak git (nilai tidak dikutip di sini) | Sebagian | `deploy/docker/docker-compose.yml`, `deploy/docker/nats.conf`, `deploy/docker/init-postgres.sql`, `docker-compose.production.yml` |
| RLS ENABLE+FORCE+WITH CHECK hanya pada auth (users, roles, refresh_tokens), employee, dan document.files; sisanya parsial atau tanpa RLS | Sebagian | `services/employee-service/migrations/001_create_employee_schema.up.sql`, `services/document-service/migrations/001_create_document_schema.up.sql` |
| gRPC antar layanan tanpa TLS dan tanpa autentikasi layanan; UploadDocument tidak dapat berhasil karena konteks aktor tidak diisi | Belum | `services/document-service/internal/interfaces/grpc/mapping.go` |
| Berkas biner terkompilasi `server.exe` ada di pohon kerja beberapa layanan tetapi **tidak** dilacak git (diabaikan oleh `*.exe`); perlu dibersihkan dari pohon kerja | Sebagian | `.gitignore` |

### 8.3 Persyaratan UU PDP

UU No. 27 Tahun 2022 tentang Pelindungan Data Pribadi berlaku bagi pemrosesan data karyawan. Repo tidak memuat kebijakan, bahasa persetujuan, retensi, atau prosedur insiden apa pun (ADR-0004 baris 47 hanya menyebut PDPA satu kali dalam konteks payroll; `docs/adr/0004-defer-payroll-to-phase2.md`). Persyaratan berikut adalah **kebutuhan produk**; isi hukum final (kategori data, dasar pemrosesan, batas waktu, kewajiban pemberitahuan) memerlukan tinjauan hukum dan ditulis `tbd` sampai Q-HRS-04 dan Q-HRS-14 dijawab.

| Area | Persyaratan produk | FR terkait | Isi hukum |
|---|---|---|---|
| Klasifikasi dan peta data | Inventaris kolom PII per kelas (8.1); NIK dan NPWP diperlakukan sebagai PII-2 secara bawaan | FR-HRS-02.8, FR-HRS-SEC.3 | pemetaan ke kategori UU PDP `tbd` |
| Enkripsi | `kit/piicrypt` untuk NIK, NPWP, tanggal lahir; kunci terpisah per lingkungan; rotasi kunci | FR-HRS-SEC.3 | - |
| Dasar pemrosesan dan persetujuan | Mencatat dasar pemrosesan per kelas data; persetujuan eksplisit untuk pemrosesan di luar kontrak kerja (mis. kirim data ke LLM, FR-HRS-08.*) | FR-HRS-SEC.8 | dasar hukum `tbd` |
| Retensi dan penghapusan | Masa retensi per kelas data (karyawan, kehadiran, cuti, dokumen, audit); penghapusan/anonimisasi saat tenggat | FR-HRS-SEC.11, FR-HRS-06.5, FR-HRS-07.5 | masa retensi `tbd` (Q-HRS-04) |
| Hak subjek data | Akses, koreksi, penghapusan, pembatasan, portabilitas data karyawan atas permintaan lewat HR; log permintaan | FR-HRS-SEC.8 | tenggat jawab `tbd` |
| Pengendali dan prosesor | Tenant = pengendali, Stery = prosesor; templat perjanjian pemrosesan data | - | templat `tbd` (Q-HRS-14) |
| Insiden | Prosedur deteksi dan pemberitahuan insiden kebocoran data; log keamanan | FR-HRS-SEC.10 | tenggat dan format `tbd` |
| Transfer ke pihak ketiga | Tidak ada PII-2 ke penyedia LLM/email/SMS tanpa dasar dan perjanjian | FR-HRS-08.3, FR-HRS-08.4 | `tbd` (Q-HRS-09) |
| Isolasi tenant | RLS dan uji lintas tenant sebagai kontrol teknis kerahasiaan | FR-HRS-SEC.1, FR-HRS-SEC.2 | - |

### 8.4 Persyaratan keamanan lintas modul

Persyaratan keamanan lintas modul FR-HRS-SEC.1 sampai FR-HRS-SEC.11 didefinisikan di bagian 4.SEC.

---

## 9. Kebutuhan Non-Fungsional Teknis

Bagian ini menerjemahkan NFR-HRS (didefinisikan di `PRD_HRIS.id.md` bagian 6) menjadi realisasi teknis dan metode verifikasi. Status dan bukti menilai kondisi saat ini. Nilai target numerik yang tidak ada di repo ditulis `tbd` (Q-HRS-07, Q-HRS-15).

| NFR | Realisasi teknis | Verifikasi | Status | Bukti |
|---|---|---|---|---|
| NFR-HRS-01 | RLS ENABLE+FORCE+`WITH CHECK` pada semua tabel bertenant lewat `kit/tenantdb.WithTenant`; peran NOBYPASSRLS (FR-HRS-SEC.1) | uji lintas tenant per modul pada Postgres nyata (FR-HRS-SEC.2) | Sebagian | `services/_shared/postgres/rls.go` |
| NFR-HRS-02 | Sesi buram, hash di Dragonfly, cookie HttpOnly Secure; TTL `tbd` (FR-HRS-SEC.4) | uji integrasi login/logout/kedaluwarsa; pemeriksaan cookie di e2e | Belum | `services/auth-service/internal/infrastructure/jwt_signer.go` |
| NFR-HRS-03 | `RequirePermission(code)` + interseptor gRPC deny-by-default (FR-HRS-SEC.5) | uji tabel rute-izin: setiap rute memiliki izin (gagal bila ada rute tanpa izin) | Belum | `services/_shared/server/grpc_interceptors.go` |
| NFR-HRS-04 | `kit/piicrypt` AES-256-GCM, AAD per tujuan, indeks buta, rotasi kunci (FR-HRS-SEC.3) | uji unit round-trip dan rotasi; uji migrasi data lama | Belum | `services/employee-service/migrations/001_create_employee_schema.up.sql` |
| NFR-HRS-05 | `audit_logs` append-only dengan trigger penolak UPDATE/DELETE/TRUNCATE (FR-HRS-06.4) | uji integrasi yang mencoba UPDATE/DELETE harus gagal | Sebagian | `services/audit-service/migrations/001_create_audit_schema.up.sql` |
| NFR-HRS-06 | Target ketersediaan `tbd`; `/healthz`; restart otomatis oleh pengelola proses | pemantauan; target hosting belum ada (Q-HRS-12) | Belum | `docker-compose.production.yml` |
| NFR-HRS-07 | Latensi p95 `tbd`; indeks sesuai pola kueri; paginasi wajib | uji beban pada tumpukan nyata (belum ada) | Belum | `services/leave-service/migrations/002_add_performance_indexes.up.sql` |
| NFR-HRS-08 | Pembatas laju Dragonfly (skrip Lua) per alamat/akun dan kelas rute; penguncian akun (FR-HRS-01.5, FR-HRS-SEC.6) | uji integrasi: N percobaan gagal menghasilkan 429 | Belum | `apps/api/cmd/api/routes.go` |
| NFR-HRS-09 | Migrasi goose reversibel; job CI `up -> down -> up` untuk **semua** modul pada DB segar | job CI hijau | Sebagian: CI menjalankan `up -> down -> up` hanya untuk auth, employee, leave dengan golang-migrate; attendance, audit, notification, document tidak tercakup, dan migrasi absensi tidak valid | `.github/workflows/ci-go.yml` |
| NFR-HRS-10 | `scripts/check.sh` seperti SELA; cakupan >= 80%; `-race` | CI hijau; ambang cakupan | Sebagian: CI Go memakai `-race`, tetapi `GO_VERSION` CI 1.24 sedangkan `go.work` 1.25.0; tidak ada ambang cakupan; document-service tanpa berkas uji | `.github/workflows/ci-go.yml`, `go.work` |
| NFR-HRS-11 | `slog` JSON tanpa PII, `/healthz`; OTel/Prometheus opsional (Q-HRS-08) | uji log tidak memuat PII; `/healthz` hijau | Sebagian: OTel `InitTracer` dan `NewMetricsRegistry` didefinisikan di `_shared` tetapi tidak pernah dipanggil layanan; `StartMetricsServer` mengikat `:8081` untuk semua layanan | `services/_shared/observability/otel.go`, `services/_shared/observability/metrics.go` |
| NFR-HRS-12 | Katalog i18n `{id,en}` dengan uji paritas; WIB; IDR bila ada nilai uang | uji paritas katalog di Vitest | Belum: tidak ada pustaka i18n | `apps/web/package.json` |
| NFR-HRS-13 | Env + validasi gagal-cepat + `.env.example`; tanpa rahasia di berkas terlacak | pemindai rahasia di CI; peninjauan | Sebagian | `services/_shared/config/loader.go`, `deploy/docker/docker-compose.yml` |
| NFR-HRS-14 | Cadangan Postgres terjadwal + uji pemulihan; RPO/RTO `tbd` (Q-HRS-07) | latihan pemulihan terjadwal | Belum | `docker-compose.production.yml` |
| NFR-HRS-15 | Pekerja retensi per kelas data (FR-HRS-SEC.11) | uji pekerja menghapus/menganonimkan data kedaluwarsa | Belum | `services/document-service/migrations/001_create_document_schema.up.sql` |
| NFR-HRS-16 | `STERY_MODE=standalone\|suite` dengan dua adaptor per port; uji regresi standalone | suite uji yang sama lulus di kedua mode untuk perilaku standalone | Belum | `proto/hris/auth/v1/auth.proto` |
| NFR-HRS-17 | Komponen UI dapat diakses papan ketik/pembaca layar; tingkat WCAG `tbd` | audit aksesibilitas otomatis + manual pada alur utama | Belum | `apps/web/src/routes/(auth)/login/+page.svelte` |
| NFR-HRS-18 | Outbox transaksional (`FOR UPDATE SKIP LOCKED`) + konsumen idempoten `processed_events` | uji: kegagalan di tengah publikasi tidak menghilangkan peristiwa | Sebagian: `processed_events` ada di semua skema; outbox tidak ada | `services/leave-service/migrations/001_create_leave_schema.up.sql` |

---

## 10. Struktur Layar

### 10.1 Halaman saat ini

| Rute web | Fungsi | Status | Bukti |
|---|---|---|---|
| `/` | halaman awal | Sebagian | `apps/web/src/routes/+page.svelte` |
| `/login` | formulir login; aksi server mengirim POST ke API Go `/api/v1/auth/login` dan mengisi cookie HttpOnly `hris_session` (token dan peran) lewat session store | Sebagian | `apps/web/src/routes/(auth)/login/+page.server.ts`, `apps/web/src/lib/server/session.ts` |
| `/dashboard` | dasbor pengguna | Sebagian | `apps/web/src/routes/(app)/dashboard/+page.svelte` |
| `/leaves/apply` | form pengajuan cuti | Sebagian | `apps/web/src/routes/(app)/leaves/apply/+page.svelte` |
| `/admin/leaves` | antrean persetujuan cuti untuk manajer/HR; proteksi peran di hook server membaca peran dari cookie `auth_user` | Sebagian | `apps/web/src/routes/(app)/admin/leaves/+page.svelte`, `apps/web/src/hooks.server.ts` |
| `403.svelte` | halaman akses ditolak; kemungkinan bukan rute karena nama berkas tanpa awalan `+` (`[INFERENCE]`) | Pending Verification | `apps/web/src/routes/403.svelte` |

Catatan: klien API (`src/lib/api/client.ts`) tidak mengirim header Authorization; store `auth.ts` mendefinisikan `setSession` yang tidak dipanggil siapa pun, sehingga klaim awal "token disimpan ganda di cookie dan localStorage" tidak terbukti (hanya cookie yang diisi). Uji e2e di `tests/e2e/` ada tetapi mengacu pada jalur yang tidak ada (`/leaves` tanpa `/apply`, `[INFERENCE]`) (`apps/web/tests/e2e/auth-and-rbac.spec.ts`). Tidak ada halaman untuk karyawan, kehadiran, departemen, notifikasi, audit, dokumen, dan pengaturan.

### 10.2 Layar target (SPA `apps/web`)

Grup rute SPA: `(auth)` publik dan `(app)` dengan penjaga sesi klien yang memanggil `GET /api/v1/auth/me`. Rute target:

| Rute target | Layar | FR terkait | Peran |
|---|---|---|---|
| `/login`, `/register` | login; registrasi tenant | FR-HRS-01.1, FR-HRS-01.2 | publik |
| `/dashboard` | ringkasan: kehadiran hari ini, saldo cuti, antrean persetujuan | FR-HRS-03.5, FR-HRS-04.7 | semua |
| `/employees`, `/employees/new`, `/employees/{id}` | daftar, buat, ubah, terminasi karyawan | FR-HRS-02.1..02.4 | `hr_admin` (manajer: baca tim) |
| `/org` | bagan organisasi | FR-HRS-02.7 | semua (baca) |
| `/departments`, `/positions` | kelola departemen dan jabatan | FR-HRS-02.5, FR-HRS-02.6 | `hr_admin` |
| `/attendance` | check-in/out, riwayat saya | FR-HRS-03.1, FR-HRS-03.2 | `employee` |
| `/admin/attendance` | daftar, ringkasan, koreksi | FR-HRS-03.3..03.5 | `hr_admin`, `manager` |
| `/leaves`, `/leaves/apply` | riwayat, ajukan, batalkan, saldo | FR-HRS-04.3, FR-HRS-04.5..04.7 | `employee` |
| `/admin/leaves`, `/admin/leave-types` | persetujuan; jenis cuti | FR-HRS-04.1, FR-HRS-04.4 | `manager`, `hr_admin` |
| `/notifications` | kotak masuk in-app | FR-HRS-05.4 | semua |
| `/documents` | daftar, unggah, unduh dokumen karyawan | FR-HRS-07.1, FR-HRS-07.3, FR-HRS-07.4 | sesuai klasifikasi |
| `/admin/audit` | penelusuran jejak audit | FR-HRS-06.2 | `hr_admin`/auditor |
| `/admin/settings` | kanal notifikasi, aturan kerja, pengguna dan peran | FR-HRS-05.5, FR-HRS-03.6, FR-HRS-01.7 | `hr_admin` |

Setiap layar mematuhi: string `{id,en}` (NFR-HRS-12), keadaan kosong/galat/memuat, aksesibilitas papan ketik (NFR-HRS-17), dan tidak menampilkan PII-2 kecuali izin sah (masking bawaan).

---

## 11. Matriks Ketertelusuran dan Status

### 11.1 Status per FR

Rujukan ringkas; definisi lengkap tiap FR ada di bagian 4 dan 8.4.

| FR | Status | Celah utama | Bukti |
|---|---|---|---|
| FR-HRS-01.1 | Sebagian | tidak atomik; hak `auth.tenants` hanya `hris_admin` | `services/auth-service/internal/application/commands/register_tenant.go` |
| FR-HRS-01.2 | Sebagian | login web kini memanggil API Go (tiruan teratasi); JWT bukan sesi buram | `apps/api/services/auth/application/commands/login.go` |
| FR-HRS-01.3 | Selesai | - | `services/auth-service/internal/domain/user.go` |
| FR-HRS-01.4 | Sebagian | tanpa deteksi pemakaian ulang; TTL hardcoded | `services/auth-service/internal/application/commands/refresh_token.go` |
| FR-HRS-01.5 | Belum | `login_attempts` tidak dipakai | `services/auth-service/migrations/001_create_auth_schema.up.sql` |
| FR-HRS-01.6 | Sebagian | tidak ditegakkan; matriks keliru | `services/auth-service/internal/domain/permission.go` |
| FR-HRS-01.7 | Belum | tanpa RPC pengguna | `proto/hris/auth/v1/auth.proto` |
| FR-HRS-01.8 | Belum | tanpa verifikasi/reset | `services/auth-service/internal/domain/user.go` |
| FR-HRS-01.9 | Sebagian | IP `0.0.0.0` | `services/auth-service/internal/interfaces/grpc/auth_service.go` |
| FR-HRS-01.10 | Belum | tanpa adaptor suite | `proto/hris/auth/v1/auth.proto` |
| FR-HRS-02.1 | Sebagian | `employee_number` tidak diisi; enum beda huruf | `services/employee-service/internal/infrastructure/postgres/employee_repository.go` |
| FR-HRS-02.2 | Sebagian | Total = panjang halaman; offset 0 | `services/employee-service/internal/application/queries/list_employees.go` |
| FR-HRS-02.3 | Belum | handler tidak ada | `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| FR-HRS-02.4 | Sebagian | tanpa kaskade | `services/employee-service/internal/application/commands/terminate_employee.go` |
| FR-HRS-02.5 | Sebagian | `parent_id` tidak dipakai | `services/employee-service/internal/application/commands/create_department.go` |
| FR-HRS-02.6 | Sebagian | kolom `description` tidak ada | `services/employee-service/internal/infrastructure/postgres/position_repository.go` |
| FR-HRS-02.7 | Belum | handler tidak ada | `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| FR-HRS-02.8 | Belum | PII polos | `services/employee-service/migrations/001_create_employee_schema.up.sql` |
| FR-HRS-02.9 | Belum | - | `proto/hris/employee/v1/employee.proto` |
| FR-HRS-02.10 | Sebagian | payload tanpa `tenant_id` | `services/employee-service/internal/infrastructure/nats_consumer.go` |
| FR-HRS-02.11 | Belum | tanpa batas | `services/employee-service/internal/application/commands/create_employee.go` |
| FR-HRS-03.1 | Sebagian | migrasi tidak valid; RBAC karyawan | `services/attendance-service/internal/application/commands/check_in.go` |
| FR-HRS-03.2 | Sebagian | idem | `services/attendance-service/internal/application/commands/check_out.go` |
| FR-HRS-03.3 | Sebagian | gagal untuk catatan baru | `services/attendance-service/internal/application/commands/override_attendance.go` |
| FR-HRS-03.4 | Sebagian | filter tanggal diabaikan | `services/attendance-service/internal/interfaces/grpc/attendance_service.go` |
| FR-HRS-03.5 | Sebagian | `employee_id`/`period` diabaikan | `services/attendance-service/internal/application/queries/get_daily_summary.go` |
| FR-HRS-03.6 | Belum | `LATE`/`HALF_DAY` tidak dihitung | `services/attendance-service/internal/domain/attendance_record.go` |
| FR-HRS-03.7 | Belum | tanpa konsumen cuti | `services/attendance-service/internal/domain/attendance_record.go` |
| FR-HRS-03.8 | Belum | tanpa zona tenant | `services/attendance-service/migrations/001_create_attendance_schema.up.sql` |
| FR-HRS-04.1 | Sebagian | tanpa pembuat/seed | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| FR-HRS-04.2 | Sebagian | validator no-op; tanpa jenis cuti | `services/leave-service/internal/infrastructure/nats/employee_created_consumer.go` |
| FR-HRS-04.3 | Sebagian | `days_count` tidak diisi; tanpa cek tumpang tindih | `services/leave-service/internal/application/commands/create_leave_request.go` |
| FR-HRS-04.4 | Sebagian | aktor dari badan; tanpa cegah persetujuan diri | `services/leave-service/internal/interfaces/grpc/context_guards.go` |
| FR-HRS-04.5 | Sebagian | `Cancel` mengizinkan `PENDING` | `services/leave-service/internal/domain/leave_request.go` |
| FR-HRS-04.6 | Sebagian | filter status tidak cocok | `services/leave-service/internal/interfaces/grpc/leave_service.go` |
| FR-HRS-04.7 | Sebagian | melintasi tahun dihitung di tahun mulai | `services/leave-service/internal/domain/leave_balance.go` |
| FR-HRS-04.8 | Belum | - | `services/leave-service/migrations/001_create_leave_schema.up.sql` |
| FR-HRS-04.9 | Belum | - | `services/leave-service/internal/domain/leave_balance.go` |
| FR-HRS-05.1 | Sebagian | - | `services/notification-service/migrations/001_create_notification_schema.up.sql` |
| FR-HRS-05.2 | Sebagian | adaptor tidak dipasang; templat tidak dimuat | `services/notification-service/internal/infrastructure/channels/email_adapter.go` |
| FR-HRS-05.3 | Sebagian | payload kosong | `services/notification-service/internal/application/consumers/leave_event_consumer.go` |
| FR-HRS-05.4 | Sebagian | `MarkRead` gagal diam-diam | `services/notification-service/internal/application/commands/mark_notification_read.go` |
| FR-HRS-05.5 | Sebagian | SMTP polos | `services/notification-service/internal/interfaces/grpc/notification_service.go` |
| FR-HRS-05.6 | Belum | - | `proto/hris/notification/v1/notification.proto` |
| FR-HRS-05.7 | Belum | - | `proto/hris/notification/v1/notification.proto` |
| FR-HRS-06.1 | Sebagian | `resource_id` UUID vs string kosong | `services/audit-service/internal/infrastructure/nats/identity_consumer.go` |
| FR-HRS-06.2 | Pending Verification | hak `hris_app` INSERT saja | `services/audit-service/internal/interfaces/grpc/audit_service.go` |
| FR-HRS-06.3 | Sebagian | `GetEntry` memakai `Limit:1` | `services/audit-service/internal/interfaces/grpc/audit_service.go` |
| FR-HRS-06.4 | Sebagian | pembatasan hak, tanpa trigger | `services/audit-service/migrations/001_create_audit_schema.up.sql` |
| FR-HRS-06.5 | Sebagian | partisi statis | `services/audit-service/migrations/001_create_audit_schema.up.sql` |
| FR-HRS-06.6 | Sebagian | subjek tidak cocok | `services/audit-service/internal/infrastructure/nats/operations_consumer.go` |
| FR-HRS-06.7 | Belum | RLS dimatikan | `services/audit-service/migrations/001_create_audit_schema.up.sql` |
| FR-HRS-07.1 | Sebagian | `Unauthenticated` selalu | `services/document-service/internal/application/commands/upload_document.go` |
| FR-HRS-07.2 | Belum | tanpa pemindai | `services/document-service/migrations/001_create_document_schema.up.sql` |
| FR-HRS-07.3 | Belum | placeholder | `services/document-service/internal/interfaces/grpc/document_service.go` |
| FR-HRS-07.4 | Sebagian | terhapus tidak disaring | `services/document-service/internal/interfaces/grpc/document_service.go` |
| FR-HRS-07.5 | Sebagian | tanpa pekerja retensi | `services/document-service/internal/application/commands/delete_document.go` |
| FR-HRS-07.6 | Belum | tidak dikabelkan | `services/document-service/internal/application/commands/create_document_version.go` |
| FR-HRS-07.7 | Sebagian | tidak ditegakkan | `services/document-service/migrations/001_create_document_schema.up.sql` |
| FR-HRS-07.8 | Belum | tanpa validasi | `services/document-service/internal/application/commands/upload_document.go` |
| FR-HRS-08.1 | Belum | hanya proto | `proto/hris/ai/v1/ai.proto` |
| FR-HRS-08.2 | Belum | hanya proto | `proto/hris/ai/v1/ai.proto` |
| FR-HRS-08.3 | Belum | hanya proto | `proto/hris/ai/v1/ai.proto` |
| FR-HRS-08.4 | Belum | `ai-service` dikomentari | `go.work` |
| FR-HRS-SEC.1 | Sebagian | RLS tidak merata | `services/_shared/postgres/rls.go` |
| FR-HRS-SEC.2 | Sebagian | CI 3 layanan, bypass RLS | `.github/workflows/ci-go.yml` |
| FR-HRS-SEC.3 | Belum | PII polos | `services/employee-service/migrations/001_create_employee_schema.up.sql` |
| FR-HRS-SEC.4 | Belum | JWT | `services/auth-service/internal/infrastructure/jwt_signer.go` |
| FR-HRS-SEC.5 | Belum | tanpa authn/authz | `services/_shared/server/grpc_interceptors.go` |
| FR-HRS-SEC.6 | Belum | tidak ada pembatas laju yang dipasang di `cmd/api` | `apps/api/cmd/api/routes.go` |
| FR-HRS-SEC.7 | Sebagian | kredensial dev terlacak | `deploy/docker/docker-compose.yml` |
| FR-HRS-SEC.8 | Belum | - | `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| FR-HRS-SEC.9 | Belum | - | `apps/web/svelte.config.js` |
| FR-HRS-SEC.10 | Belum | `zap`; tanpa redaksi | `services/_shared/observability/logging.go` |
| FR-HRS-SEC.11 | Belum | - | `services/document-service/migrations/001_create_document_schema.up.sql` |

Ringkasan hitung (dihitung dari tabel di atas, 75 FR): `Selesai` 1; `Sebagian` 40; `Belum` 33; `Pending Verification` 1; `Ditunda` 0 (modul 09-12 tidak memiliki FR). Hitungan ini bukan persentase kesiapan.

### 11.2 Matriks BR -> FR

Setiap BR-HRS-01..22 (didefinisikan di `BRD_HRIS.id.md` bagian 6) dipetakan ke FR yang memenuhinya. Dua BR berstatus Ditunda tidak memiliki FR (modul 09-12 hanya memuat pembenaran penundaan, 4.09-4.12).

| BR | FR pemenuhan | Catatan |
|---|---|---|
| BR-HRS-01 | FR-HRS-01.1, FR-HRS-01.6, FR-HRS-01.9, FR-HRS-02.10 | Registrasi mandiri tenant dan admin pertama |
| BR-HRS-02 | FR-HRS-SEC.1, FR-HRS-SEC.2, FR-HRS-06.7 | Isolasi tenant dan uji lintas tenant |
| BR-HRS-03 | FR-HRS-01.6, FR-HRS-01.7, FR-HRS-SEC.4, FR-HRS-SEC.5 | Otorisasi di backend |
| BR-HRS-04 | FR-HRS-02.1, FR-HRS-02.2, FR-HRS-02.3, FR-HRS-02.4, FR-HRS-02.5, FR-HRS-02.6, FR-HRS-02.7, FR-HRS-02.9 | Data induk |
| BR-HRS-05 | FR-HRS-03.1, FR-HRS-03.2, FR-HRS-03.3, FR-HRS-03.4, FR-HRS-03.5, FR-HRS-03.7 | Kehadiran dan koreksi |
| BR-HRS-06 | FR-HRS-03.6, FR-HRS-03.8 | Aturan kerja dan zona waktu per tenant |
| BR-HRS-07 | FR-HRS-04.1, FR-HRS-04.2, FR-HRS-04.3, FR-HRS-04.4, FR-HRS-04.5, FR-HRS-04.6, FR-HRS-04.7, FR-HRS-04.8, FR-HRS-04.9 | Cuti end-to-end |
| BR-HRS-08 | FR-HRS-05.1, FR-HRS-05.2, FR-HRS-05.3, FR-HRS-05.4, FR-HRS-05.5, FR-HRS-05.6, FR-HRS-05.7 | Notifikasi |
| BR-HRS-09 | FR-HRS-06.1, FR-HRS-06.2, FR-HRS-06.3, FR-HRS-06.4, FR-HRS-06.5, FR-HRS-06.6 | Jejak audit |
| BR-HRS-10 | FR-HRS-07.1, FR-HRS-07.2, FR-HRS-07.3, FR-HRS-07.4, FR-HRS-07.5, FR-HRS-07.6, FR-HRS-07.7, FR-HRS-07.8 | Dokumen |
| BR-HRS-11 | FR-HRS-08.1, FR-HRS-08.2, FR-HRS-08.3, FR-HRS-08.4 | AI (prasyarat kebijakan) |
| BR-HRS-12 | FR-HRS-02.8, FR-HRS-SEC.3, FR-HRS-SEC.8, FR-HRS-SEC.10, FR-HRS-SEC.11 | UU PDP (8.3) |
| BR-HRS-13 | (Ditunda, tanpa FR; lihat 4.09 dan 4.10) | ADR-0004 |
| BR-HRS-14 | (Ditunda, tanpa FR; lihat 4.11 dan 4.12) | ADR-0004 |
| BR-HRS-15 | FR-HRS-01.10, FR-HRS-02.11 | Dua mode (standalone/suite) |
| BR-HRS-16 | FR-HRS-02.11, FR-HRS-01.10 | Dimensi `hrs.employees.max` dan perilaku saat melampaui batas |
| BR-HRS-17 | FR-HRS-01.2, FR-HRS-02.3, FR-HRS-02.7, FR-HRS-07.1, FR-HRS-07.6, FR-HRS-SEC.3, FR-HRS-SEC.6, FR-HRS-SEC.7 | Gerbang stabilisasi HRIS (BRD 12); migrasi up->down->up lewat NFR-HRS-09 |
| BR-HRS-18 | FR-HRS-SEC.4, FR-HRS-SEC.5, FR-HRS-SEC.9, FR-HRS-05.3, FR-HRS-06.1 | Penyelarasan stack lewat 2.1 dan 2.7; kebutuhan teknis di bagian 9 |
| BR-HRS-19 | FR-HRS-03.8, FR-HRS-05.2 | Dua bahasa, WIB; i18n antarmuka lewat NFR-HRS-12 |
| BR-HRS-20 | FR-HRS-01.5, FR-HRS-SEC.6, FR-HRS-SEC.10 | Operabilitas: `/healthz` dan cadangan lewat NFR-HRS-06, NFR-HRS-14 |
| BR-HRS-21 | FR-HRS-01.10, FR-HRS-02.4, FR-HRS-02.10 | Hipotesis INT-1 dan INT-2 (2.5) |
| BR-HRS-22 | (tanpa FR; dipenuhi 11.1 dan 11.3 dokumen ini) | Laporan status selaras kode |

### 11.3 Kontradiksi dokumen status vs kode

Bagian ini mencatat klaim dokumen status, ADR, dan `CLAUDE.md` yang **bertentangan dengan kode**. Klaim tersebut tidak dipakai sebagai fakta di bagian lain dokumen ini. Kolom `Bukti` menunjuk kode atau berkas yang menunjukkan kenyataan; dokumen sumber klaim disebut di kolom pertama.

#### 11.3.1 Cacat teknis yang terverifikasi dari kode

| Butir | Kenyataan menurut kode | Bukti |
|---|---|---|
| Migrasi absensi 001 tidak valid | Baris 25 `CONSTRAINT attendance_unique_per_day UNIQUE (tenant_id, employee_id, date) WHERE NOT is_override` bukan sintaks PostgreSQL yang sah (klausa `WHERE` hanya untuk indeks unik). Baris 45-46 `GRANT ... TO app` memakai peran `app`, bukan `hris_app`; skrip init hanya membuat `hris_app` dan `hris_admin`. RLS `ENABLE` tanpa `FORCE` dan kebijakan tanpa `WITH CHECK`. Peran `app` dipakai juga oleh migrasi leave dan notification. | `services/attendance-service/migrations/001_create_attendance_schema.up.sql`, `deploy/docker/init-postgres.sql` |
| Indeks duplikat | Absensi 002 (`idx_attendance_employee_date`) dan leave 002 (`idx_leave_requests_employee_id`, `_status`, `_approved_by`) membuat indeks bernama sama dengan migrasi 001 tanpa `IF NOT EXISTS` (`Pending Verification`, belum dijalankan) | `services/attendance-service/migrations/002_add_performance_indexes.up.sql`, `services/leave-service/migrations/002_add_performance_indexes.up.sql` |
| Tidak ada HTTP mux yang dipasang (teratasi 2026-11-14) | Semula: tidak ada `runtime.NewServeMux` atau `ListenAndServe` di layanan mandiri; jalur grpc-gateway yang dipanggil gateway Laravel tidak pernah disajikan. ADR-0005 menghapus gateway, gRPC, dan grpc-gateway; `cmd/api` kini menyajikan REST pada satu `net/http` ServeMux | `apps/api/cmd/api/routes.go`, `apps/api/cmd/api/main.go`, `docs/adr/0005-single-module-go-api-on-sela-stack.md` |
| Login web adalah tiruan (teratasi 2026-11-14) | `apps/web/src/routes/(auth)/login/+page.server.ts` kini mengirim POST ke `${apiUrl()}/api/v1/auth/login` dan menulis cookie `hris_session` lewat `session.write`; panggilan `localhost:3000` dan fungsi store `setSession` yang tidak terpakai sudah hilang | `apps/web/src/routes/(auth)/login/+page.server.ts`, `apps/web/src/lib/server/session.ts` |
| Klien gateway tidak menyampaikan token (teratasi 2026-11-14) | Semula: `getBearerToken()` mengembalikan string kosong di klien gateway Laravel. Gateway dihapus (ADR-0005); proxy BFF SvelteKit kini menyertakan bearer token dari cookie `hris_session` | `apps/web/src/routes/api/[...path]/+server.ts`, `docs/adr/0005-single-module-go-api-on-sela-stack.md` |
| Dokumen: RPC kosong | `GetPresignedUrl`, `RollbackDocumentVersion` kosong (`document_service.go` baris 215-221) dan `DownloadDocument` juga placeholder; `UploadDocument` selalu `Unauthenticated` karena `actor_id` tidak pernah diisi | `services/document-service/internal/interfaces/grpc/document_service.go`, `services/document-service/internal/interfaces/grpc/mapping.go` |
| Notifikasi tidak terkirim | `cmd/server/main.go` baris 112-132 menetapkan `_ = emailAdapter` dan `_ = inAppAdapter` | `services/notification-service/cmd/server/main.go` |
| Payload peristiwa kosong | Peristiwa leave memuat field privat tanpa `MarshalJSON`, sehingga `payload` `{}`; `ApplyLeave` tidak mengisi `DaysCount` | `services/leave-service/internal/infrastructure/nats/event_publisher.go`, `services/leave-service/internal/domain/events.go` |
| Bentrok port gRPC | leave-service dan audit-service sama-sama mengikat `50054` secara bawaan | `services/leave-service/cmd/server/main.go`, `services/audit-service/cmd/server/main.go` |
| OTel/metrik tidak aktif | `InitTracer` dan `NewMetricsRegistry` hanya didefinisikan di `_shared`, tidak dipanggil layanan; `StartMetricsServer` mengikat `:8081` untuk semua layanan | `services/_shared/observability/otel.go`, `services/_shared/observability/metrics.go` |
| Versi Go tidak seragam | `go.work` dan semua `go.mod` Go 1.25.0, CI `GO_VERSION: '1.24'` | `go.work`, `.github/workflows/ci-go.yml` |

#### 11.3.2 Dokumen status yang bertentangan dengan kode

| Dokumen (klaim) | Klaim | Kenyataan menurut kode | Bukti |
|---|---|---|---|
| `PROJECT_PROGRESS.md` | "Phase 1 Foundation Nearly Complete" (baris 4) dan blok "100% COMPLETE / Production-ready" per layanan (baris 10-68), termasuk notifikasi | Layanan tidak dapat dijalankan end-to-end: migrasi absensi tidak valid, tidak ada HTTP mux, login web tiruan, notifikasi tidak dipasang | `PROJECT_PROGRESS.md`, `services/attendance-service/migrations/001_create_attendance_schema.up.sql` |
| `PHASE1_RELEASE.md` | "PRODUCTION READY" (baris 4) dan sertifikasi rilis 2026-06-02 (baris 346) | idem; juga tanpa autentikasi layanan dan PII polos | `PHASE1_RELEASE.md`, `services/_shared/server/grpc_interceptors.go` |
| `IMPLEMENTATION_COMPLETE.md` | "PRODUCTION READY", "Services (5/5 Complete)", "Production Hardening (8/8 Complete)", "Real email delivery with 5 templates", OpenTelemetry | Adaptor email tidak terpasang dan templat tidak dimuat; ada enam templat HTML; OTel tidak pernah diinisialisasi | `IMPLEMENTATION_COMPLETE.md`, `services/notification-service/cmd/server/main.go` |
| `PHASE2_STATUS.md` | Audit "Completion: 50% (Layers 1-3 Complete)" | Kode audit memiliki lapisan domain, aplikasi, infrastruktur, antarmuka, dan konsumen | `PHASE2_STATUS.md`, `services/audit-service/internal/interfaces/grpc/audit_service.go` |
| `PHASE2_FOUNDATION.md` | Dokumen fondasi fase 2 | Dibaca hanya sebagai rencana; tidak ada kode fase 2 (payroll) di repo, konsisten dengan ADR-0004 | `PHASE2_FOUNDATION.md`, `docs/adr/0004-defer-payroll-to-phase2.md` |
| `docs/PHASE5_SUMMARY.md` (2026-06-08) | Kesiapan 45/100; cuti "belum diimplementasi"; gateway dan frontend kosong | Kini usang ke arah sebaliknya: handler cuti ada, gateway memiliki rute auth dan cuti, web memiliki empat halaman. Skor 45/100 sendiri bukan ukuran yang diverifikasi | `docs/PHASE5_SUMMARY.md`, `apps/api-gateway/routes/api.php` |
| `CLAUDE.md` bagian 2 | "sqlc + pgx untuk semua akses DB" | sqlc hanya dipakai di auth-service | `CLAUDE.md`, `services/auth-service/cmd/server/main.go` |
| `CLAUDE.md` bagian 3 | Autentikasi gateway: Laravel Sanctum mode token dengan JWT RSA, "lihat ADR-0005" | Sanctum dan spatie/permission terpasang di `composer.json` tetapi tidak dipakai; ADR-0005 tidak ada (hanya `docs/adr/0001..0004`); token tidak diteruskan ke layanan (`getBearerToken()` kosong) | `CLAUDE.md`, `apps/api-gateway/composer.json`, `apps/api-gateway/app/Services/BaseGrpcGatewayClient.php` |
| `CLAUDE.md` bagian 6 | Semua peristiwa berskema amplop dengan `event_id` UUID v7; publikasi setelah commit (outbox bila perlu) | Tidak ada outbox; peristiwa dipublikasikan langsung ke NATS; jenis id belum diverifikasi (`Pending Verification`) | `CLAUDE.md`, `services/leave-service/internal/infrastructure/nats/event_publisher.go` |
| `CLAUDE.md` bagian 9 | span OpenTelemetry per layanan | `InitTracer` tidak pernah dipanggil | `CLAUDE.md`, `services/_shared/observability/otel.go` |
| `docs/adr/0001-grpc-gateway-over-grpc-php.md` | grpc-gateway dijalankan di setiap layanan | tidak ada mux HTTP di layanan mana pun | `docs/adr/0001-grpc-gateway-over-grpc-php.md`, `services/auth-service/cmd/server/main.go` |
| `docs/adr/0004-defer-payroll-to-phase2.md` | Rute fase 2 di gateway mengembalikan 501 | Rute tersebut tidak ada di `api.php` | `docs/adr/0004-defer-payroll-to-phase2.md`, `apps/api-gateway/routes/api.php` |

Catatan: ADR-0004 menyebut PDPA satu kali (baris 47). Pernyataan "tidak ada penyebutan UU PDP" dihaluskan menjadi: tidak ada penanganan UU PDP (klasifikasi, retensi, persetujuan, hak subjek) di repo.

#### 11.3.3 Koreksi atas klaim awal pindai asisten

Klaim berikut berasal dari pemindaian awal dan dikoreksi setelah dibaca ulang terhadap kode. Dokumen suite yang memakainya harus mengikuti koreksi ini.

| Klaim awal | Koreksi | Bukti |
|---|---|---|
| `server.exe` ter-commit di repo | Tidak dilacak git: `*.exe` diabaikan (`.gitignore` baris 7) dan `git ls-files` tidak memuat `server.exe`; berkas hanya ada di pohon kerja. Pembersihan berlaku untuk pohon kerja dan artefak build | `.gitignore` |
| Token disimpan ganda di cookie dan localStorage | Hanya cookie diisi (`auth_token` HttpOnly dan `auth_user` bukan HttpOnly); `setSession` tidak punya pemanggil | `apps/web/src/lib/stores/auth.ts`, `apps/web/src/routes/(auth)/login/+page.server.ts` |
| Dokumen 5 dari 7 RPC berisi | 4 dari 7 berisi (`UploadDocument`, `ListDocuments`, `GetDocumentMetadata`, `DeleteDocument`); `DownloadDocument` juga placeholder | `services/document-service/internal/interfaces/grpc/document_service.go` |
| Peristiwa `user.registered` membuat kerangka karyawan | Konsumen ada tetapi tidak dapat bekerja (payload tanpa `tenant_id`, id departemen/jabatan acak, galat di-ack lalu hilang) | `services/employee-service/internal/infrastructure/nats_consumer.go` |
| Saldo dan tumpang tindih cuti tidak ditegakkan | Saldo ditegakkan di kode aplikasi; tumpang tindih tidak | `services/leave-service/internal/application/commands/create_leave_request.go` |
| `deploy/k8s/base` kosong | `deploy/k8s` hanya berisi direktori kosong (`base`, `overlays/production`, `overlays/staging`); `git ls-files deploy` hanya memuat empat berkas `deploy/docker` | `deploy/docker/docker-compose.yml` |
| CD mengabaikan beberapa layanan | CD merujuk Dockerfile yang tidak ada (hanya `Dockerfile.base`, `services/auth-service/Dockerfile`, `services/employee-service/Dockerfile` yang ada); `document-service` tidak ada di matriks CI/CD; uji asap memanggil `/health`, bukan `/api/health`; matriks build CI memuat `ai-service` yang tidak memiliki titik masuk | `.github/workflows/cd-staging.yml`, `.github/workflows/ci-go.yml` |
| Tidak ada penyebutan UU PDP sama sekali | ADR-0004 menyebut PDPA satu kali; tidak ada penanganan UU PDP | `docs/adr/0004-defer-payroll-to-phase2.md` |
| CI tidak menguji migrasi | CI menjalankan `up -> down -> up` dan uji integrasi untuk auth, employee, leave saja, sebagai `hris_admin` | `.github/workflows/ci-go.yml` |

---

## 12. Catatan Penutup

1. **Perubahan dokumen.** Dokumen ini adalah versi pertama. Setiap perubahan kode HRIS yang memengaruhi FR di bagian 4 harus memperbarui 11.1 dan, bila ada kontradiksi baru, 11.3 (konvensi STACK-SPEC 18: FSD diperbarui bersama kode).
2. **Estimasi.** Seluruh estimasi usaha pada 2.7 adalah `[INFERENCE]`. Dokumen ini tidak memuat persentase kesiapan.
3. **Nilai `tbd`.** Nilai yang tidak ditemukan di repo dan menunggu keputusan pemilik: tier/harga/batas `hrs.employees.max` (Q-HRS-02), TTL sesi, ambang kunci akun, dan toleransi entitlement (Q-HRS-11), aturan kerja dan kebijakan cuti bawaan (Q-HRS-05, Q-HRS-06), retensi dan dasar hukum UU PDP (Q-HRS-04, Q-HRS-14), target ketersediaan/latensi/RPO/RTO (Q-HRS-07), observabilitas OTel (Q-HRS-08), penyedia LLM (Q-HRS-09), jadwal tinjauan hukum payroll (Q-HRS-10), hosting produksi (Q-HRS-12), penyedia SMTP/kanal tambahan (Q-HRS-13).
4. **Urutan kerja.** Perbaikan cacat pada 11.3.1 dan gerbang stabilisasi BRD 12 mendahului penyelarasan stack (D-HRS-09); penyelarasan HRIS dijadwalkan sesudah POS (tahap S3) dan satu penyelarasan dikerjakan pada satu waktu.
5. **Modul ditunda.** Payroll, pajak, rekrutmen, dan kinerja tidak memiliki FR; FR baru hanya boleh ditulis setelah prasyarat 4.09-4.12 terpenuhi dan ADR terkait dibuat.
6. **Keterbatasan bukti.** Seluruh bukti berasal dari pembacaan statis; klaim kegagalan runtime ditandai `Pending Verification` sampai dijalankan pada gerbang stabilisasi.
