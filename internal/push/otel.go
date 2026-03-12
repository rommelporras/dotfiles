package push

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	otelmetric "go.opentelemetry.io/otel/metric"

	"github.com/rommelporras/dotfiles/internal/model"
)

// MetricSet is the intermediate representation before pushing.
type MetricSet struct {
	Hostname       string
	Platform       string
	Context        string
	Up             int64
	DriftTotal     int64
	ToolsInstalled map[string]int64
	Credentials    map[string]int64
	Timestamp      int64
}

// BuildMetrics converts a MachineState into pushable metrics.
func BuildMetrics(state *model.MachineState) *MetricSet {
	tools := make(map[string]int64, len(state.Tools))
	for name, path := range state.Tools {
		if path != "" {
			tools[name] = 1
		} else {
			tools[name] = 0
		}
	}

	creds := map[string]int64{
		"ssh_agent":   boolToInt(state.SSHAgent != "none" && state.SSHAgent != "n/a"),
		"setup_creds": boolToInt(state.SetupCreds == "ran"),
		"atuin_sync":  boolToInt(state.AtuinSync == "synced"),
	}

	return &MetricSet{
		Hostname:       state.Hostname,
		Platform:       state.Platform,
		Context:        state.Context,
		Up:             1,
		DriftTotal:     int64(len(state.DriftFiles)),
		ToolsInstalled: tools,
		Credentials:    creds,
		Timestamp:      time.Now().Unix(),
	}
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func newResource(ctx context.Context) (*sdkresource.Resource, error) {
	return sdkresource.New(ctx,
		sdkresource.WithAttributes(
			semconv.ServiceName("dotctl"),
		),
	)
}

// Push sends metrics for a MachineState to the OTel Collector via OTLP gRPC.
func Push(ctx context.Context, endpoint string, state *model.MachineState) error {
	res, err := newResource(ctx)
	if err != nil {
		return fmt.Errorf("create resource: %w", err)
	}

	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("create metric exporter: %w", err)
	}

	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(
		metric.WithReader(reader),
		metric.WithResource(res),
	)
	defer provider.Shutdown(ctx)

	meter := provider.Meter("dotctl")
	ms := BuildMetrics(state)

	hostAttr := attribute.String("hostname", ms.Hostname)
	platAttr := attribute.String("platform", ms.Platform)
	ctxAttr := attribute.String("context", ms.Context)

	upGauge, _ := meter.Int64Gauge("dotctl_up")
	driftGauge, _ := meter.Int64Gauge("dotctl_drift_total")
	toolGauge, _ := meter.Int64Gauge("dotctl_tool_installed")
	credGauge, _ := meter.Int64Gauge("dotctl_credential_status")
	tsGauge, _ := meter.Int64Gauge("dotctl_collect_timestamp")

	upGauge.Record(ctx, ms.Up, otelmetric.WithAttributes(hostAttr, platAttr, ctxAttr))
	driftGauge.Record(ctx, ms.DriftTotal, otelmetric.WithAttributes(hostAttr))
	tsGauge.Record(ctx, ms.Timestamp, otelmetric.WithAttributes(hostAttr))

	for tool, val := range ms.ToolsInstalled {
		toolGauge.Record(ctx, val, otelmetric.WithAttributes(hostAttr, attribute.String("tool", tool)))
	}
	for cred, val := range ms.Credentials {
		credGauge.Record(ctx, val, otelmetric.WithAttributes(hostAttr, attribute.String("credential", cred)))
	}

	var rm metricdata.ResourceMetrics
	if err := reader.Collect(ctx, &rm); err != nil {
		return fmt.Errorf("collect metrics: %w", err)
	}
	if err := exporter.Export(ctx, &rm); err != nil {
		return fmt.Errorf("export metrics: %w", err)
	}
	if err := exporter.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown exporter: %w", err)
	}

	return nil
}

// PushLog sends the full MachineState as a structured JSON log to the OTel Collector.
func PushLog(ctx context.Context, endpoint string, state *model.MachineState) error {
	res, err := newResource(ctx)
	if err != nil {
		return fmt.Errorf("create resource: %w", err)
	}

	exporter, err := otlploggrpc.New(ctx,
		otlploggrpc.WithEndpoint(endpoint),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("create log exporter: %w", err)
	}

	provider := log.NewLoggerProvider(
		log.WithProcessor(log.NewSimpleProcessor(exporter)),
		log.WithResource(res),
	)
	defer provider.Shutdown(ctx)

	logger := provider.Logger("dotctl")

	payload, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	var record otellog.Record
	record.SetBody(otellog.StringValue(string(payload)))
	record.SetTimestamp(time.Now())
	record.AddAttributes(
		otellog.String("hostname", state.Hostname),
		otellog.String("platform", state.Platform),
		otellog.String("context", state.Context),
	)

	logger.Emit(ctx, record)
	return nil
}
