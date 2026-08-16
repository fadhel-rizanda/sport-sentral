# Struktur dan Arsitektur Umum Sistem

Dokumen ini menjelaskan rancangan arsitektur, struktur repositori, pembagian modul, serta pola interaksi antar-komponen di dalam ekosistem microservices ini secara komprehensif.

---

## 1. Topologi Arsitektur Sistem

Sistem ini dirancang menggunakan arsitektur microservices terdistribusi dengan pemisahan tanggung jawab yang jelas untuk setiap layanannya. Aliran data utama dibagi menjadi tiga kategori komunikasi: sinkron (gRPC), asinkron (NATS JetStream), dan pemantauan (OpenTelemetry).

```
                                 [ Klien Publik ]
                                        │
                                        │ HTTP/REST
                                        ▼
                               ┌─────────────────┐
                               │   API Gateway   │
                               │  (HTTP Server)  │
                               └┬───────────────┬┘
                                │               │
                     gRPC Call  │               │  gRPC Call
                     (Synchronous)              │  (Synchronous)
                                ▼               ▼
                    ┌──────────────────┐  ┌──────────────────┐
                    │ Identity Service │  │   Meta Service   │
                    │  (gRPC Backend)  │  │  (gRPC Backend)  │
                    └────────┬─────────┘  └────────┬─────────┘
                             │                     │
                             │   Asynchronous      │
                             │   Event Stream      │
                             └──────────┬──────────┘
                                        ▼
                                ┌───────────────┐
                                │ NATS Event Bus│
                                │  (JetStream)  │
                                └───────────────┘
```

### Penjelasan Peran Komponen:

1. **API Gateway (Public Entrypoint)**:
   - Berfungsi sebagai satu-satunya gerbang masuk bagi permintaan klien luar melalui protokol HTTP/REST.
   - Bertanggung jawab atas routing permintaan, validasi otentikasi awal menggunakan JSON Web Token (JWT), penanganan pembatasan akses (*rate limiting*), serta bertindak sebagai gRPC client proxy yang meneruskan permintaan ke microservices internal.

2. **Identity Service (Internal gRPC Backend)**:
   - Mengelola fungsionalitas inti otentikasi pengguna, otorisasi peran (*Role-Based Access Control* / RBAC), profil pengguna, serta hak akses izin (*permissions*).
   - Menyediakan endpoint berbasis gRPC untuk dikonsumsi oleh API Gateway.

3. **Meta Service (Internal gRPC Backend)**:
   - Mengelola data metadata pendukung, kategori status, serta pengelompokan label (*tags*).
   - Menyediakan endpoint berbasis gRPC untuk kebutuhan klasifikasi data internal.

4. **NATS JetStream (Asynchronous Event Bus)**:
   - Berfungsi sebagai broker pesan bertenaga tinggi untuk memfasilitasi integrasi asinkron antarlayanan (*event-driven integration*).
   - Digunakan untuk melakukan sinkronisasi data antar-database layanan secara seketika (*real-time data synchronization*) tanpa mengganggu kinerja request gRPC utama.

---

## 2. Struktur Ruang Kerja Proyek (Monorepo Layout)

Proyek ini dikembangkan di dalam satu repositori tunggal (*monorepo*) dengan pemisahan modul Go yang diintegrasikan melalui mekanisme **Go Workspaces (`go.work`)**.

```
microservice-golang/
├── documentation/                 # Dokumentasi panduan teknis proyek
├── gen/                           # Berkas kode Go hasil generate dari proto (Modul Go)
├── proto/                         # Defenisi kontrak API Protocol Buffers (Modul Go)
├── services/                      # Direktori kumpulan layanan mandiri
│   ├── academy-service/           # Layanan Akademi & Roster
│   ├── attachment-service/        # Layanan Manajemen Berkas & Object Storage (S3/MinIO)
│   ├── competition-service/      # Layanan Kompetisi & Turnamen
│   ├── gateway/                   # API Gateway (Fiber Framework) (Modul Go)
│   ├── identity-service/          # Layanan Identitas dan Akun Pengguna (Modul Go)
│   ├── log-service/               # Layanan Audit Log & Aktivitas
│   ├── meta-service/              # Layanan Status dan Metadata (Modul Go)
│   ├── scout-service/             # Layanan Talent Scouting & Profil Atlet
│   ├── sport-service/             # Layanan Olahraga & Regulating Body
│   └── venue-service/             # Layanan Lapangan & Pemesanan (Booking)
├── shared/                        # Pustaka utilitas bersama (Modul Go)
├── Procfile                       # Daftar konfigurasi startup aplikasi terdistribusi
└── go.work                        # Konfigurasi workspace Go lintas modul
```

### Pemetaan Workspace (`go.work`):
Sistem memanfaatkan fitur Go Workspace untuk membebaskan setiap layanan memiliki modul dependensi (`go.mod`) sendiri, namun tetap dapat mereferensikan modul lokal lainnya (seperti `shared` dan `gen`) secara langsung tanpa perlu melakukan proses publish eksternal.

---

## 3. Desain Pola Internal Layanan (Clean Architecture)

Layanan internal backend (`identity-service` dan `meta-service`) mengadopsi prinsip **Clean Architecture** (Arsitektur Bersih) guna memisahkan logika bisnis dari detail infrastruktur teknologi luar.

```
                      ┌─────────────────────────────────┐
                      │            DELIVERY             │
                      │  - gRPC Handlers / Servers      │
                      │  - NATS Publishers/Subscribers  │
                      └────────────────┬────────────────┘
                                       │
                                       ▼
                      ┌─────────────────────────────────┐
                      │             USECASE             │
                      │  - Logika Bisnis Utama (Pure Go)│
                      │  - Orkestrasi Data & Validasi   │
                      └────────────────┬────────────────┘
                                       │
                                       ▼
                      ┌─────────────────────────────────┐
                      │           REPOSITORY            │
                      │  - Kueri Database (Postgres)    │
                      │  - Operasi Redis Cache          │
                      └─────────────────────────────────┘
```

Setiap modul internal layanan dibagi menjadi beberapa lapisan berikut:

1. **Lapisan Entity (`internal/entity`)**:
   - Berisi definisi objek bisnis murni, struktur model data, serta fungsi-fungsi aturan bisnis dasar (*domain logic*).

2. **Lapisan Repository (`internal/repository`)**:
   - Bertanggung jawab penuh atas transaksi data ke media penyimpanan fisik.
   - Mengisolasi detail implementasi kueri SQL (menggunakan GORM PostgreSQL) dan penyimpanan memori sementara (Redis caching) dari lapisan atas.

3. **Lapisan Usecase (`internal/usecase`)**:
   - Tempat bernaungnya logika bisnis utama sistem. Lapisan ini mengatur aliran data dari dan ke lapisan repository, melakukan orkestrasi proses, serta melakukan penegakan aturan validasi bisnis. Lapisan ini tidak boleh bergantung pada teknologi pengiriman data luar (seperti HTTP atau gRPC).

4. **Lapisan Delivery (`internal/delivery`)**:
   - Gerbang interaksi layanan dengan jaringan eksternal. Terbagi menjadi dua komponen:
     - **gRPC Handler**: Menerima request gRPC dari API Gateway, memproses parameter, memanggil usecase terkait, dan menyusun kembali response biner proto.
     - **NATS Publisher/Subscriber**: Menerbitkan event mutasi ke NATS JetStream, atau berlangganan (*subscribe*) event dari stream layanan lain untuk disinkronkan ke dalam database lokal.

---

## 4. Aliran Komunikasi Terdistribusi

### Komunikasi Sinkron (Synchronous Path)
Digunakan untuk operasi baca-tulis langsung di mana klien membutuhkan respons seketika:
- Klien mengirim request HTTP REST ke API Gateway.
- API Gateway mengekstrak data request, memvalidasi JWT token, dan mengubahnya menjadi format request biner gRPC.
- Gateway mengirim kueri ke gRPC Server internal (`identity-service` atau `meta-service`) melalui HTTP/2.
- Layanan internal mengeksekusi request, memperbarui database relasional PostgreSQL, dan mengembalikan respons secara langsung.

### Komunikasi Asinkron (Asynchronous Path)
Digunakan untuk sinkronisasi data antar-database microservices agar tetap konsisten (*eventual consistency*) tanpa memblokir request utama:
- Pengguna mendaftarkan akun baru via API Gateway.
- `identity-service` memproses pendaftaran dan memperbarui database lokal `users`.
- Segera setelah penyimpanan sukses, `identity-service` menerbitkan event `identity.user.created` berisi payload data user baru ke `IDENTITY_EVENTS` Stream di NATS JetStream.
- Proses pendaftaran pengguna langsung dinyatakan selesai dan respons sukses dikembalikan ke Gateway (tanpa menunggu proses sinkronisasi selesai).
- Secara asinkron, `meta-service` yang berlangganan (*subscribe*) subjek tersebut menangkap event baru dan memperbarui database lokalnya secara mandiri agar memiliki referensi data pengguna yang valid.

---

## 5. Arsitektur Telemetry dan Observabilitas

Untuk memastikan visibilitas penuh terhadap kesehatan sistem terdistribusi, diimplementasikan kerangka pemantauan tiga pilar (Traces, Metrics, dan Logs) terintegrasi menggunakan **OpenTelemetry**:

```
[ Pemicu Request ] ➔ Inisialisasi Span Terkait (Gateway)
                          │
                          ▼ (Propagasi via gRPC Metadata)
[ Downstream Services ] ➔ Teruskan Span ID & Buat Child Span (Identity/Meta)
                          │
                          ▼ (Propagasi via GORM Context)
[ Database Queries ] ➔ Buat Span Kueri SQL Terkait
                          │
                          ▼ (Export secara Batch - OTLP gRPC)
┌───────────────────────────────────────────────────────────────┐
│                    OpenTelemetry Collector                    │
│  - Menerima spans dari semua layanan                          │
│  - Menghasilkan RED Metrics menggunakan modul spanmetrics      │
└────────────────┬──────────────────────────────┬───────────────┘
                 │                              │
                 ▼ (Export Traces)              ▼ (Scrape Metrics)
          ┌──────────────┐              ┌──────────────┐
          │  Jaeger UI   │              │  Prometheus  │
          │ (Trace View) │              │ (Deret Waktu)│
          └──────────────┘              └──────────────┘
```

1. **Distributed Tracing**:
   Setiap request diberi Trace ID unik di API Gateway. Context tracing disalurkan melintasi jaringan gRPC via metadata header W3C, lalu diteruskan ke GORM database query. Seluruh span yang terkumpul dikirimkan ke **Jaeger** untuk visualisasi lini masa detail.
2. **Metrics Monitoring**:
   OpenTelemetry Collector mengagregasi status durasi eksekusi span menjadi metrik kinerja berupa data kuantil latensi (P50, P90, P95), request rate, dan error rate (RED metrics). Metrik ini disimpan di **Prometheus** dan divisualisasikan pada menu Monitor Jaeger.
3. **Structured Logging**:
   Semua penulisan log dienkapsulasi menggunakan pustaka berkinerja tinggi **Zap Logger**. Format log di lingkungan produksi berupa JSON terstruktur untuk memudahkan agregasi berkas log terpusat.
