package telemetry

import (
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"

	"gorm.io/gorm"
)

func InitGORMTracing(db *gorm.DB, serviceName string) error {
	return db.Use(otelgorm.NewPlugin(
		otelgorm.WithDBName(serviceName),
	))
}
