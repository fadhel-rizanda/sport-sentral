# Panduan Pengembangan dan Lembar Ringkasan (Cheatsheet)

Dokumen ini berisi panduan praktis harian untuk alur pengembangan microservices, contoh kueri database GORM, manajemen migrasi, penataan container telemetry di lingkungan Windows (Podman/Docker), serta metode penyelesaian masalah (troubleshooting) jaringan terdistribusi.

---

## 1. Manajemen Pembuatan Service Baru

Gunakan langkah-langkah terstruktur berikut jika Anda ingin membuat modul layanan microservice baru di dalam monorepo:

```powershell
# 1. Masuk ke direktori gateway baru (contoh)
cd services/gateway
go mod init microservice-golang/services/gateway

# 2. Lakukan sinkronisasi Go Workspace di root folder proyek
cd ../..
go work sync
go work use -r .
```

---

## 2. Pola Penulisan Query Database GORM (dengan Context)

> [!IMPORTANT]
> **Selalu sertakan Context (`ctx`)**:
> Pastikan setiap pemanggilan query GORM diawali dengan `.WithContext(ctx)`. Tanpa context, metadata Distributed Tracing (Span OTel) tidak akan tersalurkan ke tingkat database, sehingga Anda tidak dapat menganalisis latensi kueri SQL di Jaeger UI.

```go
// 0. Inisialisasi Query Dasar (Base Query)
query := r.db.WithContext(ctx)

// 1. Filter Logika Dasar (AND, OR, NOT)
query.Where("type = ?", t).Where("active = ?", true) // AND (Secara Implisit)
query.Or("is_priority = ?", true)                    // Kondisi OR
query.Not("status = ?", "deleted")                   // Kondisi NOT

// 2. Filter Pencarian Lanjutan (LIKE, IN, BETWEEN)
query.Where("name LIKE ?", "%"+search+"%")           // Pencarian Teks (Search)
query.Where("id IN ?", []string{"uuid1", "uuid2"})   // Kueri Multi-ID
query.Where("created_at BETWEEN ? AND ?", start, end)// Rentang Waktu / Tanggal

// 3. Eager Loading & Pembatasan Kolom
query.Preload("Permissions")                                 // Eager load relasi (Query terpisah)
query.Joins("JOIN profiles ON profiles.user_id = users.id")  // Join Tabel secara Manual
query.Select("id", "name", "slug")                           // Mengambil kolom tertentu saja
query.Omit("password", "internal_note")                      // Mengecualikan kolom sensitif

// 4. Urutan & Halaman (Sorting, Limit & Pagination)
query.Order("created_at DESC")                       // Urutan terbaru
query.Limit(10)                                      // Batasi jumlah data yang diambil
query.Offset(20)                                     // Skip data (untuk halaman ke-3)

// 5. Eksekusi Kueri Baca (Read Query)
err := query.Find(&items).Error                      // Mengambil banyak record (List)
err := query.First(&item).Error                      // Mengambil satu data (memicu error jika kosong)

// 6. Eksekusi Mutasi Data (Write Query)
err := r.db.Create(&item).Error                      // Insert data baru
err := r.db.Save(&item).Error                        // Upsert (Simpan semua bidang struct)
err := r.db.Model(&item).Update("name", "new").Error // Update satu kolom saja
err := r.db.Model(&item).Updates(mapData).Error      // Update banyak kolom menggunakan Map
err := r.db.Delete(&item).Error                      // Hapus data (Soft delete jika DeletedAt aktif)
```

---

## 3. Skema dan Migrasi Database

Proyek ini menggunakan pustaka `golang-migrate` untuk melacak perubahan skema database secara berurutan (*versioned migrations*):

```powershell
# 1. Instalasi Tool CLI golang-migrate (dengan tag driver PostgreSQL)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 2. Membuat file migrasi SQL baru secara sekuensial
migrate create -ext sql -dir db/migrations -seq create_statuses_table
# Perintah di atas akan menghasilkan dua file di folder db/migrations/:
# - [nomor]_create_statuses_table.up.sql   (Berisi DDL untuk menambah tabel)
# - [nomor]_create_statuses_table.down.sql (Berisi DDL untuk merestore/menghapus tabel)
```

---

## 4. Cara Menjalankan Seluruh Aplikasi Secara Bersamaan

Proyek ini mendefinisikan seluruh microservice di dalam berkas `Procfile`. Kita dapat menggunakan tool `goreman` untuk menyalakan semua service dengan satu perintah:

```powershell
# Jalankan goreman dari direktori root proyek
goreman start
```

---

## 5. Konfigurasi Container Telemetry di Windows (Podman/Docker)

> [!IMPORTANT]
> **Penyelarasan Jaringan WSL2 (Mirrored Mode)**:
> Di lingkungan Windows dengan WSL2 yang mengaktifkan mirrored networking, seluruh container monitoring wajib dijalankan dengan flag `--network host` agar dapat berkomunikasi langsung menggunakan localhost dan menghindari masalah Firewall Windows.
> **Path Absolut Windows**: Selalu gunakan path Windows absolut (lengkap) saat melakukan mount volume file konfigurasi `.yml` di Windows CLI.

```powershell
# 1. Jalankan Jaeger (In-Memory Trace & Monitor Collector)
podman run -d --name sport-sentral-jaeger `
  --network host `
  -e COLLECTOR_OTLP_GRPC_HOST_PORT=0.0.0.0:14317 `
  -e METRICS_STORAGE_TYPE=prometheus `
  -e PROMETHEUS_SERVER_URL=http://localhost:9090 `
  -e PROMETHEUS_QUERY_SUPPORT_SPANMETRICS_CONNECTOR=true `
  -e PROMETHEUS_QUERY_NAMESPACE=traces_span_metrics `
  -e PROMETHEUS_QUERY_NORMALIZE_DURATION=true `
  -e PROMETHEUS_QUERY_NORMALIZE_CALLS=true `
  jaegertracing/all-in-one:latest

# 2. Jalankan Prometheus (Basis Data Metrik Deret Waktu)
podman run -d --name sport-sentral-prometheus `
  --network host `
  -v C:\MAGANG_mandiri\microservice-golang\prometheus.yml:/etc/prometheus/prometheus.yml `
  prom/prometheus

# 3. Jalankan OpenTelemetry (OTel) Collector
podman run -d --name sport-sentral-otel-collector `
  --network host `
  -v C:\MAGANG_mandiri\microservice-golang\otel-collector.yml:/etc/otelcol-contrib/config.yaml `
  otel/opentelemetry-collector-contrib:latest
```

---

## 6. Panduan Penyelesaian Masalah Tracing Terdistribusi (Step-by-Step)

Jika grafik tab **Monitor** di Jaeger kosong atau trace antar-layanan terputus, ikuti prosedur diagnostik di bawah ini:

### Langkah 1: Periksa Kesehatan Container
```powershell
# Memeriksa apakah ada container yang mati secara mendadak (Exited)
podman ps -a

# Jika container mendadak exit, segera buka log kesalahannya:
podman logs sport-sentral-otel-collector
```
- **Penyebab Utama**: OTel Collector dapat crash jika volume mount mengarah ke `/etc/otelcol/config.yaml`. Pastikan path tujuan di dalam container adalah `/etc/otelcol-contrib/config.yaml` (untuk versi contrib).

### Langkah 2: Verifikasi Jaringan (Host dan Container WSL)
Jika Anda tidak menggunakan `--network host`, container sering kali terisolasi oleh firewall internal WSL.
```powershell
# Uji DNS internal resolver di container
podman run --rm alpine nslookup host.containers.internal

# Uji apakah container dapat memanggil port microservice di host Windows Anda
podman run --rm alpine wget -S --spider http://10.88.0.1:8080
```
- **Solusi**: Jika terjadi timeout, gunakan opsi jaringan `--network host` agar semua container berada di domain localhost yang sama dengan aplikasi Go Anda.

### Langkah 3: Mengatasi Monitor Jaeger yang Kosong ("No Data yet!")
Jika Tracing masuk namun tab visualisasi RED Metrics kosong:
1. Pastikan metrik span terkumpul di Prometheus dengan memanggil query PromQL:
   ```powershell
   curl -s "http://localhost:9090/api/v1/query?query=traces_span_metrics_calls_total"
   ```
2. Pastikan flag normalisasi durasi dan panggilan (`PROMETHEUS_QUERY_NORMALIZE_*`) pada container Jaeger sudah bernilai `true`.

### Langkah 4: Pastikan Propagasi Context Tidak Terputus
Pastikan semua pemanggilan gRPC client atau Usecase meneruskan `ctx` aktif dari request framework, **bukan** membuat context kosong baru:
```go
// Salah: Rantai Trace/Parent ID terputus secara keseluruhan
resp, err := h.authClient.Login(context.Background(), req)

// Benar: Meneruskan Trace ID aslinya
resp, err := h.authClient.Login(c.UserContext(), req)
```
