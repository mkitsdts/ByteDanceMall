package interceptor

import (
	"context"
	"log/slog"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

const UserIDKey = "id"

// wrappedServerStream 包装grpc.ServerStream，使其使用自定义上下文
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context 返回包含用户ID的上下文
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// TracerInterceptor 实现链路追踪拦截器
type TracerInterceptor struct {
}

// NewTracerInterceptor 创建链路追踪拦截器
func NewTracerInterceptor() *TracerInterceptor {
	return &TracerInterceptor{}
}

// Unary 一元RPC的链路追踪拦截器
func (i *TracerInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		tracer := otel.Tracer("seckill")
		ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		// 设置span的属性
		span.SetAttributes(
			attribute.String("rpc.method", info.FullMethod),
			attribute.String("rpc.service", info.FullMethod[:len(info.FullMethod)-len(info.FullMethod[strings.LastIndex(info.FullMethod, "/")+1:])]),
		)

		// 从上下文中获取用户ID（如果有）
		if userID, ok := ctx.Value(UserIDKey).(string); ok {
			span.SetAttributes(attribute.String("id", userID))
		}

		resp, err := handler(ctx, req)

		if err != nil {
			slog.Error("RPC failed", "method", info.FullMethod, "error", err)
		}
		slog.Info("Tracer completed", "method", info.FullMethod, "error", err)
		return resp, err
	}
}

// Stream 流式RPC的链路追踪拦截器
func (i *TracerInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		tracer := otel.Tracer("tx-service")
		ctx, span := tracer.Start(ss.Context(), info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		// 设置span的属性
		span.SetAttributes(
			attribute.String("rpc.method", info.FullMethod),
			attribute.String("rpc.service", info.FullMethod[:len(info.FullMethod)-len(info.FullMethod[strings.LastIndex(info.FullMethod, "/")+1:])]),
		)
		// 从上下文中获取用户ID（如果有）
		if userID, ok := ctx.Value(UserIDKey).(string); ok {
			span.SetAttributes(attribute.String("id", userID))
		}

		// 包装流，使用带有跟踪信息的上下文
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		err := handler(srv, wrappedStream)
		if err != nil {
			slog.Error("Stream RPC failed", "method", info.FullMethod, "error", err)
		}
		slog.Info("Stream Tracer completed", "method", info.FullMethod, "error", err)
		return err
	}
}
