**Proto (Protocol Buffers)**

---

**Apa itu Proto?**

Proto adalah bahasa untuk mendefinisikan struktur data dan service yang language-agnostic. Lu tulis satu file `.proto`, lalu di-generate ke bahasa apapun — Go, Java, Python, dll. Semua bahasa dapat struct dan interface yang sama persis.

Analoginya seperti kontrak. Semua pihak (service) harus ikut kontrak yang sama.

---

**Kenapa pakai Proto + gRPC?**

Kalau pakai REST + JSON:
```
user-service kirim JSON string → auth-service parse JSON → rawan typo, tidak ada type safety
```

Kalau pakai gRPC + Proto:
```
user-service kirim binary terstruktur → auth-service langsung dapat Go struct → type safe, lebih cepat
```

---

**Anatomi file `.proto`**

```protobuf
syntax = "proto3";  // versi proto yang dipakai

package user.v1;    // namespace, mencegah naming conflict antar proto

option go_package = "microservice-golang/gen/user/v1;userv1";
// ↑ ini memberitau buf/protoc: generated Go file taruh di mana dan package name-nya apa
```

**Message** — definisi struct:
```protobuf
message User {
  string id       = 1;  // angka = field number, bukan value
  string email    = 2;  // dipakai untuk encoding binary, JANGAN diubah setelah production
  string username = 3;
}
```

Field number itu penting — binary encoding pakai nomor ini, bukan nama field. Makanya kalau sudah production, field number tidak boleh diubah atau dihapus.

**Service** — definisi RPC endpoint:
```protobuf
service UserService {
  rpc CreateUser (CreateUserRequest) returns (CreateUserResponse);
  //  ↑ nama method  ↑ input             ↑ output
}
```

---

**Apa yang di-generate?**

Dari `user.proto` kita, `buf generate` menghasilkan dua file:

**`user.pb.go`** — semua message jadi Go struct. Ini yang lu pakai sebagai tipe data di seluruh service:
```go
type User struct {
    Id       string
    Email    string
    Username string
}

type CreateUserRequest struct {
    Email    string
    Username string
    Password string
}

type CreateUserResponse struct {
    User *User
}
// semua message di proto jadi struct di sini
```

Jadi kalau lu tulis `*userv1.User` atau `*userv1.CreateUserRequest` di handler, itu berasal dari file ini.

---

**`user_grpc.pb.go`** — ini yang mengurus komunikasi gRPC-nya (seperti inisialisasi endpoint). Isinya dua hal:

**1. `UserServiceServer` — interface untuk service yang menyediakan endpoint (user-service)**

```go
type UserServiceServer interface {
    CreateUser(context.Context, *CreateUserRequest) (*CreateUserResponse, error)
    GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error)
    // dll...
}
```

`UserHandler` lu harus implement interface ini. Ketika ada request gRPC masuk ke user-service, gRPC framework yang routing request ke method yang tepat berdasarkan interface ini.

Makanya lu perlu:
```go
func (h *UserHandler) RegisterGRPC(s *grpc.Server) {
    userv1.RegisterUserServiceServer(s, h) // daftarkan handler sebagai implementasi
}
```

**2. `UserServiceClient` — interface untuk service yang mau memanggil user-service (auth-service, gateway)**

```go
type UserServiceClient interface {
    CreateUser(context.Context, *CreateUserRequest) (*CreateUserResponse, error)
    GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error)
    // dll...
}
```

Nanti ketika auth-service butuh data user, auth-service tidak perlu tau apapun tentang implementasi user-service. Cukup pakai client ini:

```go
// Di auth-service
conn, _ := grpc.Dial("user-service:50051", grpc.WithInsecure())
client := userv1.NewUserServiceClient(conn)  // ← dari user_grpc.pb.go

// Panggil seperti function biasa, padahal ini network call ke user-service
user, err := client.GetUserByEmailInternal(ctx, &userv1.GetUserByEmailRequest{
    Email: "test@example.com",
})
```

---

**Analoginya:**

`user.pb.go` = definisi bentuk datanya (amplop surat)

`user_grpc.pb.go` = sistem pengirimannya — satu sisi yang terima surat (`Server`), satu sisi yang kirim surat (`Client`)

---

**Ringkasan relasi antar file:**

```
user.proto
    ↓ buf generate
    ├── user.pb.go       → struct *userv1.User, *userv1.CreateUserRequest, dll
    │                      dipakai semua layer (handler, client)
    │
    └── user_grpc.pb.go  → UserServiceServer  : diimplementasi UserHandler (user-service)
                           UserServiceClient  : dipakai auth-service/gateway untuk call user-service
```
---

**Kenapa harus embed `UnimplementedUserServiceServer`?**

```go
type UserHandler struct {
    userv1.UnimplementedUserServiceServer  // ← ini
    uc usecase.UserUseCase
}
```

`UserServiceServer` interface punya method `mustEmbedUnimplementedUserServiceServer()` yang private — lu tidak bisa implement sendiri dari luar package. Satu-satunya cara implement interface ini adalah embed `UnimplementedUserServiceServer`.

Fungsinya untuk forward compatibility: kalau nanti proto ditambah RPC baru, service yang belum implement method baru itu tidak langsung crash — dia fallback ke `Unimplemented` yang return `codes.Unimplemented`.

---

**Kenapa `buf.validate`?**

Tanpa validasi, lu harus validasi manual di handler:
```go
if req.Email == "" {
    return nil, apperr.InvalidArgument("email is required")
}
if len(req.Password) < 8 {
    return nil, apperr.InvalidArgument("password too short")
}
// dll...
```

Dengan `buf.validate`, validasi didefinisikan di proto:
```protobuf
message CreateUserRequest {
  string email = 1 [(buf.validate.field).string.email = true];
  string password = 4 [(buf.validate.field).string.min_len = 8];
}
```

Lalu di interceptor gRPC cukup satu baris:
```go
if err := protovalidate.Validate(req); err != nil {
    return nil, status.Error(codes.InvalidArgument, err.Error())
}
```

Semua request otomatis tervalidasi sebelum masuk ke handler.

---

**Flow lengkap dari proto sampai running:**

```
1. Tulis user.proto
        ↓
2. buf generate
        ↓
3. user.pb.go + user_grpc.pb.go ter-generate di gen/
        ↓
4. UserHandler implement UserServiceServer interface
        ↓
5. RegisterGRPC(s) → daftarkan handler ke gRPC server
        ↓
6. gRPC server listen di port tertentu
        ↓
7. Client (auth-service/gateway) pakai UserServiceClient untuk panggil RPC
```