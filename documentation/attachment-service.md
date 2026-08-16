# Spesifikasi Teknis dan Dokumentasi Attachment Service

Dokumen ini menyajikan spesifikasi teknis, arsitektur sistem, desain antarmuka pemrograman aplikasi (API), abstraksi penyimpanan objek, serta mekanisme integrasi berbasis peristiwa (*event-driven*) untuk **Attachment Service** (`services/attachment-service`) dalam ekosistem *microservices* **Sport-Sentral**.

---

## 1. Ikhtisar Layanan (*Service Overview*)

**Attachment Service** merupakan komponen infrastruktur inti yang bertanggung jawab atas pengelolaan siklus hidup berkas biner dan aset multimedia (gambar, dokumen, avatar, logo, sertifikat, dan video). Layanan ini memisahkan logika penyimpanan berkas dari domain bisnis lain (`academy-service`, `venue-service`, `scout-service`, dan sebagainya) melalui antarmuka terpadu yang menyediakan kapabilitas:

* **Penerbitan URL Unggah Bertanda Tangan (*Presigned Upload URL*)**: Memfasilitasi klien untuk mengunggah berkas secara langsung ke penyimpanan objek (*Object Storage*) tanpa membebani *bandwidth* API Gateway atau server gRPC.
* **Pengunggahan Biner Langsung (*Direct Binary Upload*)**: Menyediakan saluran pengunggahan berkas secara langsung melalui *multipart/form-data* pada API Gateway atau aliran biner gRPC.
* **Verifikasi dan Konfirmasi Unggahan**: Memvalidasi integritas dan keberadaan fisik berkas pada media penyimpanan sebelum mengubah status metadata menjadi aktif (`ACTIVE`).
* **Manajemen Metadata dan Caching**: Menyimpan catatan struktural berkas di PostgreSQL serta menerapkan strategi *cache-aside* menggunakan Redis untuk menjamin latensi baca yang rendah.
* **Publikasi Peristiwa Asinkron (*Event-Driven Architecture*)**: Menerbitkan peristiwa status berkas (`attachment.created`, `attachment.updated`, `attachment.deleted`) ke NATS JetStream.

---

## 2. Arsitektur Komponen dan Aliran Data

```
                               ┌───────────────────────────┐
                               │       Aplikasi Klien      │
                               │   (Web / Mobile / Admin)  │
                               └─────────────┬─────────────┘
                                             │ HTTP/REST (Port 8080)
                                             ▼
                               ┌───────────────────────────┐
                               │        API Gateway        │
                               └─────────────┬─────────────┘
                                             │ gRPC Unary (Port 50060)
                                             ▼
                               ┌───────────────────────────┐
                               │    Attachment Service     │
                               └──────┬──────┬──────┬──────┘
                                      │      │      │
           ┌──────────────────────────┘      │      └──────────────────────────┐
           │ PostgreSQL                      │ Redis (Database 8)              │ Storage Engine
           ▼                                 ▼                                 ▼
┌────────────────────┐            ┌────────────────────┐            ┌────────────────────┐
│   Tabel Basis Data │            │   Cache Metadata   │            │   Penyedia Media   │
│   `attachments`    │            │ attachment:id:{id} │            │  Local / MinIO / S3│
└────────────────────┘            └────────────────────┘            └────────────────────┘
                                             │
                                             │ NATS JetStream (Subject: attachment.>)
                                             ▼
                               ┌───────────────────────────┐
                               │   Stream JetStream        │
                               │  `ATTACHMENT_EVENTS`      │
                               └───────────────────────────┘
```

---

## 3. Abstraksi Penyimpanan (*Storage Provider Abstraction*)

Attachment Service menerapkan pola desain *Strategy* melalui antarmuka `StorageProvider` untuk mengisolasi logika dependensi *storage driver*.

### Antarmuka `StorageProvider`

```go
type StorageProvider interface {
	Save(ctx context.Context, bucket string, objectKey string, reader io.Reader, size int64, contentType string) (string, error)
	GetPresignedUploadURL(ctx context.Context, bucket string, objectKey string, expiry time.Duration) (string, map[string]string, error)
	GetPresignedDownloadURL(ctx context.Context, bucket string, objectKey string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, bucket string, objectKey string) error
	Exists(ctx context.Context, bucket string, objectKey string) (bool, error)
}
```

### Implementasi Penyimpanan

| Penyedia (*Provider*) | Konfigurasi Environment | Deskripsi Penggunaan |
| :--- | :--- | :--- |
| **Local Storage** | `STORAGE_PROVIDER=local` | Menyimpan berkas pada sistem berkas lokal (`./public/uploads`). Digunakan untuk lingkungan pengujian dan pengembangan lokal. |
| **MinIO / AWS S3** | `STORAGE_PROVIDER=minio` | Menggunakan pustaka resmi MinIO yang sepenuhnya kompatibel dengan protokol AWS S3 v4 Signature. Cocok untuk lingkungan *staging* dan *production*. |

### Mekanisme Pengalihan Otomatis (*Auto-Fallback*)

Apabila konfigurasi diatur ke `STORAGE_PROVIDER=minio` namun server MinIO/S3 tidak dapat dijangkau pada saat inisialisasi awal, sistem secara otomatis mencatat peringatan (*warning log*) dan mengalihkan penyimpanan ke **Local Storage Provider** guna menjamin ketersediaan layanan (*high availability*).

---

## 4. Alur Kerja Pengunggahan Berkas (*Upload Workflows*)

### A. Alur Pengunggahan Langsung (*Direct Upload Flow*)

Digunakan ketika klien mengirimkan berkas langsung ke backend dalam satu siklus permintaan.

```
[ Klien ]               [ API Gateway ]          [ Attachment Service ]       [ Storage Engine ]
    │                          │                           │                          │
    │── POST /upload (Form) ──>│                           │                          │
    │   (multipart/form-data)  │── UploadAttachment (gRPC)─>│                          │
    │                          │   (Byte Stream & Meta)    │── Save (Write File) ────>│
    │                          │                           │<── URL Unduhan Berkas ───│
    │                          │                           │── Simpan Metadata ke DB ─│
    │                          │                           │── Terbitkan Event NATS ──│
    │                          │<── Attachment DTO ────────│                          │
    │<── 201 Created (JSON) ───│                           │                          │
```

### B. Alur Pengunggahan Bertanda Tangan (*Presigned URL 3-Step Flow*)

Direkomendasikan untuk berkas berukuran besar guna menghindari pemborosan *resource memory* dan *bandwidth* pada API Gateway.

```
[ Klien ]               [ API Gateway ]          [ Attachment Service ]       [ Storage Engine ]
    │                          │                           │                          │
    │ 1. Minta Presigned URL   │                           │                          │
    │── POST /presigned-url ──>│── CreatePresignedURL ────>│                          │
    │                          │                           │── Buat Record PENDING ───│
    │                          │                           │── Generate Upload URL ──>│
    │<── Upload URL & ID ──────│<── Response Presigned ────│                          │
    │                          │                           │                          │
    │ 2. Upload Biner Langsung ke Storage Engine                                      │
    │───────────────────────── PUT {upload_url} (Binary Payload) ────────────────────>│
    │<──────────────────────── 200 OK ────────────────────────────────────────────────│
    │                          │                           │                          │
    │ 3. Konfirmasi Status     │                           │                          │
    │── POST /confirm ────────>│── ConfirmUpload (gRPC) ──>│                          │
    │   {"attachment_id": id}  │                           │── Exists(bucket, key)? ─>│
    │                          │                           │<── Status Keberadaan ────│
    │                          │                           │── Update Status ACTIVE ──│
    │                          │                           │── Terbitkan Event NATS ──│
    │<── 200 OK (Status ACTIVE)│<── Attachment DTO ────────│                          │
```

---

## 5. Definisi Protobuf dan Antarmuka gRPC

Kontrak antarmuka gRPC didefinisikan pada direktori `proto/attachment/v1/`:

```protobuf
syntax = "proto3";

package attachment.v1;

service AttachmentService {
  rpc CreatePresignedUploadUrl(CreatePresignedUploadUrlRequest) returns (CreatePresignedUploadUrlResponse);
  rpc ConfirmUpload(ConfirmUploadRequest) returns (ConfirmUploadResponse);
  rpc UploadAttachment(UploadAttachmentRequest) returns (UploadAttachmentResponse);
  rpc GetAttachment(GetAttachmentRequest) returns (GetAttachmentResponse);
  rpc GetAttachments(GetAttachmentsRequest) returns (GetAttachmentsResponse);
  rpc DeleteAttachment(DeleteAttachmentRequest) returns (DeleteAttachmentResponse);
  rpc ListAttachments(ListAttachmentsRequest) returns (ListAttachmentsResponse);
}
```

---

## 6. Integrasi Berbasis Peristiwa (*NATS JetStream*)

Attachment Service menerbitkan data perubahan status ke JetStream Stream **`ATTACHMENT_EVENTS`**.

### Daftar Topik (*Subjects*) dan Definisi Peristiwa

| Nama Peristiwa | Subjek Peristiwa | Deskripsi Pemicu |
| :--- | :--- | :--- |
| `AttachmentCreatedEvent` | `attachment.created` | Diterbitkan saat presigned URL dibuat atau pengunggahan langsung berhasil. |
| `AttachmentUpdatedEvent` | `attachment.updated` | Diterbitkan saat status berkas terkonfirmasi (`ACTIVE` atau `FAILED`). |
| `AttachmentDeletedEvent` | `attachment.deleted` | Diterbitkan saat berkas fisik dan metadata dihapus dari sistem. |

---

## 7. Spesifikasi REST API Gateway

Seluruh endpoint REST dipublikasikan melalui API Gateway dengan prefix `/api/v1/attachments`.

### A. Pengunggahan Langsung (*Direct Upload*)

* **Endpoint:** `POST /api/v1/attachments/upload`
* **Otentikasi:** Wajib (*Bearer Token*)
* **Content-Type:** `multipart/form-data`
* **Parameter Form:**
  * `file` (*Binary, Wajib*): Berkas yang akan diunggah.
  * `file_category` (*String, Opsional*): Kategori berkas (`GENERAL`, `IMAGE`, `AVATAR`, `LOGO`, `DOCUMENT`, `CERTIFICATE`, `VIDEO`).

**Format Respons (`201 Created`):**
```json
{
  "success": true,
  "data": {
    "id": "9110b0df-1729-4754-86a9-581c0142d545",
    "filename": "react_certif.png",
    "file_path": "uploads/2026/08/9110b0df-1729-4754-86a9-581c0142d545/react_certif.png",
    "bucket_name": "attachments",
    "file_size": 293531,
    "mime_type": "image/png",
    "file_category": "CERTIFICATE",
    "uploaded_by_user_id": "019ff8f2-09b6-70aa-ab4f-d83a431d03be",
    "status": "ACTIVE",
    "url": "http://localhost:8080/public/uploads/2026/08/9110b0df-1729-4754-86a9-581c0142d545/react_certif.png",
    "created_at": "2026-08-16T14:35:03Z",
    "updated_at": "2026-08-16T14:35:03Z"
  }
}
```

---

### B. Pembuatan Presigned Upload URL

* **Endpoint:** `POST /api/v1/attachments/presigned-url`
* **Otentikasi:** Wajib (*Bearer Token*)
* **Content-Type:** `application/json`

**Skema Permintaan:**
```json
{
  "filename": "document_legalitas.pdf",
  "mime_type": "application/pdf",
  "file_size": 1500000,
  "file_category": "DOCUMENT"
}
```

**Format Respons (`200 OK`):**
```json
{
  "success": true,
  "data": {
    "attachment_id": "e72e1baf-5201-4d49-b2ef-e8f03f7c6387",
    "upload_url": "http://localhost:8080/public/uploads/2026/08/e72e1baf-5201-4d49-b2ef-e8f03f7c6387/document_legalitas.pdf",
    "file_path": "uploads/2026/08/e72e1baf-5201-4d49-b2ef-e8f03f7c6387/document_legalitas.pdf"
  }
}
```

---

### C. Konfirmasi Status Pengunggahan

* **Endpoint:** `POST /api/v1/attachments/confirm`
* **Otentikasi:** Wajib (*Bearer Token*)
* **Content-Type:** `application/json`

**Skema Permintaan:**
```json
{
  "attachment_id": "e72e1baf-5201-4d49-b2ef-e8f03f7c6387",
  "is_success": true
}
```

**Format Respons (`200 OK`):**
```json
{
  "success": true,
  "data": {
    "id": "e72e1baf-5201-4d49-b2ef-e8f03f7c6387",
    "filename": "document_legalitas.pdf",
    "file_path": "uploads/2026/08/e72e1baf-5201-4d49-b2ef-e8f03f7c6387/document_legalitas.pdf",
    "bucket_name": "attachments",
    "file_size": 1500000,
    "mime_type": "application/pdf",
    "file_category": "DOCUMENT",
    "uploaded_by_user_id": "019ff8f2-09b6-70aa-ab4f-d83a431d03be",
    "status": "ACTIVE",
    "url": "http://localhost:8080/public/uploads/2026/08/e72e1baf-5201-4d49-b2ef-e8f03f7c6387/document_legalitas.pdf",
    "created_at": "2026-08-16T14:35:03Z",
    "updated_at": "2026-08-16T14:36:12Z"
  }
}
```

---

### D. Pengambilan Metadata Berkas Tunggal

* **Endpoint:** `GET /api/v1/attachments/:id`
* **Otentikasi:** Wajib (*Bearer Token*)

**Format Respons (`200 OK`):**
```json
{
  "success": true,
  "data": {
    "id": "e72e1baf-5201-4d49-b2ef-e8f03f7c6387",
    "filename": "document_legalitas.pdf",
    "file_path": "uploads/2026/08/e72e1baf-5201-4d49-b2ef-e8f03f7c6387/document_legalitas.pdf",
    "bucket_name": "attachments",
    "file_size": 1500000,
    "mime_type": "application/pdf",
    "file_category": "DOCUMENT",
    "uploaded_by_user_id": "019ff8f2-09b6-70aa-ab4f-d83a431d03be",
    "status": "ACTIVE",
    "url": "http://localhost:8080/public/uploads/2026/08/e72e1baf-5201-4d49-b2ef-e8f03f7c6387/document_legalitas.pdf",
    "created_at": "2026-08-16T14:35:03Z",
    "updated_at": "2026-08-16T14:36:12Z"
  }
}
```

---

### E. Daftar Berkas dengan Paginasi dan Filter

* **Endpoint:** `GET /api/v1/attachments?page=1&limit=10&file_category=CERTIFICATE`
* **Otentikasi:** Wajib (*Bearer Token*)

**Format Respons (`200 OK`):**
```json
{
  "success": true,
  "data": [
    {
      "id": "e72e1baf-5201-4d49-b2ef-e8f03f7c6387",
      "filename": "react_certif.png",
      "file_path": "uploads/2026/08/e72e1baf-5201-4d49-b2ef-e8f03f7c6387/react_certif.png",
      "bucket_name": "attachments",
      "file_size": 293531,
      "mime_type": "image/png",
      "file_category": "CERTIFICATE",
      "uploaded_by_user_id": "019ff8f2-09b6-70aa-ab4f-d83a431d03be",
      "status": "ACTIVE",
      "url": "http://localhost:8080/public/uploads/2026/08/e72e1baf-5201-4d49-b2ef-e8f03f7c6387/react_certif.png",
      "created_at": "2026-08-16T14:35:03Z",
      "updated_at": "2026-08-16T14:36:12Z"
    }
  ],
  "meta": {
    "page": 1,
    "page_size": 10,
    "total": 1
  }
}
```

---

### F. Penghapusan Berkas

* **Endpoint:** `DELETE /api/v1/attachments/:id`
* **Otentikasi:** Wajib (*Bearer Token*)

**Format Respons (`200 OK`):**
```json
{
  "success": true,
  "message": "Attachment deleted successfully"
}
```

---

## 8. Skema Basis Data dan Mekanisme Caching

### Skema Tabel Relasional (`PostgreSQL`)

```sql
CREATE TABLE attachments (
    id UUID PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    file_path VARCHAR(512) NOT NULL,
    bucket_name VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_category VARCHAR(50) NOT NULL,
    uploaded_by_user_id UUID,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_attachments_status ON attachments (status);
CREATE INDEX idx_attachments_file_category ON attachments (file_category);
CREATE INDEX idx_attachments_uploaded_by_user_id ON attachments (uploaded_by_user_id);
CREATE INDEX idx_attachments_created_at_desc ON attachments (created_at DESC);
```

### Strategi Caching Redis

* **Pola Kunci (*Key Pattern*)**: `attachment:id:{uuid}`
* **Masa Berlaku (*TTL*)**: 15 Menit.
* **Invalidasi Otomatis**: Kunci *cache* otomatis dihapus saat pembaruan status (`Update`) atau penghapusan berkas (`Delete`).

---

## 9. Konfigurasi Variabel Lingkungan (*Environment Variables*)

```env
# Konfigurasi Aplikasi Umum
APP_ENV=development
APP_NAME=sport-sentral
APP_VERSION=0.0.1
SERVICE_NAME=attachment-service
SERVICE_VERSION=0.0.1

# Konfigurasi Port Jaringan
GRPC_PORT=50060
METRIC_PORT=9010

# Konfigurasi Basis Data PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=attachment_db
DB_SSLMODE=disable

# Konfigurasi Cache Redis (Database Khusus Attachment)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=8

# Konfigurasi Penyimpanan (Local vs Object Storage)
STORAGE_PROVIDER=local
STORAGE_BASE_URL=http://localhost:8080/public
STORAGE_LOCAL_DIR=./public

# Konfigurasi MinIO / AWS S3
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false
MINIO_BUCKET=attachments

# Konfigurasi Observabilitas & Telemetri
TELEMETRY_ENABLED=true
JAEGER_HOST=localhost
JAEGER_PORT=4317
```

---

## 10. Prosedur Pengujian dan Verifikasi Mutu

Jalankan pengujian unit terisolasi untuk modul `attachment-service`:

```bash
go test ./services/attachment-service/... -v
```

Jalankan seluruh pengujian unit lintas layanan mikro:

```bash
.\test.bat
```
