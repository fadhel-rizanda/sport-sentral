# Structured Logging dengan Zap

Catatan teknis ini menjelaskan implementasi logging terstruktur menggunakan pustaka **`go.uber.org/zap`** untuk menjamin performa tinggi, kemudahan pencarian log di server production, dan standarisasi visual di lingkungan development.

---

## 1. Desain Arsitektur Logging

Dalam arsitektur microservices ini, penulisan log dilakukan secara selektif untuk meminimalkan polusi log (*log noise*) namun tetap mempertahankan visibilitas penuh:

```
[ HTTP/gRPC Request ]
         │
         ▼
 ┌───────────────┐
 │  Interceptor  │ ➔ Log Otomatis: Route/Method, Latency, Status Code, & Error
 └───────┬───────┘
         │
         ▼
 ┌───────────────┐
 │    Usecase    │ ➔ Log Manual: Khusus Business Events krusial (e.g. "user.created")
 └───────┬───────┘
         │
         ▼
 ┌───────────────┐
 │  Repository   │ ➔ TIDAK PERLU LOG: Error cukup di-bubble up (return) ke atas
 └───────────────┘
```

- **Interceptor (Gerbang Depan)**: Semua log HTTP/gRPC request masuk ditangani di level interceptor secara otomatis.
- **Usecase (Log Bisnis)**: Hanya mencatat kejadian mutasi bisnis krusial (seperti proses registrasi berhasil, pembayaran, atau penghapusan akun). Operasi baca (read-only) seperti `GetUser` tidak perlu di-log untuk menghemat resource disk.
- **Repository / Database Layer**: Dilarang menulis log error di tingkat repository. Cukup kembalikan error tersebut ke atas (*bubble up*) agar ditangkap oleh interceptor di gerbang paling depan.

---

## 2. Tingkat Log (Log Levels)

Penerapan tingkat kepentingan log wajib mengikuti standar berikut:

| Level | Penggunaan | Kapan Digunakan |
| :--- | :--- | :--- |
| **`Debug`** | Detailing internal developer | Hanya aktif di mode development. Berisi info payload detail, query database, dan lain-lain. |
| **`Info`** | Alur bisnis berjalan normal | Kejadian penting sistem (e.g. "service started", "user registered successfully"). |
| **`Warn`** | Masalah kecil / Input tidak valid | Masalah yang diakibatkan oleh client (e.g. bad request, unauthorized, resource not found). |
| **`Error`** | Kegagalan sistem / Masalah server | Kegagalan infrastruktur atau bug (e.g. koneksi DB terputus, query SQL rusak, file system error). |
| **`Fatal`** | Kegagalan kritis startup | Aplikasi tidak dapat berjalan dan langsung berhenti (`os.Exit(1)`). |

---

## 3. Cara Penggunaan di Kode

### A. Inisialisasi Singleton
Logger diinisialisasi sekali di dalam `main.go` saat startup aplikasi berjalan:

```go
// Menyesuaikan format log berdasarkan environment ("production" -> JSON, "development" -> Colorized Console)
log := applogger.New(os.Getenv("APP_ENV"))
defer log.Sync() // Melakukan flush sisa buffer log ke disk sebelum aplikasi mati
```

### B. Memanggil Logger Terstruktur
Gunakan logger global di mana saja tanpa perlu melakukan dependency injection:

```go
import (
    "go.uber.org/zap"
    "microservice-golang/shared/pkg/logger"
)

// Menulis log terstruktur dengan field type-safe
logger.Get().Info("user created successfully",
    zap.String("user_id", user.ID.String()),
    zap.String("email", user.Email),
)
```

---

## 4. Penanganan Panic dan Tracing gRPC Interceptors

Untuk mengamankan gRPC server dari crash tak terduga (panic), kita menyusun rantai interceptor (*interceptor chain*) di mana Logger bertindak sebagai pembungkus luar dan Recovery sebagai pelindung dalam.

```go
grpcServer := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        interceptor.UnaryLogger(log),   // Pembungkus Luar (Outer)
        interceptor.UnaryRecovery(log), // Penyelamat Dalam (Inner)
    ),
)
```

### UnaryLogger
Secara otomatis mengukur waktu eksekusi (latency) setiap request gRPC dan mencatat metadata penting seperti nama method, durasi, status code, serta isi error jika pemanggilan gagal.

### UnaryRecovery
Jika terjadi panic (runtime crash) di dalam kode Usecase atau Handler:
1. `UnaryRecovery` menangkap panic tersebut secara otomatis (*graceful recovery*).
2. Mencatat stack trace lengkap ke file log dengan level `Error`.
3. Mengembalikan respons error standar gRPC `codes.Internal` ke sisi client agar struktur internal sistem tidak bocor ke luar.

---

## 5. Daftar Field Terstruktur Standar (Common Fields)

Demi kemudahan analisis log menggunakan log aggregator (seperti Grafana Loki atau ELK Stack), selalu gunakan penamaan bidang (*key fields*) yang konsisten:

```go
zap.Error(err)                         // Untuk menyertakan pesan error
zap.String("user_id", id)              // ID entitas pengguna
zap.String("method", info.FullMethod)  // Nama endpoint RPC/HTTP
zap.String("code", st.Code().String()) // Status respons (e.g. "InvalidArgument")
zap.Duration("duration", elapsed)      // Latensi eksekusi
zap.ByteString("stack", debug.Stack()) // Data stack trace saat terjadi panic
```