# Panduan Unit Testing di Project Microservice Golang

Dokumen ini menjelaskan **cara kerja** dan **tahapan langkah demi langkah** untuk membuat unit test di project microservice Golang ini. Unit test berfokus pada pengujian logika bisnis pada layer **UseCase** secara terisolasi tanpa menyentuh database atau network eksternal.

---

## 1. Konsep & Cara Kerja Unit Test

Unit test di proyek ini menggunakan beberapa pola standar Go untuk memastikan pengujian berjalan cepat, independen, dan andal:

### A. Black-box Testing (Package `_test`)
Semua berkas pengujian menggunakan nama package dengan suffix `_test` (contoh: `package usecase_test` di dalam folder `usecase`). Hal ini memastikan pengujian hanya mengakses fungsi, struct, dan method publik yang diekspos oleh package utama, menyimulasikan bagaimana modul lain memanggil kode tersebut.

### B. Mocking Menggunakan Implicit Interface
Di Go, struct mengimplementasikan interface secara implisit jika memiliki method-method dengan tanda tangan (*signature*) yang sama. Kita memanfaatkan ini untuk membuat struct mock (tiruan) untuk menggantikan repositori asli:
1. Kita buat struct mock (misal: `MockRoleRepository`) yang memiliki property bertipe fungsi (contoh: `GetByIDFunc`).
2. Kita implementasikan method interface asli pada struct mock tersebut dengan mengarahkan eksekusinya ke property fungsi kita.
3. Di setiap skenario test case, kita bebas mendefinisikan isi dari property fungsi tersebut secara dinamis untuk mengembalikan data sukses atau mensimulasikan error tertentu.

### C. Simulasi Context & Metadata gRPC
Banyak use case membutuhkan informasi role atau user ID dari context (gRPC metadata). Kita membuat helper function khusus di dalam file test untuk menyuntikkan metadata tiruan ke dalam `context.Context` sebelum memanggil fungsi yang diuji.

---

## 2. Skenario Pengujian Umum
Setiap pengujian UseCase biasanya dibagi ke dalam beberapa skenario menggunakan sub-test (`t.Run`):
1. **Happy Path (Success):** Input valid, repositori mock mengembalikan data sukses, event berhasil dipublikasikan (jika ada), dan tidak ada error yang dihasilkan.
2. **Unauthorized / Missing Metadata:** Memastikan UseCase menolak permintaan jika konteks tidak menyertakan metadata otentikasi.
3. **Forbidden (Hak Akses Kurang):** Memastikan UseCase mengembalikan error *Forbidden* jika pengguna yang mencoba tidak memiliki hak akses/role yang tepat.
4. **Not Found:** Memastikan UseCase mengembalikan error *NotFound* (biasanya memetakan `gorm.ErrRecordNotFound` dari db) jika data tidak ditemukan.
5. **Conflict (Duplikasi Data):** Memastikan UseCase menangkap error constraint dari database (seperti kode Postgres `23505`) dan memetakannya menjadi error domain *Conflict*.

---

## 3. Tahapan Membuat Unit Test (Step-by-Step)

Berikut adalah tahapan praktis untuk membuat unit test baru untuk sebuah UseCase:

### Langkah 1: Buat Berkas Test Baru
Buat file baru di direktori yang sama dengan kode yang ingin diuji dengan akhiran `_test.go`.
* *Contoh lokasi:* `services/identity-service/internal/usecase/role_test.go`
* *Package:* Gunakan suffix `_test` (misal: `package usecase_test`).

### Langkah 2: Definisikan Struct Mock untuk Dependency
Tentukan interface apa saja yang digunakan oleh UseCase Anda (misal: Repository atau Event Publisher), lalu buat struct mock tiruan di bagian atas file test.

```go
type MockRoleRepository struct {
    // Definisikan field fungsi agar bisa disetel secara dinamis di dalam test
    GetByIDFunc func(ctx context.Context, id uuid.UUID) (*entity.Role, error)
    CreateFunc  func(ctx context.Context, role *entity.Role) error
}
```

### Langkah 3: Implementasikan Method Interface pada Struct Mock
Tulis method-method agar struct mock tersebut secara legal mengimplementasikan interface aslinya.

```go
func (m *MockRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Role, error) {
    return m.GetByIDFunc(ctx, id)
}

func (m *MockRoleRepository) Create(ctx context.Context, role *entity.Role) error {
    return m.CreateFunc(ctx, role)
}
```

> [!NOTE]
> Jika interface asli memiliki method lain yang tidak Anda uji, Anda tetap harus menuliskan method kosongnya di file mock ini agar Go Compiler tidak menganggap struct tersebut tidak mengimplementasikan interface secara utuh.

### Langkah 4: Buat Helper Context (Opsional)
Jika logika bisnis Anda membutuhkan informasi metadata (misalnya role aktif), buat fungsi pembantu untuk mempermudah pembuatan Context:

```go
func withActiveRole(role string) context.Context {
    md := metadata.New(map[string]string{
        "active-role": role,
    })
    return metadata.NewIncomingContext(context.Background(), md)
}
```

### Langkah 5: Tulis Fungsi Test Utama & Skenario `t.Run`
Tulis fungsi dengan nama diawali `Test...` dan gunakan `t.Run` untuk memisahkan skenario pengujian.

```go
func TestRoleUseCase_Create(t *testing.T) {
    // 1. Siapkan data request tiruan
    req := dto.CreateRoleRequest{Name: "Admin", Slug: "admin"}

    t.Run("success_as_platform_admin", func(t *testing.T) {
        // 2. Instansiasi Mock
        repo := &MockRoleRepository{}
        uc := usecase.NewRoleUseCase(repo)
        ctx := withActiveRole("platform_admin")

        // 3. Konfigurasikan respons Mock khusus untuk skenario ini
        repo.CreateFunc = func(ctx context.Context, r *entity.Role) error {
            r.ID = uuid.New() // simulasikan auto-generate ID di DB
            return nil
        }

        // 4. Jalankan fungsi yang diuji
        res, err := uc.Create(ctx, req)

        // 5. Lakukan Assertion (pengecekan hasil)
        if err != nil {
            t.Fatalf("expected no error, got %v", err)
        }
        if res.Name != "Admin" {
            t.Errorf("expected name to be Admin, got %s", res.Name)
        }
    })

    t.Run("forbidden_role", func(t *testing.T) {
        repo := &MockRoleRepository{}
        uc := usecase.NewRoleUseCase(repo)
        ctx := withActiveRole("member") // Role biasa tidak boleh membuat role baru

        _, err := uc.Create(ctx, req)

        if err == nil {
            t.Fatal("expected error, got nil")
        }
        if !apperr.IsForbidden(err) {
            t.Errorf("expected forbidden error, got %v", err)
        }
    })
}
```

### Langkah 6: Jalankan Pengujian
Buka terminal Anda dan jalankan perintah test dari root workspace:
```powershell
go test ./services/nama-service/...

// atau untuk semua sekaligus
.\test.bat
```

---

## 4. Contoh Nyata di Codebase

Untuk melihat penerapan nyata dari pola unit test ini, silakan pelajari file-file berikut:
* **Role UseCase Test:** [role_test.go](file:///C:/MAGANG_mandiri/microservice-golang/services/identity-service/internal/usecase/role_test.go)
* **Academy Holding UseCase Test:** [academy_holding_test.go](file:///C:/MAGANG_mandiri/microservice-golang/services/academy-service/internal/usecase/academy_holding_test.go)
* **Sport UseCase Test:** [sport_test.go](file:///C:/MAGANG_mandiri/microservice-golang/services/sport-service/internal/usecase/sport_test.go)
