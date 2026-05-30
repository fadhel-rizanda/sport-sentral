# gRPC dan Protocol Buffers (Proto) Guide

Dokumen ini menjelaskan integrasi **gRPC** dan **Protocol Buffers (Proto)** sebagai backbone komunikasi sinkron antarlayanan (HTTP/2) di dalam ekosistem microservices ini.

---

## 1. Pendahuluan Protocol Buffers (Proto)

**Protocol Buffers** adalah bahasa pendefinisian data (Interface Definition Language - IDL) besutan Google yang bersifat *language-agnostic* (tidak bergantung pada bahasa pemrograman tertentu). Anda cukup menuliskan satu kontrak skema dalam format `.proto`, lalu generator (`buf` atau `protoc`) akan memproduksi kode (struct, validation logic, client, dan server interfaces) ke berbagai bahasa sasaran seperti Go, Java, TypeScript, dan lain-lain.

### Kenapa gRPC dan Proto dibanding REST dan JSON?

| Fitur | REST dan JSON | gRPC dan Proto (HTTP/2) |
|---|---|---|
| **Format Data** | Teks (JSON String, membutuhkan bandwidth lebih besar) | Biner terkompresi (Sangat ringkas dan cepat) |
| **Type Safety** | Lemah (Raw string/map, rentan terhadap kesalahan ketik dan bug runtime) | Sangat Kuat (Type-safe, diverifikasi pada saat compile-time) |
| **Protokol** | HTTP/1.1 (Satu request per koneksi TCP) | HTTP/2 (Multiplexing, bidirectional streaming) |
| **Kontrak API** | Opsional (Dokumentasi OpenAPI/Swagger sering kali tidak selaras) | Wajib (Berkas `.proto` adalah single source of truth) |

---

## 2. Anatomi Berkas `.proto`

Berikut adalah contoh skema di berkas proto kita (contoh: `proto/user/v1/user.proto`):

```protobuf
syntax = "proto3";  // Versi sintaks proto yang digunakan

package user.v1;    // Namespace untuk mencegah bentrokan nama (naming conflict)

// Opsi penamaan package Go hasil generate
option go_package = "microservice-golang/gen/user/v1;userv1";

import "buf/validate/validate.proto"; // Validasi deklaratif

// Message: Definisi struktur data (seperti Go Struct)
message User {
  string id       = 1;  // Angka 1, 2, 3 adalah Field Number (bukan nilai data)
  string email    = 2;  // Digunakan untuk encoding biner
  string username = 3;  // Tidak boleh diubah setelah masuk ke lingkungan produksi
}

// Service: Definisi RPC endpoints
service UserService {
  rpc CreateUser (CreateUserRequest) returns (CreateUserResponse);
  rpc GetUser (GetUserRequest) returns (GetUserResponse);
}
```

> [!IMPORTANT]
> **Pentingnya Field Number**:
> Di dalam pesan protobuf, angka setelah tanda sama dengan (`= 1`, `= 2`) mewakili **Field Number**. Ketika data diserialisasi ke format biner, nomor ini digunakan untuk mengidentifikasi field, bukan nama variabelnya.
> - **Aturan Utama**: Setelah sistem berjalan di lingkungan produksi, dilarang mengubah, menukar, atau menghapus Field Number yang sudah ada demi menjaga *backward dan forward compatibility*.

---

## 3. Proses Code Generation (Menggunakan `buf`)

Dibanding menggunakan perintah `protoc` manual yang kompleks, proyek ini menggunakan **`buf` CLI** (modern protobuf toolchain) untuk manajemen skema.

Konfigurasi `buf.gen.yaml` mengatur ke mana berkas hasil generate diletakkan:
```yaml
version: v1
plugins:
  - plugin: buf.build/protocolbuffers/go
    out: gen
    opt: paths=source_relative
  - plugin: buf.build/grpc/go
    out: gen
    opt: paths=source_relative
```

### Berkas yang Dihasilkan
Menjalankan perintah `buf generate` pada root directory akan memproduksi dua berkas utama di dalam direktori `gen/`:

#### A. `[nama].pb.go`
Berisi representasi tipe data pesan protobuf sebagai Go struct. Berkas ini digunakan di seluruh layer kode (handler, usecase, repository) sebagai media transfer data.
```go
type User struct {
    Id       string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
    Email    string `protobuf:"bytes,2,opt,name=email,proto3" json:"email,omitempty"`
    Username string `protobuf:"bytes,3,opt,name=username,proto3" json:"username,omitempty"`
}
```

#### B. `[nama]_grpc.pb.go`
Mengatur fungsionalitas transport gRPC. Berkas ini memproduksi dua interface krusial:
1. **`UserServiceServer`**: Interface wajib yang harus diimplementasikan oleh Handler penyedia layanan (sisi server).
2. **`UserServiceClient`**: Client interface yang digunakan oleh layanan lain untuk memanggil layanan ini (sisi client).

```
[user.proto]
     │
     ▼ (buf generate)
┌─────────────────────────────────────────────────────────────┐
│                            gen/                             │
│  ├── user.pb.go        ➔ Go structs (Tipe data payload)      │
│  │                                                          │
│  └── user_grpc.pb.go   ➔ Server Interface (untuk Handler)   │
│                        ➔ Client Interface (untuk Pemanggil) │
└─────────────────────────────────────────────────────────────┐
```

---

## 4. Implementasi Sisi Server (gRPC Server)

### A. Handler dan Unimplemented Server Embedding

Agar handler dianggap valid sebagai server gRPC, handler tersebut harus mengimplementasikan interface `UserServiceServer` dari berkas `_grpc.pb.go`.

```go
type UserHandler struct {
    // Wajib di-embed untuk forward compatibility
    userv1.UnimplementedUserServiceServer 
    userUC usecase.UserUseCase
}
```

> [!TIP]
> **Tujuan embedding `Unimplemented[Service]Server`**:
> gRPC mewajibkan embedding ini karena interface server memiliki method privat terenkapsulasi.
> Jika di masa mendatang Anda menambahkan method RPC baru di berkas `.proto` namun handler Go Anda belum mengimplementasikannya, aplikasi **tidak akan mengalami kegagalan kompilasi**. Method baru tersebut secara otomatis akan diarahkan ke method bawaan `Unimplemented` yang mengembalikan kode status `codes.Unimplemented`.

### B. Registrasi Server

Di berkas `main.go`, kami membuat gRPC server dan mendaftarkan handler ke dalamnya:

```go
// 1. Inisialisasi gRPC Server dengan Interceptor Tracing & Logger
grpcServer := grpc.NewServer(
    grpc.StatsHandler(otelgrpc.NewServerHandler()), // Tracing otomatis
    grpc.ChainUnaryInterceptor(
        interceptor.UnaryLogger(log),   // Logging request
        interceptor.UnaryRecovery(log), // Penangan panic crash
    ),
)

// 2. Registrasi Handler
userHandler := handler.NewUserHandler(userUC)
userv1.RegisterUserServiceServer(grpcServer, userHandler)

// 3. Listen dan Serve
listener, err := net.Listen("tcp", ":50051")
if err != nil {
    log.Fatal("failed to listen", zap.Error(err))
}
go func() {
    if err := grpcServer.Serve(listener); err != nil {
        log.Fatal("failed to serve gRPC", zap.Error(err))
    }
}()
```

---

## 5. Implementasi Sisi Klien (gRPC Client)

Layanan lain (seperti API Gateway atau Auth-Service) memanggil gRPC Server menggunakan Client interface hasil generate.

```go
// 1. Inisialisasi koneksi TCP ke port server gRPC dengan Tracing & Insecure Creds
conn, err := grpc.NewClient(
    "localhost:50051",
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithStatsHandler(otelgrpc.NewClientHandler()), // Tracing context propagation
)
if err != nil {
    log.Fatal("failed to connect to user-service", zap.Error(err))
}
defer conn.Close()

// 2. Membuat instance client menggunakan connection pool di atas
client := userv1.NewUserServiceClient(conn)

// 3. Eksekusi RPC seperti memanggil fungsi lokal
resp, err := client.GetUser(ctx, &userv1.GetUserRequest{Id: "user-uuid-123"})
if err != nil {
    st, ok := status.FromError(err)
    if ok {
         // Mengambil info error gRPC terstruktur
         log.Error("gRPC Error", zap.String("code", st.Code().String()), zap.String("msg", st.Message()))
    }
    return err
}
fmt.Println("User Email:", resp.User.Email)
```

---

## 6. Validasi Otomatis dengan `buf.validate`

Validasi input pada gRPC tidak perlu ditulis secara manual di tingkat handler. Kita dapat menggunakan plugin `buf.validate` untuk mendeklarasikan validasi langsung di skema proto.

### A. Mendefinisikan Validasi di Proto
```protobuf
message CreateUserRequest {
  string email    = 1 [(buf.validate.field).string.email = true]; // Validasi format email
  string password = 2 [(buf.validate.field).string.min_len = 8];  // Minimal 8 karakter
}
```

### B. Validasi Otomatis via Interceptor
Dengan pustaka `protovalidate-go`, pengecekan validasi ditempatkan pada layer interceptor global:

```go
func UnaryValidationInterceptor() grpc.UnaryServerInterceptor {
    validator, _ := protovalidate.New()
    
    return func(
        ctx context.Context,
        req any,
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (any, error) {
        // Validasi otomatis request yang masuk
        if err := validator.Validate(req.(proto.Message)); err != nil {
            return nil, status.Error(codes.InvalidArgument, err.Error())
        }
        return handler(ctx, req)
    }
}
```
Semua request yang tidak memenuhi spesifikasi skema proto akan ditolak secara otomatis dengan status `INVALID_ARGUMENT`, sehingga tidak membebani performa usecase bisnis Anda.
