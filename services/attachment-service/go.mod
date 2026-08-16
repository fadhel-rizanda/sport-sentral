module microservice-golang/services/attachment-service

go 1.25.0

require (
	buf.build/go/protovalidate v1.2.0
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/minio/minio-go/v7 v7.0.86
	github.com/prometheus/client_golang v1.23.2
	github.com/redis/go-redis/v9 v9.19.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.69.0
	go.uber.org/zap v1.28.0
	google.golang.org/grpc v1.81.1
	google.golang.org/protobuf v1.36.11
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.31.1
)
