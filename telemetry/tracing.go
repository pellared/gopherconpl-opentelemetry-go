package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// ShutdownTracing is a delegate that closes the tracing.
type ShutdownTracing func(ctx context.Context) error

// SetupTracing sets the global OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The
// TracerProvider will also use a Resource configured with the information
// about the application.
func SetupTracing(service, url string) (ShutdownTracing, error) {
	// Create the Jaeger exporter.
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return func(ctx context.Context) error { return nil }, err
	}

	// Create the TracerProvider.
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		// Caution! Always be sure use WithBatcher to batch in production.
		tracesdk.WithSyncer(exp),
		// Record information about this application in an Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
		)),
	)

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)

	// Register W3C Trace Context propagator as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Return the Shutdown function so that it can be used by the caller to
	// send all the spans before the application closes.
	return tp.Shutdown, nil
}

// AddErrorEvent records the error to the current span.
func AddErrorEvent(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
}
