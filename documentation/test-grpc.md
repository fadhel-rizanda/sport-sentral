# Panduan Pengujian gRPC Services

Dokumen ini menjelaskan tata cara pengujian (*testing*) layanan-layanan gRPC di dalam arsitektur microservices ini menggunakan tool **`grpcurl`**.

---

## 1. Pendahuluan `grpcurl`

`grpcurl` adalah tool CLI (*Command Line Interface*) yang berfungsi mirip seperti `curl`, namun dirancang khusus untuk berinteraksi dengan server gRPC. Dengan `grpcurl`, Anda dapat mengirim payload request gRPC (dalam format JSON) dan menerima respons terstruktur secara langsung dari terminal.

### Cara Instalasi:
```powershell
# Menggunakan Go CLI
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Atau menggunakan Scoop di Windows
scoop install grpcurl
```

---

## 2. Pemetaan Port Layanan (Port Mapping)

Pastikan layanan gRPC tujuan Anda sudah aktif dan mendengarkan pada port yang sesuai:

| Layanan (Service) | Port gRPC | Namespace Proto | Deskripsi |
| :--- | :--- | :--- | :--- |
| **Identity Service** | `50051` | `user.v1`, `rbac.v1` | Mengurusi data pengguna, otorisasi, hak akses (RBAC). |
| **Auth Service** | `50052` | `auth.v1` | Mengurusi token JWT, session, otentikasi login. |

---

## 3. Lembar Ringkasan Perintah Uji (Test Commands Cheatsheet)

Semua perintah di bawah ini menyertakan flag `-plaintext` karena komunikasi gRPC lokal kita dijalankan tanpa enkripsi TLS (HTTP/2 Cleartext).

### A. Layanan RBAC (RBAC Service - Port `50051`)

#### 1. Membuat Izin Hak Akses Baru (Create Permission)
```powershell
grpcurl -plaintext -d '{
  "resource": "user",
  "action": "delete",
  "description": "Can delete user accounts"
}' localhost:50051 rbac.v1.RBACService/CreatePermission
```

#### 2. Membuat Peran Baru (Create Role)
```powershell
grpcurl -plaintext -d '{
  "name": "admin",
  "description": "System Administrator"
}' localhost:50051 rbac.v1.RBACService/CreateRole
```

---

### B. Layanan Pengguna (User Service - Port `50051`)

#### 1. Mendaftarkan Akun Baru (Create User)
```powershell
grpcurl -plaintext -d '{
  "email": "test@example.com",
  "username": "testuser",
  "full_name": "Test User",
  "password": "password123",
  "role_id": "e3bce012-6d45-47ba-afe2-89ff4c2ec5e2"
}' localhost:50051 user.v1.UserService/CreateUser
```
*(Catatan: Sesuaikan nilai `role_id` dengan UUID role yang berhasil didaftarkan sebelumnya di database).*

#### 2. Memverifikasi Akun (Verify Account via Token)
```powershell
grpcurl -plaintext -d '{
  "token": "token_dari_email_atau_log"
}' localhost:50051 user.v1.UserService/VerifyAccount
```

#### 3. Mengirim Email Lupa Password (Forgot Password Request)
```powershell
grpcurl -plaintext -d '{
  "email": "test@example.com"
}' localhost:50051 user.v1.UserService/ForgotPassword
```

#### 4. Reset Password Akun
```powershell
grpcurl -plaintext -d '{
  "token": "token_reset_dari_email",
  "password": "newpassword123"
}' localhost:50051 user.v1.UserService/ResetPassword
```

---

### C. Layanan Otentikasi (Auth Service - Port `50052`)

#### 1. Melakukan Login Pengguna (Authentication)
```powershell
grpcurl -plaintext -d '{
  "email": "test@example.com",
  "password": "password123"
}' localhost:50052 auth.v1.AuthService/Login
```

---

## 4. Tips Diagnostik Kueri `grpcurl`

1. **Melihat Daftar Service yang Tersedia**:
   Jika server gRPC mengaktifkan fitur **gRPC Reflection**, Anda dapat mengintip semua nama service terdaftar tanpa mengetahui proto-nya:
   ```powershell
   grpcurl -plaintext localhost:50051 list
   ```
2. **Melihat Daftar Method di Dalam Service**:
   ```powershell
   grpcurl -plaintext localhost:50051 describe user.v1.UserService
   ```
3. **Melihat Skema Request JSON secara Detail**:
   ```powershell
   grpcurl -plaintext localhost:50051 describe user.v1.CreateUserRequest
   ```
