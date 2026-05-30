package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	JaegerEndpoint string
	Enabled        bool
}

type Telemetry struct {
	tracerProvider *sdktrace.TracerProvider
	tracer         trace.Tracer
	meterProvider  *sdkmetric.MeterProvider
	meter          metric.Meter
	logger         *zap.Logger
}

func New(cfg Config, logger *zap.Logger) (*Telemetry, error) {
	if !cfg.Enabled {
		logger.Warn("telemetry disabled")
		return &Telemetry{
			tracerProvider: sdktrace.NewTracerProvider(),
			tracer:         otel.Tracer(cfg.ServiceName),
			meterProvider:  sdkmetric.NewMeterProvider(),
			meter:          otel.Meter(cfg.ServiceName),
			logger:         logger,
		}, nil
	}

	ctx := context.Background()

	// 1. Setup Resource (bisa dipakai bareng buat Tracing & Metrics)
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// ==========================================
	// TRACING SETUP (Jaeger / OTLP)
	// ==========================================
	conn, err := grpc.NewClient(
		cfg.JaegerEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	// Propagator standar buat distributed tracing
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	// ==========================================
	// METRICS SETUP (Prometheus)
	// ==========================================
	promExporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create prometheus exporter: %w", err)
	}

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(promExporter),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	logger.Info("telemetry initialized",
		zap.String("service", cfg.ServiceName),
		zap.String("jaeger_endpoint", cfg.JaegerEndpoint),
		zap.String("metrics", "prometheus"),
	)

	return &Telemetry{
		tracerProvider: tp,
		tracer:         otel.Tracer(cfg.ServiceName),
		meterProvider:  mp,
		meter:          otel.Meter(cfg.ServiceName),
		logger:         logger,
	}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	var errs []error

	if t.tracerProvider != nil {
		if err := t.tracerProvider.Shutdown(ctx); err != nil {
			t.logger.Error("failed to shutdown tracer provider", zap.Error(err))
			errs = append(errs, err)
		}
	}

	if t.meterProvider != nil {
		if err := t.meterProvider.Shutdown(ctx); err != nil {
			t.logger.Error("failed to shutdown meter provider", zap.Error(err))
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("encountered %d errors during telemetry shutdown", len(errs))
	}

	t.logger.Info("telemetry shutdown successfully")
	return nil
}

func (t *Telemetry) Tracer() trace.Tracer {
	return t.tracer
}

func (t *Telemetry) Meter() metric.Meter {
	return t.meter
}
