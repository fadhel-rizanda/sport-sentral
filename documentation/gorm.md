**GORM — Ringkasan Lengkap**

---

**Apa itu GORM?**

ORM (Object Relational Mapper) untuk Go. Tugasnya menjembatani antara Go struct dan tabel database — lu tidak perlu tulis raw SQL untuk operasi umum.

```go
// Tanpa GORM
rows, err := db.QueryContext(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", id, email)

// Dengan GORM
db.WithContext(ctx).Create(&user)
```

---

**Setup & Koneksi**

```go
import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

dsn := "host=localhost user=postgres password=secret dbname=mydb port=5432"
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
```

---

**Model Convention**

GORM punya konvensi default yang bisa di-override dengan tag:

```go
type User struct {
    ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`      // custom type
    Email     string         `gorm:"uniqueIndex;not null"`      // unique + not null
    Username  string         `gorm:"column:user_name"`          // custom column name
    FullName  string         `gorm:"not null;default:''"`       // default value
    RoleID    *uuid.UUID     `gorm:"type:uuid"`                 // nullable FK
    CreatedAt time.Time                                         // auto-managed GORM
    UpdatedAt time.Time                                         // auto-managed GORM
    DeletedAt gorm.DeletedAt `gorm:"index"`                    // soft delete
}
```

Konvensi default:
- Nama tabel = plural snake_case dari struct name → `User` jadi `users`
- Primary key = field `ID`
- `CreatedAt`, `UpdatedAt`, `DeletedAt` di-manage otomatis

Override nama tabel:
```go
func (User) TableName() string { return "app_users" }
```

---

**AutoMigrate**

Generate/update tabel dari struct secara otomatis:

```go
db.AutoMigrate(&User{}, &Role{}, &Permission{})
```

Penting: AutoMigrate hanya **tambah** kolom baru, tidak pernah hapus kolom yang sudah ada. Untuk production sebaiknya pakai migration tool seperti `golang-migrate`.

---

**CRUD Dasar**

**Create:**
```go
user := &User{Email: "test@example.com"}
result := db.Create(user)
// Setelah Create, user.ID sudah ter-set oleh GORM
// result.Error — error jika ada
// result.RowsAffected — jumlah row yang dibuat
```

**Read:**
```go
// First — ambil satu, order by primary key, error jika tidak ada
var user User
db.First(&user, "id = ?", id)

// Find — ambil banyak, tidak error jika kosong
var users []User
db.Find(&users)

// Where
db.Where("email = ? AND deleted_at IS NULL", email).First(&user)
```

**Update:**
```go
// Updates dengan map — hanya update field yang ada di map
db.Model(&user).Updates(map[string]any{
    "full_name": "John",
    "username":  "john",
})

// Save — update semua field (hati-hati, bisa overwrite data)
db.Save(&user)
```

**Delete:**
```go
// Soft delete — set deleted_at = NOW() jika struct punya gorm.DeletedAt
db.Delete(&user)

// Hard delete — bypass soft delete
db.Unscoped().Delete(&user)
```

---

**Soft Delete**

Kalau struct punya field `DeletedAt gorm.DeletedAt`, GORM otomatis:
- `Delete()` → set `deleted_at = NOW()`, tidak hapus row
- Semua query (`Find`, `First`, dll) otomatis tambah `WHERE deleted_at IS NULL`
- Untuk query include soft-deleted: pakai `db.Unscoped()`

```go
// Hanya return user yang tidak soft-deleted
db.Find(&users)

// Return semua termasuk soft-deleted
db.Unscoped().Find(&users)
```

---

**Hooks**

Function yang dipanggil otomatis sebelum/sesudah operasi DB:

```go
// BeforeCreate — dipanggil sebelum INSERT
func (u *User) BeforeCreate(_ *gorm.DB) error {
    if u.ID == uuid.Nil {
        u.ID = uuid.New() // generate UUID sebelum insert
    }
    return nil
}

// Hook lain yang tersedia:
// BeforeSave, AfterSave
// BeforeCreate, AfterCreate
// BeforeUpdate, AfterUpdate
// BeforeDelete, AfterDelete
// AfterFind
```

---

**Associations**

**BelongsTo** — User belongs to Role:
```go
type User struct {
    RoleID *uuid.UUID
    Role   *Role      `gorm:"foreignKey:RoleID"`
}
```

**HasMany** — Role has many Users:
```go
type Role struct {
    Users []User `gorm:"foreignKey:RoleID"`
}
```

**Many2Many** — Role ↔ Permission:
```go
type Role struct {
    Permissions []Permission `gorm:"many2many:role_permissions;"`
}
```
GORM otomatis buat join table `role_permissions` dengan kolom `role_id` dan `permission_id`.

**Preload** — load association saat query:
```go
// Load Role beserta semua Permissions-nya
db.Preload("Permissions").Where("id = ?", id).First(&role)

// Load nested
db.Preload("Role.Permissions").First(&user)
```

**Association operations:**
```go
// Tambah permission ke role (insert ke join table)
db.Model(&role).Association("Permissions").Append(&permission)

// Hapus permission dari role (delete dari join table)
db.Model(&role).Association("Permissions").Delete(&permission)

// Replace semua permissions
db.Model(&role).Association("Permissions").Replace(&newPermissions)
```

---

**Transactions**

```go
err := db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&user).Error; err != nil {
        return err // auto rollback
    }
    if err := tx.Create(&profile).Error; err != nil {
        return err // auto rollback
    }
    return nil // auto commit
})
```

Manual transaction:
```go
tx := db.Begin()
if err := tx.Create(&user).Error; err != nil {
    tx.Rollback()
    return err
}
tx.Commit()
```

---

**Scopes**

Reusable query conditions:
```go
func ActiveUsers(db *gorm.DB) *gorm.DB {
    return db.Where("is_active = ?", true)
}

func Paginate(page, pageSize int) func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Limit(pageSize).Offset((page - 1) * pageSize)
    }
}

// Pemakaian
db.Scopes(ActiveUsers, Paginate(1, 20)).Find(&users)
```

---

**Raw SQL**

Kalau query terlalu kompleks untuk GORM API:
```go
// Raw query
db.Raw("SELECT * FROM users WHERE email = ?", email).Scan(&user)

// Exec untuk non-select
db.Exec("UPDATE users SET role_id = NULL WHERE role_id = ?", roleID)
```

---

**Hal yang Perlu Diwaspadai**

**1. N+1 query** — paling umum terjadi saat load association tanpa Preload:
```go
// BAD — query N+1: 1 query untuk users, N query untuk setiap role
for _, user := range users {
    fmt.Println(user.Role.Name) // query per user
}

// GOOD — 2 query saja
db.Preload("Role").Find(&users)
```

**2. `Updates` vs `Save`:**
```go
// Updates — hanya update field yang di-pass, aman
db.Model(&user).Updates(map[string]any{"full_name": "John"})

// Save — update SEMUA field, field kosong akan overwrite data existing
db.Save(&user) // hati-hati
```

**3. Zero value diabaikan oleh `Updates`:**
```go
// Ini TIDAK akan update is_active ke false karena false adalah zero value
db.Model(&user).Updates(User{IsActive: false})

// Gunakan map untuk update zero value
db.Model(&user).Updates(map[string]any{"is_active": false})
```

**4. `First` vs `Find`:**
```go
// First — error ErrRecordNotFound jika tidak ada
db.First(&user, "id = ?", id) // harus handle not found

// Find — tidak error jika kosong, slice tetap kosong
db.Find(&users) // aman untuk list query
```

---

**Singkatnya:**

```
GORM = struct tag sebagai schema
     + hooks untuk lifecycle
     + associations untuk relasi
     + soft delete otomatis via DeletedAt
     + transaction support
     + raw SQL kalau butuh
```