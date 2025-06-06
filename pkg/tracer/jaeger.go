package tracer

import (
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"bytedancemall/seckill/internal/config"
)

// InitJaeger 初始化Jaeger TracerProvider和Exporter
func InitJaeger(cfg *config.Config) (*tracesdk.TracerProvider, error) {
	// 创建Jaeger Exporter
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(
		jaeger.WithEndpoint(cfg.Jaeger.Endpoint), // Jaeger收集器端点
	))
	if err != nil {
		slog.Error("Failed to create Jaeger exporter", "error", err)
		return nil, err
	}

	// 创建TracerProvider
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exporter), // 使用批量导出器提高性能
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.Jaeger.ServiceName), // 服务名 - 会在Jaeger UI中显示
			semconv.ServiceVersionKey.String("1.0.0"),             // 服务版本
			semconv.DeploymentEnvironmentKey.String("dev"),        // 环境
		)), // 附加服务信息
		tracesdk.WithSampler(tracesdk.AlwaysSample()), // 100%采样（生产环境可调整）
	)
	slog.Info("Jaeger exporter created successfully")

	// 设置为全局TracerProvider！
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})) // 添加这一行
	client := http.Client{Timeout: 2 * time.Second}
	for range 30 {
		_, err = client.Get(cfg.Jaeger.Endpoint)
		if err == nil {
			return tp, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil, err
}
