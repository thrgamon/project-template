package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/exemplar"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Init sets up OpenTelemetry trace, metric, and log providers.
// If OTEL_EXPORTER_OTLP_ENDPOINT is not set, telemetry is disabled.
// Returns a logger that bridges slog records to OTEL when enabled.
func Init(ctx context.Context, serviceName string) (logger *slog.Logger, shutdown func(context.Context) error, err error) {
	baseHandler := slog.NewJSONHandler(os.Stdout, nil)

	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		logger = slog.New(baseHandler)
		logger.Info("telemetry disabled (OTEL_EXPORTER_OTLP_ENDPOINT not set)")
		return logger, func(context.Context) error { return nil }, nil
	}

	var shutdowns []func(context.Context) error

	shutdown = func(ctx context.Context) error {
		var errs []error
		for _, fn := range shutdowns {
			if err := fn(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	}

	handleErr := func(inErr error) {
		if inErr != nil {
			err = errors.Join(inErr, shutdown(ctx))
		}
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, nil, err
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	traceExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter, trace.WithBatchTimeout(5*time.Second)),
		trace.WithResource(res),
	)
	shutdowns = append(shutdowns, tp.Shutdown)
	otel.SetTracerProvider(tp)

	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		handleErr(err)
		return
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(30*time.Second))),
		metric.WithResource(res),
		metric.WithExemplarFilter(exemplar.AlwaysOnFilter),
	)
	shutdowns = append(shutdowns, mp.Shutdown)
	otel.SetMeterProvider(mp)

	if rtErr := runtime.Start(runtime.WithMeterProvider(mp)); rtErr != nil {
		handleErr(rtErr)
		return
	}

	logExporter, err := otlploghttp.New(ctx)
	if err != nil {
		handleErr(err)
		return
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(res),
	)
	shutdowns = append(shutdowns, lp.Shutdown)

	otelHandler := otelslog.NewHandler(serviceName, otelslog.WithLoggerProvider(lp))
	traceAwareStdout := &traceContextHandler{inner: baseHandler}
	fanoutHandler := &fanoutSlogHandler{handlers: []slog.Handler{traceAwareStdout, otelHandler}}
	logger = slog.New(fanoutHandler)

	logger.Info("telemetry enabled", "endpoint", endpoint)
	return logger, shutdown, nil
}

type fanoutSlogHandler struct {
	handlers []slog.Handler
	attrs    []slog.Attr
	groups   []string
}

func (h *fanoutSlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *fanoutSlogHandler) Handle(ctx context.Context, record slog.Record) error {
	var errs []error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, record.Level) {
			if err := handler.Handle(ctx, record); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func (h *fanoutSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &fanoutSlogHandler{
		handlers: newHandlers,
		attrs:    append(h.attrs, attrs...),
		groups:   h.groups,
	}
}

func (h *fanoutSlogHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &fanoutSlogHandler{
		handlers: newHandlers,
		attrs:    h.attrs,
		groups:   append(h.groups, name),
	}
}

type traceContextHandler struct {
	inner slog.Handler
}

func (h *traceContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *traceContextHandler) Handle(ctx context.Context, record slog.Record) error {
	if sc := oteltrace.SpanContextFromContext(ctx); sc.IsValid() {
		record.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
		if sc.TraceFlags().IsSampled() {
			record.AddAttrs(slog.Bool("trace_sampled", true))
		}
	}
	return h.inner.Handle(ctx, record)
}

func (h *traceContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceContextHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h *traceContextHandler) WithGroup(name string) slog.Handler {
	return &traceContextHandler{inner: h.inner.WithGroup(name)}
}
