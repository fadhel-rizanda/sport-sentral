# Distributed Tracing dan Monitoring dengan OpenTelemetry & Jaeger

Catatan teknis ini menjelaskan arsitektur observabilitas, distributed tracing, dan metrics monitoring yang diimplementasikan di dalam sistem microservices ini menggunakan **OpenTelemetry** dan **Jaeger**.

---

## 1. Pendahuluan Distributed Tracing

Pada arsitektur monolith, debugging performa atau pelacakan bug cukup dilakukan dengan membaca berkas log tunggal. Namun pada arsitektur microservices, satu request dari klien dapat melewati rangkaian panjang panggilan gRPC antarlayanan sebelum respons dikembalikan.

**Distributed Tracing** memecahkan masalah ini dengan merekam perjalanan penuh satu request di sepanjang alirannya melintasi berbagai batasan jaringan layanan. Hasil rekaman ini disebut **Trace**, yang divisualisasikan dalam bentuk lini masa (timeline) terstruktur untuk mengetahui secara instan:
- Layanan (*service*) mana yang memakan waktu paling lama (bottleneck).
- Titik terdalam terjadinya error (misal query database yang gagal).
- Hubungan rantai pemanggilan antarlayanan.

### Contoh Visualisasi Aliran Trace:
```
GET /users/123 ➔ total durasi 850ms

├── API Gateway: 850ms
│   └── Panggil Identity Service (gRPC): 840ms
│
├── Identity Service: 840ms
│   ├── Validasi Token JWT: 2ms
│   ├── Query DB Users (GORM): 8ms
│   └── Panggil Meta Service (gRPC): 825ms ── [Bottleneck di sini]
│
└── Meta Service: 820ms
    ├── Deserialisasi Request: 5ms
    └── Query DB Statuses (GORM): 815ms ── [Index DB tidak aktif]
```

Dari hasil trace di atas, kita langsung mengetahui masalah utama terletak di database **Meta Service** karena query memakan waktu 815ms akibat ketiadaan indeks tabel.

---

## 2. Konsep Inti OpenTelemetry

### A. Trace dan Trace ID
Satu **Trace** mewakili keseluruhan perjalanan satu request. Setiap trace diidentifikasi dengan **Trace ID** unik (contoh: `00-4bf92f3577b34da6a3ce929d0e0e4736`). Seluruh potongan proses (span) dalam satu request akan memiliki Trace ID yang sama persis.

### B. Span dan Span ID
**Span** adalah unit kerja terkecil dalam trace yang merepresentasikan operasi tunggal (seperti query database, panggilan fungsi usecase, atau request gRPC). Setiap span memiliki **Span ID** unik, timestamp mulai/selesai, status (OK/Error), dan metadata penunjang:
- **Attributes**: Kunci-nilai statis untuk mendeskripsikan span (e.g. `db.system = "postgresql"`, `user.id = "123"`).
- **Events**: Log penanda waktu (*timestamped annotations*) di dalam rentang hidup span (seperti penanda `"cache miss, fetching from db"`).

```
Trace ID: abc-123
├── Span 1: GET /users/123             [0ms ──────────────────────── 850ms]
│     └── Span 2: gRPC GetUser         [5ms ──────────────────────── 845ms]
│           └── Span 3: GetByID Usecase [10ms ─────────────────────── 840ms]
│                 ├── Span 4: SQL Query [10ms ── 18ms] (durasi: 8ms)
│                 └── Span 5: gRPC Status [20ms ─────────────────────── 838ms]
```

### C. TracerProvider and Exporter
- **TracerProvider**: Pabrik (*factory*) global yang memproduksi objek `Tracer` untuk membuat span. Provider dikonfigurasi sekali di `main.go`.
- **Exporter**: Pengirim span dari aplikasi ke **OTel Collector** secara asinkron menggunakan protokol OTLP gRPC (port `4317`) dengan mekanisme batching untuk efisiensi performa memori.

---

## 3. Arsitektur Aliran Data Observabilitas

Komponen di dalam sistem berkolaborasi mengirimkan traces dan metrics ke backend visualisasi:

```
                                    ┌──────────────────────┐
                                    │ CLIENT (HTTP Request)│
                                    └──────────┬───────────┘
                                               │
                                               ▼
                                    ┌──────────────────────┐
                                    │     API GATEWAY      │
                                    │                      │
                                    │ otelfiber.Middleware │
                                    └────┬────────────┬────┘
                                         │            │
                              gRPC Call  │            │  gRPC Call
                              (Tracing)  ▼            ▼  (Tracing)
                       ┌──────────────────┐          ┌──────────────────┐
                       │ IDENTITY SERVICE │          │   META SERVICE   │
                       │                  │          │                  │
                       │ otelgrpc Server  │          │ otelgrpc Server  │
                       │ GORM + otelgorm  │          │ GORM + otelgorm  │
                       └─────────┬────────┘          └────────┬─────────┘
                                 │                            │
                                 │   Span Data (OTLP gRPC)    │
                                 └─────────────┬──────────────┘
                                               │
                                               ▼
                       ┌────────────────────────────────────────┐
                       │          OPENTELEMETRY COLLECTOR       │
                       │          (Port 4317 - OTLP gRPC)       │
                       │                                        │
                       │  - Menerima semua trace dari services  │
                       │  - Menerusan trace ke Jaeger          │
                       │  - Membuat RED metrics via spanmetrics │
                       │  - Mengekspos metrik di port 8889      │
                       └──────┬──────────────────────────┬──────┘
                              │                          │
               Export Traces  │                          │ Scrape Metrics
               (Port 14317)   ▼                          ▼ (Port 8889)
                       ┌──────────────┐          ┌──────────────────────┐
                       │    JAEGER    │◄─────────┤      PROMETHEUS      │
                       │ (Port 16686) │  Query   │     (Port 9090)      │
                       │              │  Metrics │                      │
                       │  Menyajikan  │          │  Menyimpan metrik    │
                       │ UI & Monitor │          │  span & app metrics  │
                       └──────────────┘          └──────────────────────┘
```

---

## 4. Alur Context Propagation (Cross-Process)

Agar rantai tracing tidak terputus saat request melompat melintasi batas jaringan (dari Gateway ke layanan gRPC internal), OTel menerapkan standar **W3C Trace Context**.

```
[ API Gateway ] (Memiliki Span Aktif)
     │
     │ 1. Client Handler menyuntikkan ID ke Metadata gRPC
     ▼
gRPC Header: traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
                               ^ Trace ID               ^ Parent Span ID   ^ Sampled
     │
     │ Panggilan Jaringan (Network Call)
     ▼
[ Identity Service ]
     │
     │ 2. Server Handler membaca Metadata (Extract)
     ▼
Reconstruct Context -> Buat Child Span baru yang terhubung ke Parent Span ID
```

> [!CAUTION]
> **Larangan Penggunaan context.Background() di dalam Layanan**:
> Jika Anda memanggil gRPC client dengan menyertakan `context.Background()`, Anda sedang memutuskan rantai propagasi context. Span baru yang terbentuk di layanan downstream tidak akan memiliki relasi parent-child dengan layanan upstream (menjadi trace yang terisolasi).
> - **Salah**: `client.GetUser(context.Background(), req)`
> - **Benar**: `client.GetUser(ctx, req)` (Selalu teruskan context aktif)

---

## 5. Instrumentasi Kode (Go Implementation)

### A. Instrumentasi Otomatis (Automatic Instrumentation)

Kita memanfaatkan pustaka middleware OTel untuk membuat span secara otomatis di gerbang masuk dan keluar:

```go
// 1. HTTP Server (Fiber) - Menangkap request HTTP masuk
app.Use(otelfiber.Middleware())

// 2. gRPC Server - Menangkap incoming gRPC request
grpcServer := grpc.NewServer(
    grpc.StatsHandler(otelgrpc.NewServerHandler()),
)

// 3. gRPC Client - Menyuntikkan Trace Context ke outgoing request
conn, _ := grpc.NewClient(address,
    grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
)

// 4. Database (GORM) - Tracing query SQL otomatis
db.Use(otelgorm.NewPlugin(otelgorm.WithTracerProvider(otel.GetTracerProvider())))
```

### B. Instrumentasi Manual (Usecase / Logic Layer)

Gunakan instrumentasi manual untuk mendokumentasikan fungsionalitas bisnis krusial atau mengisolasi bagian kode yang diidentifikasi lambat:

```go
func (uc *userUseCase) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.User, error) {
    // 1. Ambil tracer global berdasarkan nama service
    tracer := otel.Tracer("identity-service")

    // 2. Mulai span baru. Objek `ctx` baru berisi span aktif ini
    ctx, span := tracer.Start(ctx, "userUseCase.Register",
        trace.WithAttributes(
            attribute.String("user.email", req.Email),
        ),
    )
    defer span.End() // Memastikan span ditutup saat eksekusi fungsi selesai

    // 3. Eksekusi proses bisnis
    user, err := uc.repo.CreateUser(ctx, req)
    if err != nil {
        span.RecordError(err)                    // Catat pesan error di span log
        span.SetStatus(codes.Error, err.Error()) // Tandai span sebagai Error
        return nil, err
    }

    span.SetAttributes(attribute.String("user.id", user.ID.String()))
    return user, nil
}
```

---

## 6. Service Performance Monitoring (SPM) via Spanmetrics

Selain menampilkan visualisasi trace per request, sistem kita mengintegrasikan metrik kinerja layanan (**SPM**) menggunakan konektor `spanmetrics` pada OpenTelemetry Collector.

OTel Collector secara otomatis mengumpulkan span yang selesai, mengekstrak statistiknya, lalu mengeksposnya sebagai metrik **RED**:
1. **R**equest Rate (Jumlah request masuk per detik).
2. **E**rror Rate (Persentase request yang gagal).
3. **D**uration (P50, P90, dan P95 Latensi kuantil).

Metrik deret waktu ini ditarik (*scraped*) oleh **Prometheus** dan ditampilkan kembali ke dalam dashboard **Monitor** di **Jaeger UI** secara real-time.

---

## 7. Praktik Terbaik Observabilitas (Best Practices)

- **Sanitisasi Data Sensitif**: Jangan pernah merekam data sensitif (seperti password, nomor kartu kredit, token JWT) di dalam tag `Attributes` atau `Events` span.
- **Nama Span yang Konsisten**: Gunakan format `Layer.Fungsi` (contoh: `userUseCase.Register`, `userRepository.GetByEmail`). Hindari nama generik seperti `process` atau `handle`.
- **Sampling di Production**: Di lingkungan development, kita merekam 100% request (`AlwaysSample()`). Di production, gunakan sampling rasional (misalnya 5% - 10% request) menggunakan `ParentBased(TraceIDRatioBased(0.05))` untuk menghemat disk storage dan mengurangi overhead jaringan.
