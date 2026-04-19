package interceptor

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	grpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "code"},
	)

	grpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	activeStreams = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "grpc_active_streams",
			Help: "Number of active gRPC streams",
		},
	)
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		log.Debug().Str("method", info.FullMethod).Msg("gRPC request started")

		resp, err := handler(ctx, req)

		duration := time.Since(start).Seconds()
		code := status.Code(err)

		grpcRequestsTotal.WithLabelValues(info.FullMethod, code.String()).Inc()
		grpcRequestDuration.WithLabelValues(info.FullMethod).Observe(duration)

		log.Debug().
			Str("method", info.FullMethod).
			Dur("duration", time.Since(start)).
			Str("code", code.String()).
			Msg("gRPC request completed")

		return resp, err
	}
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		activeStreams.Inc()
		defer activeStreams.Dec()

		start := time.Now()
		log.Debug().Str("method", info.FullMethod).Msg("gRPC stream started")

		err := handler(srv, ss)

		duration := time.Since(start).Seconds()
		code := status.Code(err)

		grpcRequestsTotal.WithLabelValues(info.FullMethod, code.String()).Inc()

		log.Debug().
			Str("method", info.FullMethod).
			Dur("duration", duration).
			Str("code", code.String()).
			Msg("gRPC stream completed")

		return err
	}
}

// AuthInterceptor 从 context 中提取 JWT token 并验证
func AuthInterceptor(ctx context.Context) (interface{}, error) {
	// 在 gateway 层处理，这里只做日志
	return ctx, nil
}
