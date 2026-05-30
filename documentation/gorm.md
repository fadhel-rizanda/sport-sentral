# GORM — Panduan dan Praktik Terbaik

Dokumen ini menyajikan ringkasan komprehensif penggunaan **GORM** (Object Relational Mapper) untuk Go, mulai dari koneksi, manajemen skema, relasi asosiasi, hingga daftar kendala yang wajib dihindari.

---

## 1. Pendahuluan GORM

GORM adalah ORM untuk bahasa pemrograman Go. Tugas utamanya adalah memetakan (mapping) Go Struct ke tabel database relasional. Dengan GORM, Anda dapat meminimalisir penulisan query SQL mentah (raw SQL) untuk operasi dasar CRUD.

```go
// Tanpa GORM (Raw SQL)
rows, err := db.QueryContext(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", id, email)

// Dengan GORM (Type-Safe dan Ringkas)
db.WithContext(ctx).Create(&user)
```

---

## 2. Inisialisasi Koneksi dan Konfigurasi

Inisialisasi database PostgreSQL menggunakan driver resmi GORM:

```go
import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

dsn := "host=localhost user=postgres password=secret dbname=mydb port=5432 sslmode=disable"
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
```

---

## 3. Konvensi Model dan Struktur Struct

GORM memiliki konvensi bawaan yang kuat. Namun, Anda selalu dapat menyesuaikan perilaku tersebut menggunakan tag struct.

```go
type User struct {
    ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`      // Primary key UUID manual
    Email     string         `gorm:"uniqueIndex;not null"`      // Unique index & non-nullable
    Username  string         `gorm:"column:user_name"`          // Nama kolom kustom di DB
    FullName  string         `gorm:"not null;default:''"`       // Nilai default
    RoleID    *uuid.UUID     `gorm:"type:uuid"`                 // Nullable Foreign Key
    CreatedAt time.Time                                         // Dikelola otomatis oleh GORM
    UpdatedAt time.Time                                         // Dikelola otomatis oleh GORM
    DeletedAt gorm.DeletedAt `gorm:"index"`                     // Mengaktifkan fitur Soft Delete
}
```

### Konvensi Default GORM:
1. **Nama Tabel**: Secara otomatis diubah menjadi jamak (plural) dan berformat `snake_case` (contoh: struct `User` dipetakan ke tabel `users`).
2. **Primary Key**: Menggunakan kolom bernama `id`.
3. **Auto-timestamp**: Kolom `created_at` dan `updated_at` diperbarui otomatis oleh framework pada operasi insert dan update.

Jika ingin menggunakan nama tabel kustom:
```go
func (User) TableName() string {
    return "app_users" // Meng-override konvensi tabel
}
```

---

## 4. Migrasi Otomatis (AutoMigrate)

GORM menyediakan fitur skema sinkronisasi instan melalui fungsi `AutoMigrate`.

```go
db.AutoMigrate(&User{}, &Role{}, &Permission{})
```

> [!WARNING]
> **Penting untuk Lingkungan Produksi**:
> `AutoMigrate` hanya akan **menambahkan** kolom baru, indeks baru, atau tabel baru. Fitur ini **tidak akan menghapus atau mengubah tipe data** kolom yang sudah ada untuk menghindari kehilangan data secara tidak sengaja.
> - **Rekomendasi**: Untuk lingkungan produksi, gunakan alat migrasi terpisah yang terstruktur seperti `golang-migrate` agar mutasi skema tercatat dalam versi file `.sql` yang terkontrol.

---

## 5. Operasi CRUD Dasar

### A. Create (Insert)
```go
user := &User{Email: "john@example.com", Username: "john_doe"}
result := db.Create(user)

// result.Error        -> Menampung error jika query gagal
// result.RowsAffected -> Menghitung jumlah record yang berhasil disimpan
// Setelah Create berhasil, user.ID akan terisi secara otomatis.
```

### B. Read (Querying)
```go
// 1. First: Mengambil 1 baris pertama berdasarkan primary key, mengembalikan error `gorm.ErrRecordNotFound` jika tidak ditemukan.
var user User
err := db.First(&user, "id = ?", id).Error

// 2. Find: Mengambil banyak baris, tidak akan mengembalikan error jika data kosong (mengembalikan slice kosong).
var users []User
err := db.Find(&users).Error

// 3. Where: Menyaring data secara terstruktur
db.Where("email = ? AND deleted_at IS NULL", email).First(&user)
```

### C. Update (Mutasi)
```go
// Updates (Menggunakan Map): Mengubah kolom spesifik secara aman
db.Model(&user).Updates(map[string]any{
    "full_name": "John Doe Updated",
    "username":  "john_updated",
})

// Save: Menyimpan seluruh field struct (Hati-hati! Field kosong pada struct akan menimpa data di database)
db.Save(&user)
```

### D. Delete (Penghapusan)
```go
// 1. Soft Delete: Jika model memiliki DeletedAt, data tidak dihapus secara fisik melainkan diperbarui timestamp penghapusannya.
db.Delete(&user)

// 2. Hard Delete: Menghapus data secara fisik dan permanen dari database.
db.Unscoped().Delete(&user)
```

---

## 6. Fitur Soft Delete

Jika model memiliki bidang `DeletedAt gorm.DeletedAt`, GORM akan menyesuaikan kueri penghapusan:
- Operasi `Delete()` akan mengeksekusi kueri `UPDATE` untuk mengisi kolom `deleted_at` dengan waktu saat ini.
- Semua kueri pencarian bawaan (`Find`, `First`, dll.) secara otomatis disisipi kondisi `WHERE deleted_at IS NULL`.
- Jika ingin mengambil data yang telah dihapus sementara, gunakan filter **`Unscoped()`**:

```go
// Hanya menampilkan data yang aktif (belum dihapus)
db.Find(&users)

// Menampilkan seluruh data, termasuk yang sudah dihapus sementara (soft delete)
db.Unscoped().Find(&users)
```

---

## 7. Penanganan Hook Lifecycle

Hook adalah fungsi yang dieksekusi secara otomatis oleh GORM sebelum atau sesudah operasi database tertentu dijalankan.

```go
// BeforeCreate: Dipanggil otomatis sebelum baris baru dimasukkan ke database (INSERT)
func (u *User) BeforeCreate(tx *gorm.DB) error {
    if u.ID == uuid.Nil {
        u.ID = uuid.New() // Membuat UUID baru jika kosong
    }
    return nil
}
```

Daftar hook yang tersedia:
- **Create**: `BeforeSave`, `BeforeCreate`, `AfterCreate`, `AfterSave`
- **Update**: `BeforeSave`, `BeforeUpdate`, `AfterUpdate`, `AfterSave`
- **Delete**: `BeforeDelete`, `AfterDelete`
- **Query**: `AfterFind`

---

## 8. Manajemen Relasi Asosiasi

### Belongs To
Model `User` memiliki satu `Role`:
```go
type User struct {
    RoleID *uuid.UUID
    Role   *Role `gorm:"foreignKey:RoleID"`
}
```

### Has Many
Satu `Role` dapat dimiliki oleh banyak `User`:
```go
type Role struct {
    Users []User `gorm:"foreignKey:RoleID"`
}
```

### Many to Many
Relasi banyak-ke-banyak (misalnya `Role` memiliki banyak `Permission`):
```go
type Role struct {
    Permissions []Permission `gorm:"many2many:role_permissions;"`
}
// GORM otomatis membuat join-table bernama `role_permissions`
```

### Eager Loading (Preload)
Secara default, GORM tidak memuat data relasi untuk efisiensi performa. Gunakan `Preload` untuk mengambil data relasi terkait:

```go
// Mengambil Role beserta seluruh data list Permissions sekaligus
db.Preload("Permissions").Where("id = ?", roleID).First(&role)

// Mengambil User beserta data Role dan nested Permission di dalam Role tersebut
db.Preload("Role.Permissions").First(&user)
```

---

## 9. Penggunaan Transaksi

Transaksi memastikan prinsip ACID. Jika salah satu operasi database gagal di dalam blok transaksi, seluruh rangkaian perubahan akan dibatalkan secara otomatis (*rollback*).

```go
err := db.Transaction(func(tx *gorm.DB) error {
    // Gunakan objek `tx` di dalam blok transaksi, bukan `db`
    if err := tx.Create(&user).Error; err != nil {
        return err // Mengembalikan error otomatis memicu ROLLBACK
    }
    
    if err := tx.Create(&profile).Error; err != nil {
        return err // Mengembalikan error otomatis memicu ROLLBACK
    }
    
    return nil // Mengembalikan nil memicu COMMIT
})
```

---

## 10. Kendala Umum yang Wajib Dihindari

### 1. Masalah N+1 Query
Terjadi saat Anda melakukan iterasi hasil kueri untuk mengambil relasi tanpa menggunakan eager loading `Preload()`.
```go
// Kurang Baik: Memicu N+1 Query (1 query untuk mengambil daftar user, N query tambahan untuk setiap Role)
for _, user := range users {
    fmt.Println(user.Role.Name) // Mengakses bidang relasi memicu kueri SQL terpisah
}

// Baik: Hanya mengeksekusi total 2 kueri SQL secara efisien
db.Preload("Role").Find(&users)
```

### 2. Nilai Nol (Zero Value) Diabaikan oleh `Updates()` via Struct
Saat melakukan pembaruan data menggunakan struct, GORM secara default **tidak akan mengubah** data di database jika nilai bidang adalah *zero value* (seperti `false`, `0`, `""`).
```go
// Salah: Tidak akan memperbarui is_active menjadi false
db.Model(&user).Updates(User{IsActive: false})

// Benar: Gunakan Map untuk memperbarui zero value
db.Model(&user).Updates(map[string]any{"is_active": false})
```

### 3. Tidak Menyertakan Go Context
GORM memerlukan context (`ctx`) untuk menyalurkan informasi Distributed Tracing (Span OTel) dan penanganan pembatalan (timeout).
```go
// Kurang Baik: Tracing terputus di tingkat database
db.First(&user)

// Baik: Query tercatat dengan terstruktur di Jaeger UI
db.WithContext(ctx).First(&user)
```