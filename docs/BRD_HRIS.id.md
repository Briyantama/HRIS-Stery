# BRD HRIS-Stery

*Business Requirements Document (BRD): sistem informasi sumber daya manusia (HRIS) multi-tenant untuk perusahaan di Indonesia, dijual sebagai produk `HRS` secara mandiri atau sebagai bagian dari Stery Suite*

|                         |                                                                                                                                                                                                                                                                  |
|-------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Nama Dokumen**        | BRD - HRIS-Stery (SKU `HRS`)                                                                                                                                                                                                                                     |
| **Versi**               | 1.0 (Draft)                                                                                                                                                                                                                                                      |
| **Tanggal**             | 6 November 2026                                                                                                                                                                                                                                                  |
| **Disusun oleh**        | Tim Produk & Engineering                                                                                                                                                                                                                                         |
| **Status**              | Draft untuk ditinjau                                                                                                                                                                                                                                             |
| **Dokumen Terkait**     | PRD HRIS-Stery (`PRD_HRIS.id.md`), FSD HRIS-Stery (`FSD_HRIS.id.md`), `00_Analisis_Dampak_Risiko`, BRD/PRD/FSD Stery Suite (folder `stery-suite/docs`), `SELA/CLAUDE.md` (standar stack), ADR-0001 s.d. ADR-0004 (`HRIS-Stery/docs/adr`)                              |
| **Perubahan dari vX**   | Dokumen pertama                                                                                                                                                                                                                                                  |

## 1. Ringkasan Eksekutif

HRIS-Stery adalah aplikasi HRIS multi-tenant: satu perusahaan (tenant) mendaftar, mengelola data karyawan, absensi harian, dan cuti, lalu HR dan atasan bekerja dengan alur persetujuan serta jejak audit. Repositori ini belum pernah memiliki BRD, PRD, atau FSD; dokumen ini adalah BRD pertama dan ditulis dari pembacaan kode (kebenaran = kode), bukan dari dokumen status lama.

Posisi produk dalam portofolio: HRIS-Stery dijual sebagai SKU `HRS`, dapat berdiri sendiri (`STERY_MODE=standalone`, bawaan) atau terhubung ke platform Stery Suite (`STERY_MODE=suite`) bersama POS-Stery (SKU `POS`) dan SELA (SKU `SEL`). Seluruh stack, pola, dan arsitektur HRIS-Stery akan diselaraskan dengan standar SELA (keputusan pemilik, lihat 9.3 dan FSD 2.1) sehingga satu tim dapat mengoperasikan ketiga produk dengan cara yang sama.

Temuan inti dari pembacaan kode (status per modul di bagian 13): kode HRIS-Stery sudah memiliki arsitektur berlapis (domain, application, infrastructure, interfaces) di tujuh layanan Go dan 39 dari 47 RPC proto memiliki handler berisi, tetapi **belum ada satu pun alur ujung-ke-ujung yang terbukti berjalan**: migrasi modul absensi memuat sintaks tidak valid (`services/attendance-service/migrations/001_create_attendance_schema.up.sql:25`), tidak ada server HTTP yang melayani jalur grpc-gateway yang dipanggil gateway Laravel, halaman login web dimock ke `http://localhost:3000/api/auth/login`, tidak ada satu pun layanan yang memeriksa autentikasi atau otorisasi, dan NIK/NPWP/tanggal lahir tersimpan sebagai teks polos. Dokumen status lama (`PROJECT_PROGRESS.md`, `PHASE1_RELEASE.md`, `IMPLEMENTATION_COMPLETE.md`) menyatakan "100% / production ready" dan bertentangan dengan kode (daftar di FSD 11.3). Karena itu pekerjaan sebelum penjualan dibagi tiga: (1) gerbang stabilisasi HRIS, (2) penyelarasan stack ke standar SELA, (3) kesiapan mode suite.

Modul payroll, aturan pajak, rekrutmen, dan kinerja **ditunda** sesuai ADR-0004 (diterima 2026-05-30) dan baru boleh dimulai setelah tinjauan hukum atas PPh21, BPJS, dan THR. Modul asisten AI belum memiliki kode (hanya kontrak proto).

Harga, tier, SLA, ukuran pasar, dan data kompetitor tidak ada di repositori mana pun dan dicatat `tbd` dengan pertanyaan terbuka (9.4).

## 2. Latar Belakang dan Pernyataan Masalah

Perusahaan di Indonesia mengelola karyawan, kehadiran, dan cuti dengan campuran lembar kerja, pesan instan, dan sistem terpisah. Rumusan masalah berikut adalah hipotesis produk dan **belum divalidasi riset pengguna** `[INFERENCE]` (pertanyaan riset: Q-HRS-01, Q-HRS-03):

- Data induk karyawan tersebar dan tidak punya satu sumber kebenaran; perubahan jabatan dan atasan tidak terdokumentasi.
- Pengajuan dan persetujuan cuti lewat chat sulit dilacak, saldo cuti dihitung manual, dan tidak ada riwayat keputusan.
- Pencatatan kehadiran tidak seragam dan koreksi tidak meninggalkan jejak.
- Dokumen karyawan (kontrak, identitas) tersimpan di folder bersama tanpa kontrol akses dan versi.
- Data pribadi karyawan (NIK, NPWP, tanggal lahir) memerlukan pelindungan sesuai UU Pelindungan Data Pribadi No. 27 Tahun 2022 (UU PDP), sedangkan repositori ini belum memuat satu pun kebijakan atau mekanisme pelindungan (FSD 8).

Dari sisi pemilik produk, masalah tambahan: pemilik menjual tiga produk (HRIS, POS, SELA) dan ingin memasarkannya sebagai paket sendiri-sendiri atau bundel; HRIS-Stery saat ini memakai stack berbeda dari SELA (gateway Laravel, SSR SvelteKit, JWT, NATS, golang-migrate, sqlc, zap, MinIO, Redis) sehingga tidak dapat dioperasikan seragam.

Target skala desain yang tercatat di ADR-0002 adalah "ratusan hingga rendah ribuan tenant" (`docs/adr/0002-postgres-rls-for-tenant-isolation.md`); angka itu adalah asumsi desain, bukan target bisnis.

## 3. Tujuan Bisnis

1. Menyediakan HRIS multi-tenant yang aman dipakai perusahaan nyata: data induk karyawan, absensi, cuti, notifikasi, audit, dan dokumen (BR-HRS-01 s.d. BR-HRS-10).
2. Menjamin isolasi data antar tenant dan pelindungan data pribadi sesuai UU PDP sebagai syarat penjualan, bukan fitur tambahan (BR-HRS-02, BR-HRS-12).
3. Menjadikan HRIS-Stery produk yang siap dijual berdiri sendiri maupun dalam bundel suite tanpa mengubah perilaku standalone (BR-HRS-15, BR-HRS-16).
4. Membawa kondisi kode dari "terpasang sebagian" ke "terbukti berjalan" lewat gerbang stabilisasi sebelum pelanggan nyata (BR-HRS-17).
5. Menyelaraskan stack ke standar SELA agar biaya operasi dan pengembangan tiga produk rendah (BR-HRS-18).
6. Mengelola fitur berisiko hukum (payroll, pajak) dengan disiplin: tidak dirilis sebelum tinjauan hukum (BR-HRS-13, BR-HRS-14).
7. Menjaga kejujuran status: laporan kemajuan harus selaras dengan kode (BR-HRS-22).

## 4. Ruang Lingkup

### 4.1 Segmen dan Paket

| **Aspek**                  | **Nilai**                                                                                                                        |
|----------------------------|----------------------------------------------------------------------------------------------------------------------------------|
| Segmen target              | `tbd` (tidak ada pernyataan segmen di repositori; Q-HRS-01)                                                                      |
| SKU produk                 | `HRS`                                                                                                                            |
| Bundel (didefinisikan di BRD Stery Suite) | `HRS+POS`, `HRS+SEL`, `SUITE` (ketiganya)                                                                          |
| Mode operasi               | `standalone` (bawaan; pengguna, sesi, dan data sendiri) atau `suite` (identitas, entitlement, penagihan, notifikasi, audit didelegasikan ke platform) |
| Dimensi batas              | `hrs.employees.max` (nilai per tier `tbd`; Q-HRS-02)                                                                             |
| Tier                       | `tbd` (jumlah dan nama tier; Q-HRS-02)                                                                                           |

### 4.2 Dalam Lingkup

Dalam lingkup gelombang pertama (modul katalog 01-08, lihat FSD 4):

- 01 AUTH: registrasi tenant, login, sesi, RBAC (peran `hr_admin`, `manager`, `employee` dan izin `resource:action`).
- 02 EMP: karyawan, departemen, jabatan, bagan organisasi, terminasi.
- 03 ATT: check-in/check-out, koreksi oleh HR, daftar dan ringkasan kehadiran.
- 04 LEV: jenis cuti, saldo, pengajuan, persetujuan, penolakan, pembatalan.
- 05 NTF: notifikasi email dan in-app berbasis peristiwa.
- 06 AUD: jejak audit append-only.
- 07 DOC: dokumen karyawan (unggah lewat URL bertanda tangan, versi, klasifikasi, retensi).
- 08 AI: asisten AI (analisis resume, prediksi attrisi, chat); dalam lingkup sebagai kontrak, belum dikerjakan (status Belum).
- Pelindungan data pribadi UU PDP untuk seluruh modul (FSD 8).
- Gerbang stabilisasi HRIS, penyelarasan stack ke standar SELA, dan adaptor mode suite (bagian 12).

### 4.3 Ditunda

| **Modul** | **Alasan penundaan**                                                                                                                                                                                          | **Prasyarat sebelum dimulai**                                                                                                             |
|-----------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------|
| 09 PAY payroll | Perhitungan gaji Indonesia diatur hukum dan berubah tiap tahun (PPh21, BPJS Kesehatan, BPJS Ketenagakerjaan, THR, pesangon, UMP). Kesalahan adalah pelanggaran regulasi, bukan bug (ADR-0004)               | Tinjauan hukum oleh profesional pajak Indonesia atas logika PPh21/BPJS/THR (Q-HRS-10); sprint desain aturan pajak; ADR `tax-rules-service`; ADR imutabilitas `PayrollRun`; desain saga lintas modul |
| 10 TAX aturan pajak | Aturan berversi menurut tanggal efektif; bagian dari payroll (ADR-0004)                                                                                                                                  | Sama dengan 09                                                                                                                            |
| 11 REC rekrutmen | Memperkenalkan konteks baru `Candidate`; retensi data kandidat gagal adalah liabilitas UU PDP; transisi `Candidate -> Employee` perlu desain (ADR-0004)                                                | Desain domain kandidat, kebijakan retensi kandidat dengan tinjauan hukum, keputusan penyedia LLM untuk parsing resume (Q-HRS-09)          |
| 12 PRF kinerja | Memerlukan siklus penilaian, multi-penilai, kalibrasi, dan terkait kompensasi sehingga bergantung pada payroll (ADR-0004)                                                                                  | Payroll selesai dan tinjauan hukum; desain kerangka kompetensi                                                                            |

Aturan yang berlaku selama penundaan (ADR-0004 bagian Compliance): tidak ada logika payroll, pajak, rekrutmen, atau kinerja yang boleh digabung ke cabang mana pun; pull request yang memuatnya ditolak dengan merujuk ADR-0004.

### 4.4 Di Luar Lingkup

Butir berikut diusulkan di luar lingkup fase ini dan menunggu persetujuan pemilik produk (bagian 14):

- Perhitungan gaji, pajak, BPJS, THR, pesangon (lihat 4.3).
- Aplikasi seluler native; akses lewat web (SPA) saja.
- Integrasi ke sistem pemerintah (BPJS online, e-filing pajak, Kemnaker) selama payroll ditunda.
- Penagihan standalone berbasis kontrak otomatis di dalam HRIS; penagihan suite dilakukan platform (BRD Stery Suite).
- Fitur khusus enterprise (SSO korporat pihak ketiga, kontrak tahunan) sampai ada permintaan terbukti (Q-HRS-01).

## 5. Pemangku Kepentingan

| **Peran**                                   | **Deskripsi Kepentingan**                                                                                     | **Tingkat Pengaruh** |
|---------------------------------------------|---------------------------------------------------------------------------------------------------------------|----------------------|
| Pemilik Produk / Sponsor Bisnis             | Menetapkan segmen, harga, tier, dan prioritas fase; menyetujui dokumen ini                                    | Tinggi               |
| Admin HR tenant (`hr_admin`)                | Mengelola karyawan, cuti, koreksi absensi, dokumen; pengguna harian terberat                                   | Tinggi               |
| Manajer / atasan langsung (`manager`)       | Menyetujui cuti dan melihat kehadiran tim                                                                      | Menengah             |
| Karyawan (`employee`)                       | Check-in/out, mengajukan cuti, melihat saldo dan notifikasi                                                    | Menengah             |
| Admin platform (operator Stery)             | Mengelola tenant, entitlement, dan dukungan; di mode suite bertindak lewat konsol platform                     | Tinggi               |
| Tim Engineering                             | Membangun, menstabilkan, dan menyelaraskan stack; menanggung operasi                                           | Tinggi               |
| Tim Legal / Compliance                      | Menetapkan dasar pemrosesan, retensi, hak subjek data (UU PDP) dan menyetujui sebelum payroll dimulai          | Tinggi               |
| Pemilik data pribadi (karyawan tenant)      | Subjek data UU PDP: berhak atas akses, koreksi, dan penghapusan sesuai ketentuan                              | Menengah             |
| Tim Bisnis / Pemasaran                      | Penetapan harga per tier dan bundel, riset kompetitor                                                          | Menengah             |

## 6. Kebutuhan Bisnis (Business Requirements)

Prioritas: M = Must have, S = Should have, C = Could have. Butir yang ditunda ditandai (Ditunda) dan tetap tercatat agar jejak keputusan ada. Kolom Modul merujuk katalog modul FSD 4 (SEC = lintas modul, bagian FSD 8).

| **ID**    | **Kebutuhan Bisnis**                                                                                                                                                                                                  | **Prioritas** | **Modul**       |
|-----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------|-----------------|
| BR-HRS-01 | Sebuah perusahaan (tenant) dapat mendaftar sendiri dan langsung memiliki admin HR pertama dengan peran dan izin bawaan.                                                                                               | M             | 01 AUTH         |
| BR-HRS-02 | Data satu tenant tidak pernah dapat dibaca atau ditulis tenant lain; isolasi ditegakkan di basis data (RLS) dan dibuktikan uji lintas tenant di setiap modul.                                                         | M             | SEC, 01-07      |
| BR-HRS-03 | Hak akses berbasis peran (HR admin, manager, karyawan) ditegakkan di backend pada setiap endpoint, bukan hanya disembunyikan di antarmuka.                                                                            | M             | 01 AUTH         |
| BR-HRS-04 | HR admin dapat mengelola data induk: membuat, mengubah, dan menerminasi karyawan; mengelola departemen (berhierarki) dan jabatan; melihat bagan organisasi.                                                          | M             | 02 EMP          |
| BR-HRS-05 | Karyawan dapat mencatat kehadiran harian (check-in/check-out); HR dapat mengoreksi catatan dengan alasan dan jejak audit; HR dan atasan melihat daftar serta ringkasan kehadiran.                                     | M             | 03 ATT          |
| BR-HRS-06 | Aturan jam kerja, ambang keterlambatan, hari kerja, zona waktu tenant, dan (opsional) lokasi absensi dapat dikonfigurasi per tenant; nilai bawaan `tbd` (Q-HRS-05).                                                  | S             | 03 ATT          |
| BR-HRS-07 | Karyawan dapat mengajukan cuti; atasan atau HR menyetujui atau menolak; hari kerja dihitung sistem; saldo selalu konsisten; pengajuan tumpang tindih ditolak; pembatalan memulihkan saldo.                              | M             | 04 LEV          |
| BR-HRS-08 | Pihak terkait menerima notifikasi email dan in-app untuk peristiwa penting (pengajuan dan keputusan cuti, karyawan baru, aktivasi akun) dan dapat menandainya sudah dibaca.                                           | S             | 05 NTF          |
| BR-HRS-09 | Tindakan sensitif (login, perubahan karyawan, koreksi absensi, keputusan cuti, akses dokumen rahasia, perubahan PII) tercatat di jejak audit append-only yang dapat dikueri HR atau auditor dan tidak dapat diubah.   | M             | 06 AUD          |
| BR-HRS-10 | Dokumen karyawan disimpan aman: unggah dan unduh lewat URL bertanda tangan berumur pendek, pemindaian malware, versi dan pemulihan versi, klasifikasi, serta retensi.                                                 | S             | 07 DOC          |
| BR-HRS-11 | Asisten AI (analisis resume, prediksi attrisi, chat HR) dapat ditawarkan setelah kebijakan pemrosesan PII ke penyedia LLM, batas biaya per tenant, dan persetujuan ditetapkan (Q-HRS-09).                              | C             | 08 AI           |
| BR-HRS-12 | Data pribadi karyawan (NIK, NPWP, tanggal lahir, lokasi absensi, alasan cuti) dilindungi sesuai UU PDP: klasifikasi, enkripsi, dasar pemrosesan dan persetujuan, retensi, hak subjek data, prosedur insiden.         | M             | SEC             |
| BR-HRS-13 | Payroll, aturan pajak (PPh21), BPJS, dan THR ditunda sampai tinjauan hukum dan prasyarat ADR-0004 selesai; tidak ada logika payroll dirilis sebelum itu. (Ditunda)                                                     | C             | 09 PAY, 10 TAX  |
| BR-HRS-14 | Rekrutmen dan penilaian kinerja ditunda sesuai ADR-0004. (Ditunda)                                                                                                                                                    | C             | 11 REC, 12 PRF  |
| BR-HRS-15 | HRIS dapat dijual dan dijalankan berdiri sendiri maupun sebagai bagian Stery Suite; perilaku `standalone` tidak berubah oleh keberadaan mode `suite`.                                                                 | M             | SEC, 01         |
| BR-HRS-16 | Penjualan dibatasi dimensi `hrs.employees.max`; tier dan harga `tbd` (Q-HRS-02); pelampauan batas atau berakhirnya entitlement ditangani tanpa kehilangan data (hanya pembuatan karyawan baru yang ditolak).        | S             | 02 EMP          |
| BR-HRS-17 | Produk harus lulus gerbang stabilisasi HRIS (bagian 12) sebelum dijual atau dipakai pelanggan nyata.                                                                                                                   | M             | SEC             |
| BR-HRS-18 | Produk diselaraskan ke stack standar suite (SELA) agar operasi, keamanan, dan pengembangan seragam antar produk.                                                                                                       | M             | SEC             |
| BR-HRS-19 | Antarmuka tersedia dalam Bahasa Indonesia (bawaan) dan Inggris; zona waktu WIB; mata uang IDR bila ada nilai uang.                                                                                                      | S             | SEC             |
| BR-HRS-20 | Layanan dapat dioperasikan dan dipulihkan: `/healthz`, log terstruktur tanpa PII, pembatasan laju dan penguncian akun, cadangan data (RPO/RTO `tbd`, Q-HRS-07).                                                       | M             | SEC             |
| BR-HRS-21 | Hipotesis integrasi suite (validasi permintaan dulu): INT-1 tautan staf karyawan HRIS dengan pengguna POS (karyawan terminasi menonaktifkan pengguna POS); INT-2 acara gathering perusahaan membuat acara SELA.      | C             | 02 EMP          |
| BR-HRS-22 | Laporan status implementasi wajib selaras dengan kode dan memuat bukti; dokumen status lama yang bertentangan dengan kode dikonsolidasikan atau dicabut.                                                               | S             | SEC             |

## 7. Model Bisnis dan Monetisasi

Seluruh nilai uang HRS adalah `tbd` (tidak ada harga, tier, atau mekanisme penagihan di repositori; Q-HRS-02). Model yang diusulkan, mengikuti keputusan suite D5:

| **Tier**                       | **Batas karyawan (`hrs.employees.max`)** | **Harga** | **Siklus tagih** | **Trial** | **Catatan**                                                                                   |
|--------------------------------|------------------------------------------|-----------|------------------|-----------|-----------------------------------------------------------------------------------------------|
| `tbd` (jumlah dan nama tier)   | `tbd`                                    | `tbd`     | `tbd`            | `tbd`     | Satu dimensi batas; nilai dan tier diputuskan pemilik (Q-HRS-02)                              |
| Bundel `HRS+POS`, `HRS+SEL`, `SUITE` | mengikuti tier HRS                | `tbd`     | `tbd`            | `tbd`     | Aturan diskon bundel `tbd` (dikelola di BRD Stery Suite)                                      |

- Penagihan pada mode `suite` dilakukan platform memakai abstraksi `payment.Provider` SELA (Midtrans QRIS); penagihan pada mode `standalone` belum ada dan mekanismenya `tbd` (kontrak manual atau platform billing; Q-HRS-02).
- Hak pakai fitur HRS diterapkan lewat entitlement: kunci fitur modul dan dimensi `hrs.employees.max`; pada pelampauan batas, pembuatan karyawan baru ditolak sedangkan data yang ada tetap dapat dibaca dan diekspor.
- Modul payroll, pajak, rekrutmen, dan kinerja (ditunda) berpotensi menjadi add-on berbayar; keputusan `tbd`.

## 8. Analisis Alternatif dan Kompetitor

### 8.1 Kompetitor

Seluruh isi bagian ini `tbd`: tidak ada riset kompetitor di repositori HRIS-Stery (`deep-research-report.md` adalah rencana teknis, bukan riset pasar).

| **Kompetitor**     | **Bukti skala pasar** | **Model harga** | **Fokus segmen** | **Fitur payroll/PPh21** | **Lokalisasi Indonesia** |
|--------------------|-----------------------|-----------------|------------------|-------------------------|--------------------------|
| `tbd` (daftar dari riset) | `tbd`          | `tbd`           | `tbd`            | `tbd`                   | `tbd`                    |

Pertanyaan riset (Q-HRS-03): (1) HRIS apa yang dipakai target segmen sekarang dan berapa biayanya per karyawan per bulan? (2) Apakah pembeli menuntut payroll/PPh21 di tahap pertama atau cukup data induk, absensi, dan cuti? (3) Apakah absensi dengan geofence atau swafoto menjadi syarat beli? (4) Bagaimana pembeli menilai penyimpanan data karyawan di cloud terhadap UU PDP? (5) Apakah bundel dengan POS atau SELA bernilai bagi pembeli HRS?

### 8.2 Alternatif Arsitektur

Alternatif arsitektur tingkat portofolio (A polyrepo + platform, B monorepo, C satu basis kode) dibahas di BRD Stery Suite bagian 8 dan `00_Analisis_Dampak_Risiko` bagian 5; keputusan: opsi A (D-HRS-01). Alternatif tingkat produk yang dipertimbangkan untuk HRIS:

| **Alternatif**                                              | **Penilaian**                                                                                                                                                     |
|-------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Mempertahankan stack saat ini (Laravel + banyak layanan gRPC + NATS) | Ditolak: bertentangan dengan keputusan pemilik "stack = SELA"; gateway tidak berfungsi (FSD 11.3); biaya operasi tinggi untuk tim kecil                  |
| Menulis ulang dari nol di stack SELA                        | Ditolak sebagai default: lapisan domain dan application dapat dipakai ulang; yang ditulis ulang adalah infrastruktur dan antarmuka (FSD 2.7)                         |
| Menyelaraskan bertahap: pertahankan domain/application, tulis ulang infrastruktur dan antarmuka | Dipilih (D-HRS-11, D-HRS-12)                                                                                                             |
| Membeli/mengintegrasikan HRIS pihak ketiga                  | `tbd` bergantung hasil riset Q-HRS-03; tidak dievaluasi di dokumen ini                                                                                            |

## 9. Asumsi dan Batasan

### 9.1 Asumsi

- Pengguna mengakses lewat peramban modern; tidak ada kebutuhan aplikasi native pada fase ini (usulan, 4.4).
- Satu tenant = satu perusahaan; hubungan satu organisasi suite dengan satu tenant HRS dijelaskan di FSD Stery Suite 7.5.
- Tenant bertindak sebagai pengendali data dan Stery sebagai prosesor data atas data karyawan; pembagian peran dan perjanjian pemrosesan data `tbd` (Q-HRS-14) `[INFERENCE]`.
- Tim pengembangan kecil sehingga satu penyelarasan dikerjakan pada satu waktu (keputusan suite D12).

### 9.2 Batasan

- Payroll, pajak, rekrutmen, kinerja tidak boleh dikerjakan sebelum prasyarat ADR-0004 (4.3).
- Stack target mengikuti SELA: Go >= 1.25.1, PostgreSQL 18, pgx/v5 tanpa ORM, goose, Dragonfly, RustFS, SvelteKit SPA, tanpa JWT dan tanpa NATS (FSD 2.1).
- Perilaku `standalone` harus tetap utuh ketika platform tidak ada (keputusan D2).
- Tidak ada angka harga, SLA, kuota, KPI, atau data kompetitor yang boleh dikarang; semuanya `tbd` sampai pemilik memutuskan.
- Dokumen ini tidak mengubah kode apa pun; migrasi stack dispesifikasikan sebagai alur kerja roadmap.

### 9.3 Keputusan yang Berlaku

Keputusan tingkat portofolio merujuk rencana suite (D1-D12); keputusan tingkat produk merujuk ADR HRIS.

| **ID**     | **Keputusan**                                                                                                                                                                            | **Sumber**                                              |
|------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------|
| D-HRS-01   | Komposisi, bukan konsolidasi: HRIS tetap deployable dan basis data terpisah; platform (`stery-suite`) memiliki organisasi, identitas, entitlement, penagihan, hub notifikasi, penyimpan audit, konsol | Suite D1; ADR-0001 diganti (lihat D-HRS-12)             |
| D-HRS-02   | Dua mode per produk lewat `STERY_MODE=standalone\|suite` (bawaan `standalone`); mode suite mendelegasikan identitas, entitlement, penagihan, notifikasi, audit lewat enam port Go dengan dua adaptor tiap port | Suite D2                                  |
| D-HRS-03   | Sesi suite = cookie buram `stery_session` (domain induk `tbd`), divalidasi lewat `SessionService/Introspect` dengan cache lokal 30 detik; HRIS mempertahankan login kata sandi (bcrypt, `services/auth-service/internal/domain/user.go:244-250`) tetapi sesi menjadi opaque, JWT dihapus | Suite D3, D10 |
| D-HRS-04   | Entitlement lewat `EntitlementService/Check` dan snapshot bertanda tangan; saat platform mati, snapshot terakhir dipakai hingga 24 jam (`tbd`) lalu fitur berbayar gagal-tertutup, tidak pernah di mode standalone | Suite D4                              |
| D-HRS-05   | Peristiwa: tanpa NATS; peristiwa domain dalam proses + outbox transaksional Postgres + konsumen idempoten; lintas produk lewat outbox -> platform `IntegrationService/Publish` | Suite D6; ADR-0003 diganti                |
| D-HRS-06   | Data: satu database `hris` dan satu peran aplikasi `hris` per profil; tanpa FK lintas database; RLS tetap wajib (ENABLE+FORCE, `USING` dan `WITH CHECK`, peran NOBYPASSRLS) lewat `kit/tenantdb.WithTenant` | Suite D7, D8; ADR-0002 dipertahankan      |
| D-HRS-07   | Web: satu `apps/web` SvelteKit SPA (`ssr=false`, `adapter-static`)                                                                                                                         | Suite D9                                                |
| D-HRS-08   | Observabilitas baseline = `slog` JSON + `/healthz`; OpenTelemetry/Prometheus tidak termasuk baseline, keputusan pemilik `tbd` (Q-HRS-08)                                                    | Suite D11                                               |
| D-HRS-09   | Urutan kerja: HRIS diselaraskan pada tahap S3 setelah POS (S2); satu penyelarasan berjalan pada satu waktu; berakhir di gerbang stabilisasi HRIS                                           | Suite D12                                               |
| D-HRS-10   | Payroll, aturan pajak, rekrutmen, kinerja ditunda ke fase 2                                                                                                                               | ADR-0004 (`docs/adr/0004-defer-payroll-to-phase2.md`)   |
| D-HRS-11   | Stack, pola, desain sistem, dan arsitektur seluruh suite mengikuti SELA; HRIS di-spesifikasi ulang ke stack itu                                                                             | Keputusan pemilik (rencana suite)                       |
| D-HRS-12   | Tingkat gateway Laravel dihapus: Go menyajikan REST `/api/v1` sendiri (stdlib `net/http`) dan gRPC untuk klien mesin; keputusan ADR-0001 (grpc-gateway + Laravel) digantikan                | STACK-SPEC baris 4                                      |

### 9.4 Pertanyaan Terbuka

| **ID**    | **Pertanyaan**                                                                                                                              | **Pemilik keputusan**  |
|-----------|---------------------------------------------------------------------------------------------------------------------------------------------|------------------------|
| Q-HRS-01  | Segmen target dan ukuran perusahaan tipikal (jumlah karyawan) untuk HRS.                                                                    | Pemilik produk         |
| Q-HRS-02  | Tier, nilai `hrs.employees.max` per tier, harga, siklus tagih, trial, dan mekanisme penagihan mode standalone.                               | Pemilik produk         |
| Q-HRS-03  | Daftar kompetitor dan hasil riset (8.1).                                                                                                     | Bisnis/Pemasaran       |
| Q-HRS-04  | Dasar hukum dan masa retensi untuk data karyawan (kehadiran, cuti, dokumen, audit) dan pemetaan klasifikasi NIK/NPWP ke kategori UU PDP.    | Legal                  |
| Q-HRS-05  | Aturan jam kerja/shift, ambang terlambat, hari kerja per minggu, geofence, zona waktu per tenant.                                            | Pemilik produk         |
| Q-HRS-06  | Kebijakan cuti bawaan (jenis, jatah, carry-over, pro-rata) dan sumber kalender libur nasional serta cuti bersama.                            | Pemilik produk + Legal |
| Q-HRS-07  | Target ketersediaan/SLA, latensi p95, RPO/RTO.                                                                                               | Engineering + Bisnis   |
| Q-HRS-08  | Apakah OpenTelemetry/Prometheus menjadi modul opsional `kit` (D11)?                                                                          | Pemilik produk         |
| Q-HRS-09  | Penyedia LLM, lokasi pemrosesan, dan kebijakan pengiriman PII ke pihak ketiga untuk modul 08.                                                | Pemilik + Legal        |
| Q-HRS-10  | Jadwal tinjauan hukum PPh21/BPJS/THR dan sprint desain aturan pajak (prasyarat ADR-0004).                                                    | Pemilik + Legal        |
| Q-HRS-11  | Domain induk cookie `stery_session` (D3) dan masa toleransi entitlement saat platform mati (D4, bawaan 24 jam).                              | Pemilik produk         |
| Q-HRS-12  | Tempat dan cara hosting produksi (`deploy/k8s` kosong; belum ada target deploy yang berfungsi).                                              | Pemilik + Engineering  |
| Q-HRS-13  | Penyedia SMTP produksi dan kebutuhan kanal WhatsApp/push.                                                                                    | Pemilik produk         |
| Q-HRS-14  | Pembagian peran pengendali/prosesor data dan templat perjanjian pemrosesan data dengan tenant.                                               | Legal                  |
| Q-HRS-15  | Target numerik KPI (bagian 11).                                                                                                              | Pemilik produk         |

## 10. Risiko dan Mitigasi

| **ID**       | **Risiko**                                                                                                                                                                | **Dampak**                                            | **Mitigasi**                                                                                                                                              | **Bukti**                                                                                                  |
|--------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------|
| RISK-HRS-01  | Dokumen status menyatakan 100% dan production ready padahal alur ujung-ke-ujung belum terbukti                                                                            | Keputusan jadwal dan penjualan berdasar data keliru   | BR-HRS-22; FSD 11.3 mencatat kontradiksi; status dilaporkan hanya dengan bukti kode                                                                       | `PROJECT_PROGRESS.md`, `PHASE1_RELEASE.md`, `IMPLEMENTATION_COMPLETE.md`, `docs/PHASE5_SUMMARY.md`          |
| RISK-HRS-02  | Migrasi gagal dari basis data kosong (absensi 001 sintaks tidak valid, absensi 002 dan cuti 002 menduplikasi nama indeks)                                                  | Layanan tidak dapat dipasang; pelanggan tidak bisa onboarding | Gerbang: seluruh migrasi `up -> down -> up` di CI pada DB segar                                                                                     | `services/attendance-service/migrations/001_create_attendance_schema.up.sql`, `services/leave-service/migrations/002_add_performance_indexes.up.sql` |
| RISK-HRS-03  | Tidak ada autentikasi/otorisasi di layanan; tenant dan aktor dipercaya dari badan permintaan                                                                               | Kebocoran lintas tenant, pemalsuan persetujuan        | Interceptor deny-by-default, `RequirePermission`, tenant dari sesi bukan dari body (FR-HRS-SEC.4, FR-HRS-SEC.5)                                           | `services/_shared/server/grpc_interceptors.go`, `services/employee-service/internal/interfaces/grpc/employee_service.go` |
| RISK-HRS-04  | NIK, NPWP, tanggal lahir tersimpan teks polos; tidak ada kebijakan UU PDP                                                                                                 | Pelanggaran UU PDP, sanksi, kehilangan kepercayaan    | Enkripsi `kit/piicrypt`, klasifikasi data, retensi, hak subjek data (FSD 8; BR-HRS-12)                                                                    | `services/employee-service/migrations/001_create_employee_schema.up.sql`                                    |
| RISK-HRS-05  | Cakupan penulisan ulang besar saat penyelarasan stack (gateway Laravel, SSR, JWT, NATS, golang-migrate, sqlc, zap, MinIO, Redis)                                           | Jadwal molor, regresi RLS                             | Domain dan application dipakai ulang; hanya infrastruktur dan antarmuka ditulis ulang; uji lintas tenant wajib per modul (FSD 2.7)                          | `go.work`, `apps/api-gateway/composer.json`                                                                 |
| RISK-HRS-06  | Tim kecil: pekerjaan paralel (stabilisasi, penyelarasan, suite) melebihi kapasitas                                                                                         | Antrian panjang, kualitas turun                       | Satu penyelarasan pada satu waktu (D12); HRIS setelah POS                                                                                                  | Analisis Dampak Risiko, RISK-SUI-08 (dokumen suite)                                                         |
| RISK-HRS-07  | Logika payroll/pajak dirilis tergesa-gesa                                                                                                                                  | Pelanggaran regulasi bagi tenant                      | ADR-0004: larangan merge; tinjauan hukum wajib; rute 501 sampai siap                                                                                       | `docs/adr/0004-defer-payroll-to-phase2.md`                                                                  |
| RISK-HRS-08  | Tidak ada target deploy yang berfungsi: `deploy/k8s` kosong, CD merujuk Dockerfile yang tidak ada, hanya dua layanan punya Dockerfile                                       | Tidak bisa merilis ke staging/produksi                | Deploy mengikuti pola SELA (Caddy + compose) pada penyelarasan; keputusan hosting Q-HRS-12                                                                | `.github/workflows/cd-staging.yml`, `Dockerfile.base`                                                       |
| RISK-HRS-09  | Peristiwa domain bermuatan kosong dan konsumen tidak cocok dengan penerbit sehingga rantai onboarding dan notifikasi tidak bekerja                                          | Fitur lintas modul tampak jadi tetapi tidak jalan     | Outbox + kontrak muatan berversi + uji kontrak penerbit/konsumen (FSD 2.5, 7.2)                                                                            | `services/leave-service/internal/infrastructure/nats/event_publisher.go`, `services/employee-service/internal/infrastructure/nats_consumer.go` |
| RISK-HRS-10  | Mode suite menjadikan platform titik tunggal kegagalan untuk identitas dan entitlement                                                                                     | Seluruh login HRS terganggu                           | Cache Introspect 30 detik, snapshot entitlement 24 jam, standalone tidak terpengaruh (D-HRS-04)                                                            | Analisis Dampak Risiko, RISK-SUI-11 (dokumen suite)                                                         |
| RISK-HRS-11  | Harga, tier, dan model penagihan tidak terdefinisi                                                                                                                         | Tidak dapat dijual                                    | Q-HRS-02 diputuskan sebelum S4                                                                                                                            | Tidak ada bukti di repositori (nilai `tbd`)                                                                 |
| RISK-HRS-12  | Modul AI memproses PII ke LLM pihak ketiga tanpa kebijakan                                                                                                                 | Kebocoran data, pelanggaran UU PDP                    | Modul 08 tidak dikerjakan sebelum Q-HRS-09; batas biaya dan laju per tenant                                                                                | `proto/hris/ai/v1/ai.proto`                                                                                 |
| RISK-HRS-13  | Kredensial dev dan password bawaan tertulis di berkas yang dilacak git dan sebagai nilai bawaan di kode                                                                    | Penyalahgunaan bila terbawa ke lingkungan nyata       | Rotasi dan pembersihan riwayat; konfigurasi fail-fast tanpa default rahasia (FR-HRS-SEC.7)                                                                | `deploy/docker/docker-compose.yml`, `deploy/docker/nats.conf`, `deploy/docker/init-postgres.sql`, `docker-compose.production.yml` |
| RISK-HRS-14  | Notifikasi tidak pernah terkirim (adaptor email tidak terpasang) sehingga pengguna tidak mengetahui pengajuan cuti                                                          | Persetujuan cuti tertunda, kepercayaan turun          | FR-HRS-05.2 dikerjakan dan diuji; antrean + retry                                                                                                          | `services/notification-service/cmd/server/main.go`                                                          |

## 11. KPI

Definisi saja; seluruh target `tbd` (Q-HRS-15). Sumber data: basis data produk dan jejak audit.

| **KPI**                               | **Definisi**                                                                                              | **Target** |
|---------------------------------------|-----------------------------------------------------------------------------------------------------------|------------|
| Tenant aktif                          | Jumlah tenant dengan >= 1 karyawan aktif dan >= 1 aktivitas (check-in atau pengajuan cuti) dalam 30 hari   | `tbd`      |
| Waktu onboarding tenant               | Selang waktu dari registrasi tenant sampai karyawan pertama dibuat selain admin                           | `tbd`      |
| Karyawan terkelola per tenant         | Rata-rata karyawan aktif per tenant aktif                                                                  | `tbd`      |
| Waktu keputusan cuti                  | Median waktu dari pengajuan cuti sampai disetujui atau ditolak                                            | `tbd`      |
| Cakupan absensi                       | Persentase karyawan aktif yang memiliki catatan kehadiran pada hari kerja                                  | `tbd`      |
| Keberhasilan pengiriman notifikasi    | Persentase notifikasi berstatus terkirim terhadap yang dibuat                                              | `tbd`      |
| Insiden lintas tenant                 | Jumlah temuan akses lintas tenant per rilis (target wajar: nol)                                            | `tbd`      |
| Kelulusan gerbang stabilisasi         | Jumlah butir gerbang HRIS (bagian 12) yang lulus dibanding total                                           | `tbd`      |
| Konversi trial ke berbayar            | Persentase tenant trial yang menjadi pelanggan berbayar                                                    | `tbd`      |
| Churn tenant                          | Persentase tenant berhenti berlangganan per periode                                                        | `tbd`      |

## 12. Linimasa dan Fase

Durasi seluruhnya `tbd` (tim kecil; estimasi dibuat setelah pemilik menetapkan prioritas). Urutan mengikuti keputusan D12.

| **Fase**                                           | **Cakupan**                                                                                                                                              | **Kriteria keluar (gerbang)**                                                                                                                                                                                                             | **Durasi** |
|----------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------|
| H0 Dokumentasi (S0)                                | BRD/PRD/FSD HRIS; analisis dampak                                                                                                                        | Dokumen disetujui pemilik                                                                                                                                                                                                                 | `tbd`      |
| H1 Stabilisasi HRIS (S3, bagian pertama)           | Memperbaiki cacat penghalang pada stack saat ini dan menutup gerbang berikut                                                                              | (1) semua migrasi `up -> down -> up` pada DB segar di CI; (2) login nyata lewat API; (3) `UpdateEmployee` dan org chart diimplementasi; (4) presign dan rollback dokumen diimplementasi dengan uji; (5) NIK/NPWP/tanggal lahir terenkripsi dengan `kit/piicrypt`; (6) pembatasan laju ditegakkan; (7) rahasia dan binari dibersihkan dari repositori | `tbd`      |
| H2 Penyelarasan stack (S3, bagian kedua)           | Gateway Laravel -> adaptor HTTP Go `/api/v1`; SSR -> SPA; JWT -> sesi buram; NATS -> outbox; golang-migrate -> goose; sqlc -> pgx; zap -> slog; MinIO -> RustFS; Redis -> Dragonfly; layanan -> modul dalam satu proses | Uji lintas tenant lulus pada setiap modul; `scripts/check.sh` hijau pada pola SELA; cakupan >= 80% (target SELA)                                                                                                              | `tbd`      |
| H3 Mode suite (S4)                                 | Adaptor suite untuk enam port; entitlement `hrs.employees.max`; konsol platform                                                                           | Mode `standalone` tidak berubah; mode `suite` lulus uji SSO dan degradasi entitlement                                                                                                                                                      | `tbd`      |
| H4 Integrasi suite (S5)                            | INT-1 dan INT-2 jika permintaan tervalidasi                                                                                                               | Validasi permintaan pasar; kontrak peristiwa berversi                                                                                                                                                                                     | `tbd`      |
| H5 Fase 2 domain berisiko (di luar rencana ini)    | Payroll, pajak, rekrutmen, kinerja                                                                                                                        | Prasyarat ADR-0004 selesai (tinjauan hukum, ADR, desain saga)                                                                                                                                                                              | `tbd`      |

Catatan koreksi pada gerbang (7): berkas `server.exe` yang ada di pohon kerja tidak dilacak git (`.gitignore:7` mengabaikan `*.exe`), sehingga yang perlu dibersihkan adalah rahasia yang dilacak git dan artefak lokal; verifikasi gerbang tetap memeriksa tidak ada binari terlacak.

## 13. Status Implementasi Saat Ini

Status berdasar pembacaan statis kode (tanpa build atau uji dijalankan; lihat FSD 1.4). Cabang `development`, 33 komit, komit terakhir 10 Juni 2026. Cacat rinci per FR ada di FSD bagian 4 dan 11.

| **Modul**            | **Deskripsi**                                                                                                                                                | **Status**  | **Bukti**                                                                                                                                                           |
|----------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 01 AUTH              | 6 RPC berhandler (login, refresh, validasi, revoke, izin, registrasi tenant); JWT RS256; tidak ada penguncian akun, interceptor, atau pengelolaan pengguna       | Sebagian    | `services/auth-service/internal/interfaces/grpc/auth_service.go`, `services/auth-service/cmd/server/main.go`                                                        |
| 02 EMP               | 8 dari 10 RPC berhandler; `UpdateEmployee` dan `GetOrgChart` belum; ketidakcocokan SQL (nomor karyawan, kolom `description`, huruf enum) berpotensi menggagalkan penulisan | Sebagian    | `services/employee-service/internal/interfaces/grpc/employee_service.go`, `services/employee-service/internal/infrastructure/postgres/employee_repository.go`      |
| 03 ATT               | 5 RPC berhandler; migrasi 001 tidak valid; override gagal untuk catatan baru; tanpa aturan jam kerja                                                         | Sebagian    | `services/attendance-service/migrations/001_create_attendance_schema.up.sql`, `services/attendance-service/internal/application/commands/override_attendance.go`  |
| 04 LEV               | 8 RPC berhandler; `ApplyLeave` tidak mengirim `days_count`; tumpang tindih dan hari kerja tidak ditegakkan; tidak ada pembuat jenis cuti                      | Sebagian    | `services/leave-service/internal/interfaces/grpc/leave_service.go`, `services/leave-service/internal/domain/leave_request.go`                                       |
| 05 NTF               | 5 RPC berhandler; notifikasi hanya disimpan (PENDING); adaptor email tidak dipasang                                                                          | Sebagian    | `services/notification-service/cmd/server/main.go`, `services/notification-service/internal/infrastructure/channels/email_adapter.go`                               |
| 06 AUD               | 3 RPC berhandler (`GetEntry` salah); empat konsumer; tanpa RLS dan tanpa trigger imutabilitas                                                                | Sebagian    | `services/audit-service/internal/interfaces/grpc/audit_service.go`, `services/audit-service/migrations/001_create_audit_schema.up.sql`                              |
| 07 DOC               | 4 dari 7 RPC berisi; `DownloadDocument`, `GetPresignedUrl`, dan `RollbackDocumentVersion` masih placeholder yang mengembalikan respons kosong; 0 berkas uji; tanpa pemindaian malware | Sebagian    | `services/document-service/internal/interfaces/grpc/document_service.go`, `services/document-service/internal/interfaces/grpc/mapping.go`                          |
| 08 AI                | Hanya kontrak proto (3 RPC); tidak ada kode layanan                                                                                                          | Belum       | `proto/hris/ai/v1/ai.proto`, `go.work`                                                                                                                              |
| 09 PAY               | Ditunda (ADR-0004)                                                                                                                                           | Ditunda     | `docs/adr/0004-defer-payroll-to-phase2.md`                                                                                                                          |
| 10 TAX               | Ditunda (ADR-0004)                                                                                                                                           | Ditunda     | `docs/adr/0004-defer-payroll-to-phase2.md`                                                                                                                          |
| 11 REC               | Ditunda (ADR-0004)                                                                                                                                           | Ditunda     | `docs/adr/0004-defer-payroll-to-phase2.md`                                                                                                                          |
| 12 PRF               | Ditunda (ADR-0004)                                                                                                                                           | Ditunda     | `docs/adr/0004-defer-payroll-to-phase2.md`                                                                                                                          |

Komponen pendukung:

| **Komponen**                    | **Deskripsi**                                                                                                                                         | **Status**  | **Bukti**                                                                                              |
|---------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------|-------------|--------------------------------------------------------------------------------------------------------|
| Gateway Laravel 11              | Hanya rute `auth/*` dan cuti; rute karyawan, absensi, dokumen, audit hanya komentar; memanggil jalur HTTP yang tidak dilayani layanan Go              | Sebagian    | `apps/api-gateway/routes/api.php`, `apps/api-gateway/app/Services/BaseGrpcGatewayClient.php`            |
| Web SvelteKit (SSR)             | Empat halaman (login, dashboard, ajukan cuti, admin cuti); login dimock; tanpa i18n                                                                    | Sebagian    | `apps/web/src/routes/(auth)/login/+page.server.ts`, `apps/web/src/hooks.server.ts`                      |
| CI/CD                           | CI Go: lint 7 layanan, uji unit dan integrasi hanya auth, employee, leave; CD merujuk Dockerfile yang tidak ada dan `deploy/k8s` yang kosong                          | Sebagian    | `.github/workflows/ci-go.yml`, `.github/workflows/cd-staging.yml`                                       |
| Berkas uji                      | 22 berkas `_test.go` di 6 layanan (document-service: 0); uji integrasi bertag build `integration`                                                                      | Sebagian    | `services/leave-service/internal/domain/leave_request_test.go`, `services/employee-service/internal/integration/employee_integration_test.go` |
| Infrastruktur lokal             | Compose untuk Postgres 16, Redis 7, NATS 2.10, MinIO, Jaeger, Prometheus, PgBouncer                                                                    | Selesai     | `deploy/docker/docker-compose.yml`                                                                      |

Ringkasan: 7 dari 8 direktori layanan berisi kode (`ai-service` kosong); 39 dari 47 RPC proto memiliki handler berisi (delapan lainnya: `UpdateEmployee`, `GetOrgChart`, tiga RPC dokumen, tiga RPC AI). "Berhandler" tidak berarti "berfungsi": lihat FSD 4 dan 11.

## 14. Persetujuan

Dokumen ini perlu ditinjau dan disetujui sebelum PRD dan FSD HRIS-Stery dijadikan acuan pengerjaan, termasuk butir di luar lingkup (4.4) yang masih berupa usulan.

| **Nama** | **Jabatan**      | **Tanda Tangan** | **Tanggal** |
|----------|------------------|------------------|-------------|
|          | Product Owner    |                  |             |
|          | Business Sponsor |                  |             |
|          | Engineering Lead |                  |             |
|          | Legal/Compliance |                  |             |
