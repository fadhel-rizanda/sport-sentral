# Nambah service baru
```
    cd services/gateway
    go mod init microservice-golang/services/gateaway
    cd ../..
    go work sync
    go work use -r .
```

# Contoh penggunaan GORM
```go
// Base Query dengan Context
query := r.db.WithContext(ctx)

// 1. Filter Dasar (AND, OR, NOT)
query.Where("type = ?", t).Where("active = ?", true) // AND (Implisit)
query.Or("is_priority = ?", true)                    // OR
query.Not("status = ?", "deleted")                   // NOT

// 2. Filter Advanced (LIKE, IN, BETWEEN)
query.Where("name LIKE ?", "%"+search+"%")           // Search
query.Where("id IN ?", []string{"uuid1", "uuid2"})   // Multiple ID
query.Where("created_at BETWEEN ? AND ?", start, end)// Range Waktu

// 3. Relasi & Seleksi Kolom
query.Preload("Permissions")                         // Eager Loading (Separate Query)
query.Joins("JOIN profiles ON profiles.user_id = users.id") // Manual Join
query.Select("id", "name", "slug")                   // Pilih kolom tertentu
query.Omit("password", "internal_note")              // Kecualikan kolom tertentu

// 4. Pengurutan & Pagination
query.Order("created_at DESC")                       // Urutan terbaru
query.Order("name ASC")                              // Urutan abjad
query.Limit(10)                                      // Batasi jumlah data
query.Offset(20)                                     // Skip data (untuk page 3)

// 5. Eksekusi Data (Read)
err := query.Find(&items).Error                      // Ambil banyak (List)
err := query.First(&item).Error                      // Ambil satu (By ID/Filter)
err := query.Take(&item).Error                       // Ambil satu (Tanpa default order)

// 6. Eksekusi Data (Write & Action)
err := r.db.Create(&item).Error                      // Simpan data baru
err := r.db.Save(&item).Error                        // Update semua field (Upsert)
err := r.db.Model(&item).Update("name", "new").Error // Update satu kolom
err := r.db.Model(&item).Updates(mapData).Error      // Update banyak kolom (Map/Struct)
err := r.db.Delete(&item).Error                      // Hapus data (Soft delete jika ada DeletedAt)

// 7. Agregasi & Utility
var count int64
query.Count(&count)                                  // Hitung total record
query.Pluck("id", &ids)                              // Ambil satu kolom saja ke Slice
```

buat bikin migration
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate create -ext sql -dir db/migrations -seq create_statuses_table