# Spesifikasi dan Konfigurasi Swagger (OpenAPI) API Gateway

Dokumen ini memuat panduan teknis, standar konvensi anotasi, arsitektur integrasi, dan prosedur pemeliharaan dokumentasi API berbasis **Swagger (OpenAPI 2.0)** pada layanan **API Gateway** (`services/gateway`) di dalam ekosistem *microservices* **Sport-Sentral**.

---

## 1. Arsitektur Dokumentasi (*Documentation Overview*)

Dokumentasi API Gateway dikelola menggunakan perkakas [Swaggo (swag)](https://github.com/swaggo/swag). Dokumentasi ini bertindak sebagai *Single Source of Truth* antarmuka REST API publik yang mengagregasikan seluruh *downstream microservices* (Identity, Meta, Academy, Sport, Competition, Scout, Venue, Attachment, dan Log Service).

### Informasi Akses Interaktif (Swagger UI)
* **Protokol:** HTTP/REST
* **Base Path:** `/api/v1`
* **Swagger JSON Spec:** `/swagger/doc.json`
* **Swagger UI URL:** `http://localhost:8080/swagger/index.html` (atau sesuai nilai `APP_PORT`)

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         Kompilasi Dokumentasi                            │
└──────────────────────────────────────────────────────────────────────────┘
           Source Code Go & Anotasi              Generator Swag CLI
   [ cmd/main.go & internal/handler/*.go ] ───> [ swag init -g ... ]
                                                        │
                                                        ▼
                                            ┌───────────────────────┐
                                            │    docs/docs.go       │
                                            │    docs/swagger.json  │
                                            │    docs/swagger.yaml  │
                                            └───────────┬───────────┘
                                                        │
                                                        ▼
                                            ┌───────────────────────┐
                                            │   Fiber Swagger UI    │
                                            │   /swagger/index.html │
                                            └───────────────────────┘
```

---

## 2. Konfigurasi Metadata Global (*General API Info*)

Informasi umum API dideklarasikan pada bagian awal fungsi `main()` di berkas [`cmd/main.go`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/cmd/main.go):

```go
// @title                      Sport-Sentral API Gateway
// @version                    1.0
// @description                Centralized API Documentation for Sport-Sentral Microservices Gateway (Auth, RBAC, Academy, Venue, Sport, Competition, Scout, Attachment, Log).
// @termsOfService             http://swagger.io/terms/

// @contact.name               Sport-Sentral Development Team
// @contact.email              dev@sportsentral.id

// @license.name               Apache 2.0
// @license.url                http://www.apache.org/licenses/LICENSE-2.0.html

// @host                       localhost:8000
// @BasePath                   /api/v1

// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Enter JWT token in the format: `Bearer <access_token>`
```

### Rincian Direktif Global

| Direktif Swagger | Nilai Standar | Keterangan |
| :--- | :--- | :--- |
| `@title` | `Sport-Sentral API Gateway` | Nama resmi portal API. |
| `@version` | `1.0` | Versi rilis spesifikasi OpenAPI. |
| `@description` | *Deskripsi Terpusat* | Gambaran umum cakupan modul gateway. |
| `@host` | `localhost:8000` | Host default untuk pengujian interaktif. |
| `@BasePath` | `/api/v1` | Prefix dasar seluruh rute HTTP API. |
| `@securityDefinitions.apikey` | `BearerAuth` | Skema autentikasi berbasis JWT Bearer Token. |
| `@in` | `header` | Lokasi penyematan kredensial otorisasi. |
| `@name` | `Authorization` | Nama HTTP Header (`Authorization: Bearer <token>`). |

---

## 3. Standar Penulisan Anotasi Handler (*Handler Annotation Standards*)

Sesuai dengan standarisasi proyek, seluruh ringkasan (`@Summary`), deskripsi (`@Description`), dan keterangan parameter (`@Param`) wajib ditulis dalam **Bahasa Inggris formal dan ringkas**.

### Pola Standar Anotasi Endpoint

```go
// EndpointName godoc
// @Summary      Short action summary in imperative English
// @Description  Detailed explanation of business flow, validation rules, and effects.
// @Tags         ModuleName
// @Accept       json
// @Produce      json
// @Param        param_name  param_type  data_type  required  "Parameter description in English"
// @Success      200         {object}    response.Response{data=dto.TargetResponse}
// @Failure      400         {object}    response.Response
// @Failure      401         {object}    response.Response
// @Failure      403         {object}    response.Response
// @Failure      404         {object}    response.Response
// @Failure      500         {object}    response.Response
// @Router       /path/endpoint [method]
// @Security     BearerAuth
```

### Komponen Kunci Anotasi

1. **`@Summary`**: Deskripsi singkat aksi (contoh: `Create new academy holding`, `Get user details by ID`).
2. **`@Description`**: Informasi kontekstual yang menjelaskan proses bisnis atau *side-effects* (contoh: *Soft-delete academy branch and cascade status updates*).
3. **`@Tags`**: Label pengelompokan pada antarmuka Swagger UI.
4. **`@Param`**: Deklarasi parameter permintaan dengan format:
   `[nama] [lokasi: path|query|header|formData|body] [tipe_data] [required: true|false] "[deskripsi]"`
5. **`@Success` & `@Failure`**: Pemetaan kode status HTTP ke struktur Go DTO atau pembungkus umum `response.Response`.
6. **`@Router`**: Jalur URL relatif terhadap `@BasePath` beserta metode HTTP dalam huruf kecil (contoh: `/admin/matches/{id}/status [put]`).
7. **`@Security`**: Disematkan jika endpoint dilindungi middleware autentikasi JWT (`@Security BearerAuth`).

---

## 4. Taksonomi Tag dan Modul Domain

Seluruh rute pada API Gateway dikategorikan ke dalam tag-tag berikut:

| Nama Tag (`@Tags`) | Lingkup Domain Bisnis | Berkas Handler |
| :--- | :--- | :--- |
| `Auth` | Autentikasi, pendaftaran, login, token refresh, reset password | [`internal/handler/auth.go`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/internal/handler/auth.go) |
| `Users` | Profil user, verifikasi akun, manajemen penugasan role | [`internal/handler/user.go`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/internal/handler/user.go) |
| `Roles` | Manajemen Role RBAC | [`internal/handler/role.go`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/internal/handler/role.go) |
| `Permissions` | Manajemen izin akses individual (*Permission*) | [`internal/handler/permission.go`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/internal/handler/permission.go) |
| `Profiles` | Pengajuan, pengaktifan, dan verifikasi profil spesialis | [`internal/handler/profile.go`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/internal/handler/profile.go) |
| `Academies` | Holding akademi, cabang, admin, pendaftaran atlet, dan roster | [`internal/handler/academy.go`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/internal/handler/academy.go) |
| `Competitions` | Turnamen, kejuaraan, babak, fase pertandingan | [`internal/handler/competition.go`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/internal/handler/competition.go) |
| `Matches` | Jadwal pertandingan, skor laga, manajemen partisipan | [`internal/handler/competition.go`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/internal/handler/competition.go) |
| `Match Stats` | Statistik atlet per laga, box score, agregasi turnamen | [`internal/handler/competition.go`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/internal/handler/competition.go) |
| `Scouts` | Profil pencari bakat, watchlist, leaderboard, activity logs | [`internal/handler/scout.go`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/internal/handler/scout.go) |
| `Sports` | Cabang olahraga, badan regulator, staf, konfigurasi statistik | [`internal/handler/sport.go`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/internal/handler/sport.go) |
| `Venues` | Gelanggang olahraga, lapangan (*court*), slot jadwal, pemesanan | [`internal/handler/venue.go`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/internal/handler/venue.go) |
| `Attachments` | Presigned URL, unggah langsung *multipart*, metadata berkas | [`internal/handler/attachment.go`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/internal/handler/attachment.go) |
| `Logs` | Audit trail perubahan sistem, user activity log, metrik log | [`internal/handler/log.go`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/internal/handler/log.go) |
| `Statuses` | Status referensi global lintas entitas | [`internal/handler/status.go`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/internal/handler/status.go) |
| `Tags` | Tag kategori, format, dan klasifikasi metadata | [`internal/handler/tag.go`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/internal/handler/tag.go) |

---

## 5. Struktur Respons dan Pembungkus Generik (*Response Envelopes*)

API Gateway menggunakan struktur pembungkus standar [`response.Response`](file:///C:/MAGANG_mandiri/microservice-golang/services/gateway/internal/response/response.go) untuk seluruh keluaran JSON.

### Skema Go Response Envelope
```go
type Response struct {
    Success bool        `json:"success"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Errors  interface{} `json:"errors,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
    Page     int   `json:"page"`
    PageSize int   `json:"page_size"`
    Total    int64 `json:"total"`
}
```

### Contoh Deklarasi Swagger untuk Berbagai Tipe Respons

#### 1. Respons Tunggal (*Single Object*)
```go
// @Success 200 {object} response.Response{data=dto.CompetitionResponse}
```

#### 2. Respons Daftar dengan Metadata Paginasi (*List with Pagination*)
```go
// @Success 200 {object} response.Response{data=[]dto.CompetitionResponse}
```

#### 3. Respons Pesan Sukses Tanpa Data (*Message Only / 204 No Content*)
```go
// @Success 200 {object} response.Response
// @Success 204
```

#### 4. Respons Upload Berkas Bertanda Tangan (*Presigned URL*)
```go
// @Success 200 {object} response.Response{data=dto.PresignedUploadURLResponse}
```

---

## 6. Prosedur Regenerasi Dokumen (*CLI Regeneration Workflow*)

Setiap kali terjadi penambahan endpoint baru atau pembaruan anotasi komentar Go, berkas Swagger harus dikompilasi ulang menggunakan CLI `swag`.

### Prasyarat
Pastikan `swag` CLI telah terinstal pada *environment*:
```bash
# Verifikasi instalasi
swag --version

# Instalasi jika belum tersedia
go install github.com/swaggo/swag/cmd/swag@latest
```

### Perintah Regenerasi Swagger
Eksekusi perintah berikut dari direktori root `services/gateway`:

```bash
cd services/gateway
swag init -g cmd/main.go -o docs
```

Perintah di atas akan membaca seluruh berkas handler dan DTO, lalu memperbarui berkas keluaran berikut:
* `services/gateway/docs/docs.go`: Berkas kode Go berisi registrasi deklarasi Swagger ke runtime.
* `services/gateway/docs/swagger.json`: Spesifikasi OpenAPI dalam format JSON.
* `services/gateway/docs/swagger.yaml`: Spesifikasi OpenAPI dalam format YAML.

### Verifikasi Hasil Kompilasi
Lakukan uji kompilasi pada service gateway untuk memastikan tidak ada kesalahan sintaks atau rekursi tipe data yang merusak:
```bash
go build ./...
```

---

## 7. Panduan Pemecahan Masalah (*Troubleshooting*)

### 1. Pesan Peringatan `recursion detected`
* **Gejala:** Muncul *warning* `Skipping 'dto.AdministrativeDivisionSimpleResponse', recursion detected.` saat `swag init`.
* **Penyebab:** Struct DTO memiliki referensi diri (*self-referencing struct pointer*, seperti parent-child tree).
* **Solusi:** `swag` secara cerdas mengabaikan kedalaman rekursi tak terhingga untuk mencegah pembengkakan JSON spec. Ini merupakan perilaku normal dan tidak memengaruhi fungsionalitas Swagger UI.

### 2. Kesalahan `ParseComment error in file ...`
* **Gejala:** `swag init` gagal dengan pesan sintaks tag tidak valid.
* **Penyebab:** Terdapat spasi yang tidak sesuai, tipe data yang tidak terdefinisi di paket DTO, atau format tag salah.
* **Solusi:** Periksa baris anotasi yang dilaporkan dan pastikan tipe DTO diekspor (*public uppercase*) serta diimpor dengan benar.

### 3. Header Authorization Tidak Dikirim di Swagger UI
* **Solusi:** Klik tombol **Authorize** di kanan atas Swagger UI, masukkan nilai:
  ```
  Bearer <JWT_TOKEN_ANDA>
  ```
  Lalu klik **Authorize** dan jalankan permintaan (*Try it out*).
