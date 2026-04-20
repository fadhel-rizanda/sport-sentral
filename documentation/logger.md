# Logging

Using `go.uber.org/zap` as structured logger.

## Setup

Init singleton once in `main.go`:

```go
log := applogger.New(os.Getenv("APP_ENV")) // "production" | "development"
defer log.Sync()
```

Access anywhere without injection:

```go
logger.Get().Info("user created", zap.String("user_id", id))
```

`APP_ENV=production` → JSON output. Default → colorized development.
Override level via `LOG_LEVEL=debug|info|warn|error`.

---

## Log Levels

| Level   | Kapan                                        |
|---------|----------------------------------------------|
| `Debug` | Verbose internal — off di production          |
| `Info`  | Business event sukses (user created, dll)     |
| `Warn`  | Input invalid dari client (not found, unauth) |
| `Error` | Kegagalan server unexpected (DB error, panic) |
| `Fatal` | Startup gagal — otomatis `os.Exit(1)`         |

---

## Arsitektur

```
Interceptor  → log semua request: latency, gRPC code, error
Usecase      → log business event penting (create, delete, dll)
Handler      → tidak perlu — interceptor sudah cover
Repository   → tidak perlu — error cukup di-return
```

---

## Interceptors

Urutan wajib — Logger di luar, Recovery di dalam:

```go
grpc.ChainUnaryInterceptor(
    interceptor.UnaryLogger(log),   // outer
    interceptor.UnaryRecovery(log), // inner
)
```

**UnaryLogger** — log semua request + latency + gRPC code.

**UnaryRecovery** — catch panic, log stack trace, return `codes.Internal`.

---

## Usecase

Log hanya business event yang meaningful:

```go
// CREATE / DELETE → log
logger.Get().Info("user created",
    zap.String("user_id", user.ID.String()),
    zap.String("email", email),
)

// READ (GetUser, ListUsers, dll) → tidak perlu di-log
```

---

## Common Fields

```go
zap.Error(err)
zap.String("user_id", id)
zap.String("method", info.FullMethod)
zap.String("code", st.Code().String())
zap.Duration("duration", elapsed)
zap.ByteString("stack", debug.Stack()) // saat panic
```