# PRD HRIS-Stery

*Product Requirements Document (PRD): HRIS multi-tenant untuk perusahaan di Indonesia: data karyawan, absensi, dan cuti yang rapi, aman, dan dapat diaudit*

|                         |                                                                                                                                                                                                              |
|-------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Nama Dokumen**        | PRD - HRIS-Stery (SKU `HRS`)                                                                                                                                                                                 |
| **Versi**               | 1.0 (Draft)                                                                                                                                                                                                  |
| **Tanggal**             | 6 November 2026                                                                                                                                                                                              |
| **Disusun oleh**        | Tim Produk & Engineering                                                                                                                                                                                     |
| **Status**              | Draft untuk ditinjau                                                                                                                                                                                         |
| **Dokumen Terkait**     | BRD HRIS-Stery (`BRD_HRIS.id.md`), FSD HRIS-Stery (`FSD_HRIS.id.md`), `00_Analisis_Dampak_Risiko`, BRD/PRD/FSD Stery Suite (folder `stery-suite/docs`), `SELA/CLAUDE.md` (standar stack), ADR-0001 s.d. ADR-0004 |
| **Perubahan dari vX**   | Dokumen pertama                                                                                                                                                                                              |

## 1. Visi dan Ringkasan

HRIS-Stery membuat urusan harian HR perusahaan di Indonesia rapi dan dapat dipertanggungjawabkan: satu sumber kebenaran data karyawan, absensi yang tercatat dan dapat dikoreksi dengan jejak, serta cuti dengan saldo dan persetujuan yang jelas. Setiap perusahaan (tenant) hanya melihat datanya sendiri; data pribadi karyawan dilindungi sesuai UU PDP No. 27/2022.

Produk dijual sebagai SKU `HRS`: berdiri sendiri (mode `standalone`, bawaan) atau terhubung ke Stery Suite (mode `suite`) bersama POS dan SELA. Payroll, pajak, rekrutmen, dan kinerja sengaja ditunda (ADR-0004) sampai ada tinjauan hukum; ini harus disampaikan jelas dalam positioning karena ADR-0004 mencatat konsekuensinya: tenant belum dapat menjalankan payroll lewat sistem.

Prinsip produk:

- **Data benar dulu**: tidak ada fitur yang diklaim selesai tanpa bukti kode (BR-HRS-22).
- **Aman secara bawaan**: isolasi tenant di basis data, otorisasi di backend, PII terenkripsi (BR-HRS-02, BR-HRS-03, BR-HRS-12).
- **Berdiri sendiri maupun suite**: perilaku standalone tidak berubah oleh mode suite (BR-HRS-15).
- **Tidak mengarang aturan**: nilai bisnis yang belum diputuskan (jam kerja, jatah cuti, harga) adalah `tbd` dengan pertanyaan terbuka.

## 2. Masalah dan Target Pengguna

### 2.1 Masalah Inti

Hipotesis (belum divalidasi riset pengguna `[INFERENCE]`; Q-HRS-01, Q-HRS-03):

- Data karyawan tersebar di lembar kerja dan pesan; tidak ada struktur organisasi yang pasti.
- Cuti diajukan lewat chat; saldo dihitung manual; keputusan sulit dilacak.
- Kehadiran dicatat tidak seragam; koreksi tidak meninggalkan jejak.
- Dokumen karyawan dan identitas (NIK, NPWP) disimpan tanpa kontrol akses dan tanpa kebijakan UU PDP.
- Perusahaan yang memakai produk Stery lain (POS, SELA) harus masuk ke tiap produk dengan akun terpisah (konteks suite).

### 2.2 Persona Pengguna

| **Persona**                        | **Deskripsi**                                                                                                   | **Kebutuhan Utama**                                                                                                                | **Sensitivitas Harga** |
|------------------------------------|-----------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------|------------------------|
| HR Admin (`hr_admin`)              | Staf HR yang mendaftarkan perusahaan, mengelola karyawan, cuti, koreksi absensi, dokumen; juga pembeli/penentu   | Onboarding cepat, impor karyawan, aturan cuti yang benar, koreksi dengan jejak, laporan, kepatuhan UU PDP                        | `tbd` (Q-HRS-02)       |
| Manajer (`manager`)                | Atasan langsung; menyetujui cuti dan melihat kehadiran timnya                                                    | Antrean persetujuan yang jelas, notifikasi, melihat siapa cuti, bagan tim                                                        | Tidak membayar langsung |
| Karyawan (`employee`)              | Pengguna akhir; check-in/out, mengajukan cuti, melihat saldo                                                     | Langkah sedikit, status pengajuan jelas, bahasa Indonesia, dapat dipakai di ponsel                                                | Tidak membayar langsung |
| Admin Platform (operator Stery)    | Staf Stery yang mengelola tenant, entitlement, dukungan, dan melihat kesehatan layanan (mode suite via konsol)   | Melihat tenant dan penggunaan, mengatur batas `hrs.employees.max`, tidak mengakses data karyawan tanpa jejak audit                  | Tidak relevan          |

## 3. User Stories

| **ID**    | **Sebagai...**              | **Saya ingin...**                                                          | **Agar...**                                                                              | **Prioritas** |
|-----------|-----------------------------|----------------------------------------------------------------------------|------------------------------------------------------------------------------------------|---------------|
| US-HRS-01 | HR Admin                    | mendaftarkan perusahaan dan akun admin pertama sendiri                     | saya langsung memakai produk tanpa menunggu bantuan teknis                               | Must          |
| US-HRS-02 | HR Admin                    | masuk dengan email dan kata sandi dengan sesi yang aman                    | data perusahaan tidak dapat diakses orang lain                                           | Must          |
| US-HRS-03 | HR Admin                    | menambah pengguna, menetapkan peran, dan menonaktifkannya                  | hak akses sesuai tugas dan akun karyawan keluar tidak bisa dipakai                       | Must          |
| US-HRS-04 | HR Admin                    | membuat departemen berhierarki dan jabatan                                 | struktur organisasi perusahaan tergambar                                                 | Must          |
| US-HRS-05 | HR Admin                    | menambah karyawan satu per satu dengan nomor karyawan otomatis             | data induk karyawan lengkap dan unik                                                     | Must          |
| US-HRS-06 | HR Admin                    | mengimpor banyak karyawan dari berkas CSV                                  | onboarding perusahaan besar tidak memakan waktu berhari-hari                             | Should        |
| US-HRS-07 | HR Admin                    | mengubah data karyawan dan menerminasi karyawan yang keluar                | data selalu mutakhir dan akses karyawan keluar otomatis ditutup                          | Must          |
| US-HRS-08 | HR Admin / Manajer          | melihat bagan organisasi                                                   | saya tahu garis pelaporan dan siapa atasan siapa                                         | Must          |
| US-HRS-09 | Karyawan                    | check-in dan check-out dengan satu ketukan                                 | kehadiran saya tercatat tanpa kertas                                                     | Must          |
| US-HRS-10 | HR Admin                    | mengoreksi catatan kehadiran dengan alasan                                 | kesalahan diperbaiki dan perubahan dapat diaudit                                         | Must          |
| US-HRS-11 | Manajer                     | melihat kehadiran dan ringkasan tim saya                                   | saya tahu siapa hadir, terlambat, atau cuti                                              | Should        |
| US-HRS-12 | HR Admin                    | mengatur jam kerja, ambang terlambat, dan hari kerja tenant                | status kehadiran dihitung sesuai kebijakan perusahaan (nilai `tbd`, Q-HRS-05)            | Should        |
| US-HRS-13 | Karyawan                    | mengajukan cuti dengan jenis, tanggal, dan alasan                          | pengajuan saya tercatat dan diproses                                                     | Must          |
| US-HRS-14 | Karyawan                    | melihat saldo cuti per jenis dan riwayat pengajuan                         | saya tahu sisa cuti dan status pengajuan                                                 | Must          |
| US-HRS-15 | Manajer                     | menyetujui atau menolak cuti anggota tim dengan alasan                     | keputusan cepat dan tercatat; saldo diperbarui otomatis                                  | Must          |
| US-HRS-16 | Karyawan                    | membatalkan pengajuan cuti                                                 | saldo kembali bila rencana berubah                                                       | Must          |
| US-HRS-17 | HR Admin                    | mengatur jenis cuti dan jatah per tahun                                    | kebijakan cuti perusahaan terwakili (nilai bawaan `tbd`, Q-HRS-06)                       | Must          |
| US-HRS-18 | Manajer / Karyawan          | menerima notifikasi email dan di aplikasi untuk pengajuan dan keputusan    | saya tidak perlu memeriksa terus-menerus                                                 | Should        |
| US-HRS-19 | HR Admin                    | menelusuri jejak audit (siapa, apa, kapan)                                 | saya dapat membuktikan riwayat tindakan saat ada sengketa atau pemeriksaan               | Must          |
| US-HRS-20 | HR Admin                    | mengunggah dan mengunduh dokumen karyawan dengan aman                      | dokumen sensitif tidak tersebar dan hanya dapat dibuka pihak berwenang                   | Should        |
| US-HRS-21 | Karyawan                    | meminta akses atau koreksi atas data pribadi saya                          | hak saya sebagai subjek data UU PDP terpenuhi                                            | Must          |
| US-HRS-22 | Admin Platform              | melihat tenant dan mengatur batas `hrs.employees.max`                      | penjualan per tier dapat ditegakkan                                                      | Should        |
| US-HRS-23 | Pemilik Organisasi (suite)  | membuka HRIS dari pengalih produk tanpa login ulang                        | saya memakai beberapa produk Stery dengan satu akun                                      | Should        |
| US-HRS-24 | HR Admin                    | diberi peringatan mendekati batas karyawan paket                           | saya dapat meng-upgrade sebelum pembuatan karyawan ditolak                               | Could         |
| US-HRS-25 | HR Admin                    | meminta analisis resume dengan bantuan AI                                  | penyaringan kandidat lebih cepat (setelah kebijakan PII LLM diputuskan, Q-HRS-09)        | Could         |
| US-HRS-26 | HR Admin                    | menjalankan payroll (PPh21, BPJS, THR) di sistem                           | gaji dan pajak terhitung otomatis                                                        | Won't         |

## 4. Fitur Produk (Prioritas MoSCoW)

Fitur dirujuk ke requirement fungsional FSD (FR). Status implementasi per FR ada di FSD 11.1.

### 4.1 Must Have

- Registrasi tenant mandiri dengan admin pertama dan peran bawaan; login kata sandi dengan sesi buram aman; pengelolaan pengguna dan peran (FR-HRS-01.1 s.d. FR-HRS-01.7).
- Data induk: karyawan (buat, lihat, ubah, terminasi), departemen berhierarki, jabatan, bagan organisasi, PII terenkripsi (FR-HRS-02.1 s.d. FR-HRS-02.8).
- Absensi: check-in, check-out, koreksi HR dengan alasan, daftar kehadiran (FR-HRS-03.1 s.d. FR-HRS-03.4).
- Cuti: jenis cuti, saldo, pengajuan dengan hari kerja otomatis dan tanpa tumpang tindih, persetujuan, penolakan, pembatalan, daftar dan saldo (FR-HRS-04.1 s.d. FR-HRS-04.7).
- Notifikasi tersimpan multi-kanal dan pengiriman email nyata (FR-HRS-05.1, FR-HRS-05.2).
- Jejak audit append-only yang tidak dapat diubah, kueri berfilter, dan isolasi tenant (FR-HRS-06.1, FR-HRS-06.2, FR-HRS-06.4, FR-HRS-06.6, FR-HRS-06.7).
- Dokumen: unggah lewat URL bertanda tangan, pemindaian malware fail-closed, unduh berumur pendek, batas ukuran dan tipe (FR-HRS-07.1 s.d. FR-HRS-07.4, FR-HRS-07.8).
- Keamanan lintas modul: RLS penuh, uji lintas tenant, PII terenkripsi, sesi buram, deny-by-default, pembatasan laju, rahasia, hak subjek data UU PDP (FR-HRS-SEC.1 s.d. FR-HRS-SEC.8, FR-HRS-SEC.10).
- Gerbang stabilisasi HRIS (bagian 11) dan penyelarasan stack (BR-HRS-17, BR-HRS-18).
- Mode `standalone` utuh dan mode `suite` lewat adaptor (FR-HRS-01.10; bagian 10).

### 4.2 Should Have

- Impor karyawan massal CSV (FR-HRS-02.9); shell karyawan otomatis dari registrasi (FR-HRS-02.10); validasi batas `hrs.employees.max` (FR-HRS-02.11).
- Ringkasan kehadiran, aturan jam kerja dan terlambat, zona waktu tenant (FR-HRS-03.5, FR-HRS-03.6, FR-HRS-03.8).
- Kalender libur nasional dan cuti bersama (FR-HRS-04.8).
- Konsumsi peristiwa untuk notifikasi, daftar dan tandai dibaca (FR-HRS-05.3, FR-HRS-05.4).
- Ambil satu entri audit, retensi dan partisi audit (FR-HRS-06.3, FR-HRS-06.5).
- Versi dokumen, rollback, hapus dan retensi, klasifikasi (FR-HRS-07.5 s.d. FR-HRS-07.7).
- Verifikasi email dan pemulihan kata sandi (FR-HRS-01.8); pencatatan IP klien (FR-HRS-01.9).
- Header keamanan, CORS, CSRF; retensi dan penghapusan data (FR-HRS-SEC.9, FR-HRS-SEC.11).
- Antarmuka dwibahasa (BR-HRS-19).

### 4.3 Could Have

- Integrasi absensi-cuti (hari cuti menandai `ON_LEAVE`) (FR-HRS-03.7); penyesuaian saldo manual, carry-over, pro-rata (FR-HRS-04.9).
- Konfigurasi kanal per tenant dan preferensi opt-out notifikasi (FR-HRS-05.5, FR-HRS-05.7).
- Asisten AI: analisis resume, prediksi attrisi, chat HR (FR-HRS-08.1 s.d. FR-HRS-08.4).
- Hipotesis integrasi suite INT-1 dan INT-2 (BR-HRS-21; "hipotesis, validasi permintaan").
- Peringatan mendekati batas paket (US-HRS-24).

### 4.4 Won't Have (fase ini)

- Payroll, aturan pajak, PPh21, BPJS, THR, pesangon (modul 09, 10; ADR-0004).
- Rekrutmen dan kinerja (modul 11, 12; ADR-0004).
- Notifikasi WhatsApp dan push (FR-HRS-05.6; ditunda).
- Aplikasi seluler native (usulan, BRD 4.4).

## 5. Alur Pengguna (User Flow)

### 5.1 Alur Utama: Onboarding Tenant sampai Absensi

1. **Registrasi tenant**: HR Admin mengisi nama perusahaan, slug (3-50 karakter `[a-z0-9-]`), email, dan kata sandi (8-128 karakter, huruf besar, huruf kecil, angka, simbol). Sistem membuat tenant, tiga peran sistem (`hr_admin`, `manager`, `employee`), izin bawaan, dan admin pertama.
2. **Login pertama**: admin masuk; sesi buram disimpan di cookie HttpOnly; layar awal menampilkan langkah onboarding.
3. **Struktur**: admin membuat departemen dan jabatan (FR-HRS-02.5, FR-HRS-02.6).
4. **Karyawan**: admin menambah karyawan satu per satu atau mengimpor CSV (FR-HRS-02.1, FR-HRS-02.9). Setiap karyawan baru memicu pembuatan saldo cuti tahunan dan notifikasi sambutan.
5. **Akun**: admin mengundang karyawan menjadi pengguna dengan peran (FR-HRS-01.7).
6. **Absensi**: karyawan check-in dan check-out setiap hari kerja (FR-HRS-03.1, FR-HRS-03.2); HR mengoreksi bila perlu (FR-HRS-03.3).

### 5.2 Alur Persetujuan Cuti

1. Karyawan memilih jenis cuti, tanggal mulai dan selesai, dan alasan; sistem menghitung hari kerja, memeriksa saldo dan tumpang tindih.
2. Pengajuan berstatus `PENDING`; saldo tertunda (`pending_days`) bertambah; atasan menerima notifikasi.
3. Atasan atau HR membuka antrean, menyetujui atau menolak (alasan wajib saat menolak); pemohon tidak dapat menyetujui pengajuannya sendiri.
4. Saat disetujui, saldo tertunda berpindah menjadi terpakai; saat ditolak, saldo tertunda dipulihkan; pemohon menerima notifikasi.
5. Pemohon atau HR dapat membatalkan; pembatalan memulihkan saldo sesuai status sebelumnya.

### 5.3 Alur Manajer

Dasbor menampilkan antrean persetujuan cuti tim, kehadiran tim hari ini, dan bagan organisasi (FR-HRS-02.7, FR-HRS-03.4, FR-HRS-04.6).

### 5.4 Alur Admin Platform dan Organisasi (mode suite)

1. Pemilik organisasi memilih paket di platform; platform mengaktifkan entitlement `hrs.*`.
2. Pengguna membuka HRIS dari pengalih produk; sesi `stery_session` dipakai tanpa login ulang.
3. Admin platform melihat tenant dan mengatur `hrs.employees.max`; saat batas tercapai atau entitlement berakhir, hanya pembuatan karyawan baru yang ditolak dan data tetap dapat dibaca (BR-HRS-16).
4. Migrasi pelanggan standalone ke suite: tautkan organisasi suite dengan tenant HRS yang ada (FSD Stery Suite 7.5).

## 6. Kebutuhan Non-Fungsional

Seluruh angka target numerik `tbd` (Q-HRS-07, Q-HRS-15) sampai pemilik memutuskan. Realisasi teknis ada di FSD 9.

| **ID**       | **Kategori**                  | **Kebutuhan**                                                                                                                                             | **Target** |
|--------------|-------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|------------|
| NFR-HRS-01   | Isolasi tenant                | Setiap tabel ber-`tenant_id` memakai RLS ENABLE+FORCE dengan `USING` dan `WITH CHECK`; uji lintas tenant wajib di setiap modul                              | 100% tabel bertenant |
| NFR-HRS-02   | Keamanan sesi                 | Sesi buram di cookie HttpOnly Secure; hanya hash disimpan di sisi server; tanpa JWT dan tanpa token di penyimpanan peramban                                 | `tbd` (TTL sesi)     |
| NFR-HRS-03   | Otorisasi                     | Deny-by-default: setiap endpoint punya izin eksplisit yang diperiksa di backend; tenant dan aktor berasal dari sesi, bukan dari badan permintaan             | 100% endpoint        |
| NFR-HRS-04   | Pelindungan PII               | NIK, NPWP, tanggal lahir dienkripsi AES-256-GCM dengan indeks buta untuk pencarian; kunci terpisah per lingkungan                                          | 100% kolom PII kelas 2 |
| NFR-HRS-05   | Auditabilitas                 | Jejak audit append-only dengan penolakan UPDATE/DELETE/TRUNCATE di tingkat basis data                                                                       | 100% tindakan sensitif |
| NFR-HRS-06   | Ketersediaan                  | Target ketersediaan layanan                                                                                                                                 | `tbd`                |
| NFR-HRS-07   | Kinerja                       | Latensi p95 operasi baca/tulis utama (login, daftar karyawan, check-in, pengajuan cuti)                                                                     | `tbd`                |
| NFR-HRS-08   | Pembatasan laju               | Pembatasan laju login dan endpoint sensitif per alamat dan per akun; penguncian akun setelah percobaan gagal berulang                                       | `tbd` (ambang)       |
| NFR-HRS-09   | Migrasi                       | Semua migrasi reversibel; CI menjalankan `up -> down -> up` pada DB segar                                                                                  | 100% migrasi         |
| NFR-HRS-10   | Kualitas kode                 | Cakupan uji Go dan web >= 80% (lantai SELA); `-race` di CI; satu skrip `scripts/check.sh` yang sama dengan SELA                                              | >= 80%               |
| NFR-HRS-11   | Observabilitas                | Log `slog` JSON tanpa PII dan rahasia; `/healthz`; OpenTelemetry/Prometheus keputusan pemilik `tbd` (Q-HRS-08)                                               | baseline D11         |
| NFR-HRS-12   | Lokalisasi                    | Antarmuka `{id,en}` dengan uji paritas katalog; WIB; IDR                                                                                                    | 100% string          |
| NFR-HRS-13   | Konfigurasi dan rahasia       | Konfigurasi dari variabel lingkungan dengan validasi gagal-cepat; tidak ada rahasia bawaan di kode atau berkas yang dilacak git                              | 0 rahasia di repo    |
| NFR-HRS-14   | Cadangan dan pemulihan        | Cadangan basis data terjadwal dan uji pemulihan                                                                                                             | RPO/RTO `tbd`        |
| NFR-HRS-15   | Retensi dan penghapusan       | Retensi tiap kelas data dan penghapusan/anonimisasi saat tenggat (UU PDP)                                                                                   | `tbd` (Q-HRS-04)     |
| NFR-HRS-16   | Portabilitas mode             | `standalone` dan `suite` dipilih lewat `STERY_MODE` tanpa mengubah perilaku standalone                                                                      | 0 regresi standalone |
| NFR-HRS-17   | Aksesibilitas                 | Antarmuka dapat dipakai dengan papan ketik dan pembaca layar pada alur utama                                                                                | `tbd` (tingkat WCAG) |
| NFR-HRS-18   | Keandalan peristiwa           | Peristiwa at-least-once dengan outbox transaksional dan konsumen idempoten (`processed_events`)                                                             | 0 peristiwa hilang   |

## 7. Model Monetisasi (Ringkasan Produk)

Mengacu pada BRD HRIS 7: satu dimensi batas `hrs.employees.max`, tier dan harga `tbd` (Q-HRS-02). Penagihan mode suite oleh platform (D5); penagihan standalone `tbd`. Payroll dan modul ditunda berpotensi add-on.

## 8. Positioning

`tbd` (menunggu segmen dan riset kompetitor, Q-HRS-01, Q-HRS-03). Hipotesis pembeda yang perlu diuji: (1) data karyawan dilindungi sesuai UU PDP sejak desain; (2) berdiri sendiri maupun dalam bundel dengan POS dan SELA satu akun; (3) jejak audit yang tidak dapat diubah; (4) Bahasa Indonesia dan WIB bawaan. Pernyataan positioning tidak ditulis sampai ada data.

## 9. Metrik Keberhasilan

Definisi KPI ada di BRD HRIS 11; seluruh target `tbd` (Q-HRS-15). Metrik produk yang diukur: tenant aktif, waktu onboarding tenant, waktu keputusan cuti, cakupan absensi, keberhasilan pengiriman notifikasi, insiden lintas tenant, dan kelulusan gerbang stabilisasi.

## 10. Mode Operasi

| **Aspek**                 | **`standalone` (bawaan)**                                              | **`suite`**                                                                                                 |
|---------------------------|------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------|
| Identitas dan sesi        | Pengguna dan kata sandi HRIS sendiri (bcrypt); sesi buram cookie HRIS  | Platform menerbitkan sesi `stery_session`; HRIS memvalidasi lewat `SessionService/Introspect` (cache 30 detik) |
| Tenant                    | `tenants` milik HRIS                                                   | Organisasi platform ditautkan ke tenant HRS (`external_ref`)                                                |
| Entitlement               | Tidak ada pembatasan komersial lokal (`tbd`; Q-HRS-02)                 | `EntitlementService/Check` dan snapshot bertanda tangan; gagal-tertutup untuk fitur berbayar setelah 24 jam |
| Penagihan                 | Tidak ada (`tbd`)                                                      | Platform (Midtrans QRIS melalui `payment.Provider`)                                                         |
| Notifikasi                | Layanan notifikasi HRIS sendiri                                        | Hub notifikasi platform lewat port `Notification`                                                           |
| Audit                     | Jejak audit HRIS sendiri                                               | Tetap ditulis lokal dan diteruskan ke penyimpan audit platform lewat port `Audit`                           |
| Peristiwa lintas produk   | Tidak ada                                                              | Outbox -> `IntegrationService/Publish`; opt-in per organisasi lewat fitur `integration.hrs_pos`, `integration.hrs_sel` (usulan) |
| Web                       | SPA HRIS di port 5173 (profil standalone)                              | SPA yang sama; pengalih produk dan konsol di `stery-suite/apps/web`                                         |
| Port bawaan               | API 8080, gRPC 9090, web 5173                                          | API 8280, gRPC 9290, web 5373                                                                               |

Kontrak Produk (yang HRIS penuhi agar menjadi produk suite): manifes (`product_code` `hrs`, kunci fitur modul, dimensi `hrs.employees.max`), enam port (`Identity`, `Tenant`, `Entitlement`, `Notification`, `Audit`, `Events`) dengan dua adaptor tiap port, `STERY_MODE`, `/healthz`, outbox publishing, daftar periksa kesesuaian STACK-SPEC, dan triple BRD/PRD/FSD (dokumen ini adalah bagian triple itu). Rincian di FSD Stery Suite 4.08.

## 11. Roadmap Rilis (Indikatif)

Tanggal `tbd`. Urutan mengikuti D12.

| **Rilis** | **Fokus**                          | **Fitur utama / kriteria keluar**                                                                                                                                                                                                                                                   |
|-----------|------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| R0        | Dokumen (S0)                       | BRD/PRD/FSD HRIS disetujui                                                                                                                                                                                                                                                         |
| R1        | Stabilisasi HRIS (S3)              | Gerbang: (1) semua migrasi `up -> down -> up` pada DB segar di CI; (2) login nyata lewat API; (3) `UpdateEmployee` dan org chart diimplementasi; (4) presign dan rollback dokumen diimplementasi dengan uji; (5) NIK/NPWP/tanggal lahir terenkripsi dengan `kit/piicrypt`; (6) pembatasan laju ditegakkan; (7) rahasia dan binari dibersihkan |
| R2        | Penyelarasan stack (S3)            | Adaptor HTTP Go `/api/v1` menggantikan gateway Laravel; SPA; sesi buram; outbox; goose; pgx; slog; RustFS; Dragonfly; modul dalam satu proses; uji lintas tenant per modul                                                                                                        |
| R3        | Mode suite (S4)                    | Adaptor suite enam port; entitlement `hrs.employees.max`; SSO; degradasi entitlement                                                                                                                                                                                               |
| R4        | Integrasi suite (S5)               | INT-1 dan INT-2 jika permintaan tervalidasi                                                                                                                                                                                                                                        |
| R5        | Fase 2 (di luar rencana ini)       | Payroll, pajak, rekrutmen, kinerja setelah prasyarat ADR-0004                                                                                                                                                                                                                       |

## 12. Di Luar Cakupan

- Payroll, pajak, BPJS, THR, pesangon (ADR-0004).
- Rekrutmen dan kinerja (ADR-0004).
- WhatsApp dan push.
- Aplikasi seluler native; integrasi ke sistem pemerintah.
- Penagihan standalone otomatis di dalam HRIS.

## 13. Ketergantungan dan Risiko

| **Ketergantungan**                                           | **Dampak**                                                            | **Risiko terkait** |
|--------------------------------------------------------------|-----------------------------------------------------------------------|--------------------|
| `kit` suite (tenantdb, session, outbox, piicrypt, audit, httpx, config, db, testdb, objstore, health) | Penyelarasan tidak dapat dimulai sebelum S1 selesai | RISK-HRS-05, RISK-HRS-06 |
| Platform `stery-suite` (IAM, ENT, BIL)                       | Mode suite hanya dapat diuji setelah platform tersedia                 | RISK-HRS-10        |
| Tinjauan hukum (UU PDP, PPh21/BPJS/THR)                      | Menghalangi modul PII penuh dan fase payroll                           | RISK-HRS-04, RISK-HRS-07 |
| Keputusan pemilik: harga, segmen, hosting, SMTP, LLM         | Menghalangi penjualan dan modul 05 dan 08                              | RISK-HRS-11, RISK-HRS-08, RISK-HRS-12 |
| Penyedia SMTP produksi                                        | Pengiriman notifikasi email                                            | RISK-HRS-14        |
| Gerbang stabilisasi dan perbaikan migrasi                    | Seluruh alur ujung-ke-ujung                                            | RISK-HRS-02, RISK-HRS-09 |

Daftar risiko lengkap: BRD HRIS 10.

## 14. Status Saat Ini

Berdasar pembacaan statis kode per 6 November 2026 (tanpa menjalankan build/uji). Status per requirement: FSD 11.1; kontradiksi dokumen status vs kode: FSD 11.3.

| **Area**                         | **Status** | **Ringkasan**                                                                                                                                  | **Bukti**                                                                                                                      |
|----------------------------------|------------|------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------|
| 01 AUTH                          | Sebagian   | Registrasi, login, refresh, revoke, izin ada; tanpa penguncian akun, tanpa pengelolaan pengguna, tanpa interceptor autentikasi                  | `services/auth-service/internal/interfaces/grpc/auth_service.go`                                                                |
| 02 EMP                           | Sebagian   | Buat/lihat/daftar/terminasi ada; ubah dan org chart belum; risiko kegagalan penulisan SQL                                                       | `services/employee-service/internal/interfaces/grpc/employee_service.go`                                                        |
| 03 ATT                           | Sebagian   | Check-in/out ada; migrasi tidak valid; tanpa aturan jam kerja                                                                                   | `services/attendance-service/migrations/001_create_attendance_schema.up.sql`                                                    |
| 04 LEV                           | Sebagian   | State machine dan saldo ada; pengajuan tidak mengirim `days_count`; overlap tak ditegakkan                                                      | `services/leave-service/internal/interfaces/grpc/leave_service.go`                                                              |
| 05 NTF                           | Sebagian   | Tersimpan PENDING; tidak pernah terkirim                                                                                                        | `services/notification-service/cmd/server/main.go`                                                                              |
| 06 AUD                           | Sebagian   | Konsumer dan kueri ada; RLS mati; tanpa trigger imutabilitas                                                                                    | `services/audit-service/migrations/001_create_audit_schema.up.sql`                                                              |
| 07 DOC                           | Sebagian   | Unggah metadata dan presigned PUT; unduh/presign/rollback placeholder                                                                           | `services/document-service/internal/interfaces/grpc/document_service.go`                                                        |
| 08 AI                            | Belum      | Hanya proto                                                                                                                                     | `proto/hris/ai/v1/ai.proto`                                                                                                     |
| 09-12 PAY, TAX, REC, PRF         | Ditunda    | ADR-0004                                                                                                                                        | `docs/adr/0004-defer-payroll-to-phase2.md`                                                                                      |
| Web                              | Sebagian   | Empat halaman; login dimock ke `http://localhost:3000/api/auth/login`                                                                           | `apps/web/src/routes/(auth)/login/+page.server.ts`                                                                              |
| Gateway                          | Sebagian   | Rute `auth/*` dan cuti; jalur HTTP ke layanan tidak dilayani                                                                                    | `apps/api-gateway/routes/api.php`                                                                                               |
