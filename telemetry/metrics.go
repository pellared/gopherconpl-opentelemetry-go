package telemetry

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/global"
	export "go.opentelemetry.io/otel/sdk/export/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// ShutdownMetrics is a delegate that closes the metrics.
type ShutdownMetrics func(ctx context.Context) error

// SetupMetrics sets the global OpenTelemetry MeterProvider configured to use
// the Prometheus exporter and will exposes it on port 2222.
func SetupMetrics(service string) (ShutdownMetrics, error) {
	config := prometheus.Config{}
	c := controller.New(
		processor.New(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(config.DefaultHistogramBoundaries),
			),
			export.CumulativeExportKindSelector(),
			processor.WithMemory(true),
		),
		// Record information about this application in an Resource.
		controller.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
		)),
	)
	exporter, err := prometheus.New(config, c)
	if err != nil {
		return func(ctx context.Context) error { return nil }, err
	}

	global.SetMeterProvider(exporter.MeterProvider())

	srv := &http.Server{Addr: ":2222", Handler: exporter}
	go func() {
		_ = srv.ListenAndServe()
	}()

	return srv.Shutdown, nil
}
