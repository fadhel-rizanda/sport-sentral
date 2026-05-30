# Event-Driven Architecture dengan NATS JetStream

Catatan teknis ini menjelaskan arsitektur asynchronous messaging yang digunakan dalam sistem microservices ini menggunakan **NATS JetStream**. NATS JetStream berfungsi sebagai message broker berbasis event (Event Bus) yang menjamin pengiriman pesan antar-layanan secara handal, toleran terhadap kegagalan, dan tersinkronisasi.

---

## 1. Pendahuluan NATS JetStream

Dibandingkan dengan komunikasi sinkron gRPC (HTTP/2), komunikasi asinkron via event-driven memiliki keunggulan:
- **Loose Coupling**: Layanan penerbit (*Publisher*) tidak perlu mengetahui siapa penerima (*Subscriber*) atau apakah penerima sedang aktif atau tidak.
- **High Availability**: Jika salah satu layanan tidak aktif (down), event akan disimpan dengan aman di NATS JetStream dan akan diproses secara otomatis saat layanan tersebut aktif kembali.
- **Toleransi Kegagalan (Reliability)**: Dilengkapi dengan mekanisme **Explicit Acknowledgment (ACK)** dan **Retry/Redelivery Policy** untuk memastikan tidak ada data yang hilang dalam proses transmisi.

---

## 2. Arsitektur Aliran Event (Event Stream)

Sistem mengadopsi komunikasi dua arah yang terbagi menjadi dua stream utama:

```
                            [ IDENTITY_EVENTS Stream ]
                         Subjects: identity.user.*
    ┌──────────────────┐                                  ┌──────────────────┐
    │ IDENTITY-SERVICE │ ───────────────────────────────> │   META-SERVICE   │
    │   (Publisher)    │                                  │   (Subscriber)   │
    └──────────────────┘                                  └──────────────────┘
                                                            durable: meta-service-durable

                            [ META_EVENTS Stream ]
                         Subjects: meta.status.*
    ┌──────────────────┐                                  ┌──────────────────┐
    │ IDENTITY-SERVICE │ <─────────────────────────────── │   META-SERVICE   │
    │   (Subscriber)   │                                  │   (Publisher)   │
    └──────────────────┘                                  └──────────────────┘
      durable: identity-service-durable
```

### Detail Stream dan Subject

1. **`IDENTITY_EVENTS` Stream**
   - **Nama Stream**: `IDENTITY_EVENTS`
   - **Subjects**: `identity.user.created`, `identity.user.updated`, `identity.user.deleted`
   - **Subscriber**: `meta-service` (Durable Name: `meta-service-durable`)
   - **Tujuan**: Sinkronisasi data user dari `identity-service` ke database lokal `meta-service`.

2. **`META_EVENTS` Stream**
   - **Nama Stream**: `META_EVENTS`
   - **Subjects**: `meta.status.created`, `meta.status.updated`, `meta.status.deleted`
   - **Subscriber**: `identity-service` (Durable Name: `identity-service-durable`)
   - **Tujuan**: Sinkronisasi status/metadata ke `identity-service`.

---

## 3. Pembungkus Client (`shared/pkg/messaging`)

Untuk mempermudah penggunaan NATS JetStream di seluruh microservice, kami membungkus SDK NATS Go dalam sebuah abstraction layer di berkas `shared/pkg/messaging/messaging.go`.

### A. Inisialisasi dan Koneksi

Fungsi `Connect` bertanggung jawab melakukan handshake ke server NATS, menginisialisasi modul JetStream, serta membuat atau memperbarui stream secara otomatis jika belum terdefinisi.

```go
func Connect(cfg Config, logger *zap.Logger) (*Client, error) {
    // 1. Opsi Koneksi NATS (Max Reconnects & Backoff)
    opts := []nats.Option{
        nats.MaxReconnects(cfg.MaxReconnects),
        nats.ReconnectWait(cfg.ReconnectWait),
        nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
            slog.Warn("NATS disconnected", "error", err)
        }),
        nats.ReconnectHandler(func(nc *nats.Conn) {
            slog.Info("NATS reconnected", "url", nc.ConnectedUrl())
        }),
    }

    nc, err := nats.Connect(cfg.URL, opts...)
    if err != nil {
        return nil, fmt.Errorf("nats connect: %w", err)
    }

    // 2. Inisialisasi JetStream Context
    js, err := jetstream.New(nc)
    if err != nil {
        nc.Drain()
        return nil, apperr.Internal(fmt.Errorf("jetstream context: %w", err))
    }

    // 3. Auto-Provisioning Stream dengan File Storage
    _, err = js.CreateOrUpdateStream(context.Background(), jetstream.StreamConfig{
        Name:      cfg.StreamName,
        Subjects:  cfg.StreamSubjects,
        MaxAge:    cfg.RetentionMaxAge,
        Retention: jetstream.LimitsPolicy,
        Storage:   jetstream.FileStorage, // Data disimpan di disk (persisten)
        Replicas:  1,
    })
    if err != nil {
        nc.Drain()
        return nil, fmt.Errorf("create/update stream: %w", err)
    }

    slog.Info("NATS JetStream ready", "stream", cfg.StreamName)
    return &Client{nc: nc, js: js, cfg: cfg, logger: logger}, nil
}
```

### B. Mekanisme Penerbitan Event (`Publish`)

Penerbitan event dilakukan secara asinkron dengan fitur **Exponential Backoff Retry** bawaan untuk menangani kegagalan jaringan sementara. Event ditransmisikan dalam format **Protobuf Biner** yang efisien dan type-safe.

```go
func (c *Client) Publish(ctx context.Context, subject string, msg proto.Message) error {
    // Serialisasi data Protobuf ke format biner
    data, err := proto.Marshal(msg)
    if err != nil {
        return fmt.Errorf("marshal payload: %w", err)
    }

    var lastErr error
    delay := c.cfg.PublishBaseDelay

    // Loop retry dengan backoff
    for attempt := 1; attempt <= c.cfg.PublishMaxAttempts; attempt++ {
        var ack *jetstream.PubAck
        ack, lastErr = c.js.Publish(ctx, subject, data)
        if lastErr == nil {
            slog.Debug("published", "subject", subject, "seq", ack.Sequence)
            return nil
        }

        slog.Warn("nats publish failed, retrying",
            "subject", subject,
            "attempt", attempt,
            "max", c.cfg.PublishMaxAttempts,
            "error", lastErr,
        )

        if attempt < c.cfg.PublishMaxAttempts {
            select {
            case <-ctx.Done():
                return fmt.Errorf("publish cancelled after %d attempt(s): %w", attempt, ctx.Err())
            case <-time.After(delay):
                delay *= 2 // Menggandakan durasi tunda (100ms -> 200ms -> 400ms -> ...)
            }
        }
    }

    return fmt.Errorf("publish to %s failed after %d attempts: %w", subject, c.cfg.PublishMaxAttempts, lastErr)
}
```

### C. Mekanisme Berlangganan Event (`Subscribe`)

Penerima (*Subscriber*) memanfaatkan konsep **Durable Consumer** dengan kebijakan pengakuan eksplisit (**Explicit ACK**) untuk menjamin pengantaran setidaknya sekali (*at-least-once delivery*).

```go
func (c *Client) Subscribe(
    ctx context.Context,
    streamName, durableName string,
    filterSubjects []string,
    handler func(ctx context.Context, subject string, data []byte) error,
) error {
    // 1. Membuat atau memperbarui Consumer persisten (Durable)
    cons, err := c.js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
        Durable:        durableName,
        FilterSubjects: filterSubjects,
        AckPolicy:      jetstream.AckExplicitPolicy, // Wajib konfirmasi penerimaan
        AckWait:        c.cfg.DefaultAckWait,        // Waktu tunggu ACK sebelum pengiriman ulang (redelivery)
        MaxDeliver:     c.cfg.DefaultMaxDeliver,     // Maksimal percobaan pengiriman ulang
        DeliverPolicy:  jetstream.DeliverAllPolicy,  // Mengirim semua event dari awal stream
        MaxAckPending:  c.cfg.DefaultMaxAckPending,
    })
    if err != nil {
        return fmt.Errorf("create consumer: %w", err)
    }

    // 2. Mengonsumsi data secara asinkron
    cc, err := cons.Consume(func(msg jetstream.Msg) {
        if err := handler(ctx, msg.Subject(), msg.Data()); err != nil {
            slog.Error("handler error – nak", "subject", msg.Subject(), "error", err)
            _ = msg.Nak() // Negatively Acknowledge -> Kirim ulang event
            return
        }
        _ = msg.Ack() // Sukses -> Tandai event telah selesai diproses
    })
    if err != nil {
        return fmt.Errorf("start consume: %w", err)
    }
    defer cc.Stop()

    <-ctx.Done()
    return nil
}
```

---

## 4. Pola Implementasi di Layanan

### A. Pengiriman Event (Publisher)

Sebagai contoh, ketika terjadi mutasi user di `identity-service`, layanan akan mengirimkan payload melalui `UserEventPublisher` pada berkas `internal/delivery/nats/user_publisher.go`:

```go
type UserEventPublisher struct {
    nats   *messaging.Client
    logger *zap.Logger
}

func (p *UserEventPublisher) PublishUserCreated(ctx context.Context, evt *userv1.UserEvent) error {
    return p.nats.Publish(ctx, events.SubjectUserCreated, evt)
}
```

Pemicuan event di level Usecase:
```go
// identity-service/internal/usecase/user.go
evt := &userv1.UserEvent{
    EventId:   uuid.New().String(),
    Type:      "CREATED",
    Timestamp: timestamppb.Now(),
    User:      protoUser,
}
_ = uc.eventPublisher.PublishUserCreated(ctx, evt)
```

### B. Penerimaan Event (Subscriber/Consumer)

Di sisi `meta-service`, event yang diterima akan ditangani secara asinkron melalui `UserSubscriber` pada berkas `internal/delivery/nats/user_subscriber.go`:

```go
type UserSubscriber struct {
    syncUC      *usecase.UserSyncUseCase
    nats        *messaging.Client
    logger      *zap.Logger
    durableName string
}

func (s *UserSubscriber) handleMessage(ctx context.Context, subject string, data []byte) error {
    var evt userv1.UserEvent
    if err := proto.Unmarshal(data, &evt); err != nil {
        s.logger.Warn("invalid protobuf – skipping", zap.String("subject", subject), zap.Error(err))
        return nil // Abaikan jika format data rusak (mencegah loop tak terbatas)
    }
    return s.syncUC.SyncUser(ctx, &evt) // Proses sinkronisasi ke DB
}

func (s *UserSubscriber) Listen(ctx context.Context) error {
    return s.nats.Subscribe(
        ctx,
        events.IdentityStreamName,
        events.MetaDurableName,
        []string{
            events.SubjectUserCreated,
            events.SubjectUserUpdated,
            events.SubjectUserDeleted,
        },
        s.handleMessage,
    )
}
```

---

## 5. Strategi Penanganan Kesalahan dan Toleransi Kegagalan

1. **Protobuf Validation Failure**:
   - Jika payload gagal didekod (misalnya tipe protobuf tidak cocok), subscriber akan melakukan pencatatan log (warning) dan langsung mengembalikan nilai `nil` (sukses). Langkah ini diambil untuk menghindari terjadinya **Dead-Letter Loops**, di mana pesan usang atau rusak terus dikirim ulang (`Nak`) tanpa akhir.

2. **Database Temporary Error (e.g. Connection Timeout)**:
   - Jika usecase mengembalikan error karena kendala koneksi database sementara, subscriber akan memberikan respon `msg.Nak()`.
   - NATS JetStream akan otomatis menjadwalkan ulang pengiriman pesan tersebut berdasarkan konfigurasi `DefaultAckWait` hingga batas maksimal percobaan `MaxDeliver` terpenuhi.

3. **Graceful Shutdown**:
   - Saat aplikasi menerima sinyal penghentian (`SIGINT`/`SIGTERM`), method `Drain()` dipanggil pada koneksi NATS:
     ```go
     defer natsClient.Drain()
     ```
   - Hal ini memastikan consumer berhenti menerima pesan baru, menyelesaikan pemrosesan event yang sedang berjalan, dan menutup koneksi secara aman ke server tanpa memutuskan transaksi yang sedang aktif.

---

## 6. Konfigurasi Lingkungan (Environment Variables)

Berikut adalah parameter konfigurasi yang wajib diatur pada berkas `.env` setiap layanan:

```env
# NATS Connection Settings
IDENTITY_NATS_HOST=nats://localhost
IDENTITY_NATS_PORT=4222

# NATS Client & Stream Configuration (Muat di internal/config/config.go)
NATS_MAX_RECONNECTS=5
NATS_RECONNECT_WAIT=2s
NATS_PUBLISH_MAX_ATTEMPTS=3
NATS_PUBLISH_BASE_DELAY=100ms
NATS_DEFAULT_ACK_WAIT=30s
NATS_DEFAULT_MAX_DELIVER=3
NATS_DEFAULT_MAX_ACK_PENDING=1000
```
