package signal

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/trace"

	logsdk "go.opentelemetry.io/otel/sdk/log"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

const (
	version = "dev"
)

func runOpenTelemetry(args []string) {
	fs := flag.NewFlagSet("opentelemetry", flag.ExitOnError)
	grpc := fs.Bool("grpc", false, "Use OTLP gRPC")

	if err := fs.Parse(args[1:]); err != nil {
		fs.Usage()
		os.Exit(1)
	}

	if *grpc {
		fmt.Println("Using OTLP gRPC")
	} else {
		fmt.Println("Using OTLP HTTP")
	}

	addr := fs.Arg(0)
	if addr == "" {
		fmt.Println("An OTLP endpoint is required.")
		os.Exit(1)
	}

	ctx := context.Background()

	name := "volt"
	baseKV := []string{
		"env", "local",
		"tenant", "test",
	}

	// Logs
	{
		logger, err := newOTelLogger(ctx, *grpc, addr, name, version, baseKV...)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for i, msg := range messages {
			sendOTelLog(ctx, logger, time.Now(), "info", msg,
				"uuid", uuid.NewString(),
			)

			fmt.Printf("Sent log:  #%-2d  message=%s\n", i+1, msg)
			time.Sleep(randMs(500, 1000))
		}
	}

	// Metrics
	{
		meter, err := newOTelMeter(ctx, *grpc, addr, name, version, baseKV...)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		counter, err := meter.Int64Counter("demo_opentelemetry_requests_total")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for i := range 10 {
			incr := int64(i * 2)
			counter.Add(ctx, incr,
				metric.WithAttributes(attribute.String("req_method", "GET")),
				metric.WithAttributes(attribute.String("req_path", "/")),
			)

			fmt.Printf("Sent metric:  #%-2d  increment=%-2d\n", i+1, incr)
			time.Sleep(randMs(500, 1000))
		}
	}

	// Traces
	{
		tracer, err := newOTelTracer(ctx, *grpc, addr, name, version, baseKV...)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		ctx, span := tracer.Start(ctx, "demo-opentelemetry-trace")
		defer span.End()

		for i, msg := range messages {
			_, span := tracer.Start(ctx, fmt.Sprintf("demo-op-%d", i+1))
			defer span.End()

			span.SetAttributes(
				attribute.String("uuid", uuid.NewString()),
				attribute.String("message", msg),
			)

			fmt.Printf("Sent span:  #%-2d  message=%s\n", i+1, msg)
			time.Sleep(randMs(500, 1000))
		}
	}
}

func newOTelLogger(ctx context.Context, grpc bool, endpoint, name, version string, kv ...string) (log.Logger, error) {
	var err error
	var exp logsdk.Exporter

	if grpc {
		exp, err = otlploggrpc.New(ctx,
			otlploggrpc.WithEndpoint(endpoint),
			otlploggrpc.WithCompressor("gzip"),
			otlploggrpc.WithInsecure(),
		)
	} else {
		exp, err = otlploghttp.New(ctx,
			otlploghttp.WithEndpoint(endpoint),
			otlploghttp.WithCompression(otlploghttp.GzipCompression),
			otlploghttp.WithInsecure(),
		)
	}

	if err != nil {
		return nil, err
	}

	prov := logsdk.NewLoggerProvider(
		logsdk.WithProcessor(
			logsdk.NewBatchProcessor(exp),
		),
		logsdk.WithResource(
			createResource(name, version, kv),
		),
	)

	// Configure scope attributes for the logger.
	logger := prov.Logger(name,
		log.WithInstrumentationVersion(version),
	)

	return logger, nil
}

func sendOTelLog(ctx context.Context, logger log.Logger, t time.Time, level, message string, kv ...string) {
	var r log.Record
	var sev log.Severity

	switch strings.ToLower(level) {
	case "fatal":
		sev = log.SeverityFatal
	case "error":
		sev = log.SeverityError
	case "warn":
		sev = log.SeverityWarn
	case "info":
		sev = log.SeverityInfo
	case "debug":
		sev = log.SeverityDebug
	default:
		sev = log.SeverityUndefined
	}

	// Add the fixed fields: timestamp, severity, body
	r.SetTimestamp(t)
	r.SetObservedTimestamp(t)
	r.SetSeverity(sev)
	r.SetSeverityText(level)
	r.SetBody(log.StringValue(message))

	// Add the fields
	for i := 0; i+1 < len(kv); i += 2 {
		k, v := kv[i], kv[i+1]
		r.AddAttributes(log.String(k, v))
	}

	logger.Emit(ctx, r)
}

func newOTelMeter(ctx context.Context, grpc bool, endpoint, name, version string, kv ...string) (metric.Meter, error) {
	var err error
	var exp metricsdk.Exporter

	if grpc {
		exp, err = otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithEndpoint(endpoint),
			otlpmetricgrpc.WithCompressor("gzip"),
			otlpmetricgrpc.WithInsecure(),
		)
	} else {
		exp, err = otlpmetrichttp.New(ctx,
			otlpmetrichttp.WithEndpoint(endpoint),
			otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression),
			otlpmetrichttp.WithInsecure(),
		)
	}

	if err != nil {
		return nil, err
	}

	prov := metricsdk.NewMeterProvider(
		metricsdk.WithReader(
			metricsdk.NewPeriodicReader(exp),
		),
		metricsdk.WithResource(
			createResource(name, version, kv),
		),
	)

	// Configure scope attributes for the meter.
	meter := prov.Meter(name,
		metric.WithInstrumentationVersion(version),
	)

	return meter, nil
}

func newOTelTracer(ctx context.Context, grpc bool, endpoint, name, version string, kv ...string) (trace.Tracer, error) {
	var err error
	var exp tracesdk.SpanExporter

	if grpc {
		exp, err = otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithCompressor("gzip"),
			otlptracegrpc.WithInsecure(),
		)
	} else {
		exp, err = otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(endpoint),
			otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
			otlptracehttp.WithInsecure(),
		)
	}

	if err != nil {
		return nil, err
	}

	prov := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithSampler(
			tracesdk.AlwaysSample(),
		),
		tracesdk.WithResource(
			createResource(name, version, kv),
		),
	)

	// Configure scope attributes for the tracer.
	tracer := prov.Tracer(name,
		trace.WithInstrumentationVersion(version),
	)

	return tracer, nil
}

func createResource(name, version string, kv []string) *resource.Resource {
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(name),
		semconv.ServiceVersionKey.String(version),
	}

	for i := 0; i+1 < len(kv); i += 2 {
		attrs = append(attrs, attribute.String(kv[i], kv[i+1]))
	}

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		attrs...,
	)

	return resource
}
