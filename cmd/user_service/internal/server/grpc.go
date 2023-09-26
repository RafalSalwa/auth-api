package server

import (
	"context"
	"fmt"
	"github.com/RafalSalwa/interview-app-srv/cmd/user_service/internal/rpc_api"
	"github.com/RafalSalwa/interview-app-srv/cmd/user_service/internal/services"
	grpc_config "github.com/RafalSalwa/interview-app-srv/pkg/grpc"
	"github.com/RafalSalwa/interview-app-srv/pkg/logger"
	"github.com/RafalSalwa/interview-app-srv/pkg/probes"
	"github.com/RafalSalwa/interview-app-srv/pkg/tracing"
	pb "github.com/RafalSalwa/interview-app-srv/proto/grpc"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpclogrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	grpcctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"net"
	"os"
	"runtime/debug"
	"time"
)

type GRPC struct {
	pb.UnimplementedAuthServiceServer
	pb.UnimplementedUserServiceServer
	config      grpc_config.Config
	probing     probes.Config
	userService services.UserService
}

func NewGrpcServer(config grpc_config.Config,
	probesCfg probes.Config,
	userService services.UserService) (*GRPC, error) {

	srv := &GRPC{
		config:      config,
		probing:     probesCfg,
		userService: userService,
	}

	return srv, nil
}

func (s GRPC) Run() {
	logEntry := logger.NewGRPCLogger()
	grpclogrus.ReplaceGrpcLogger(logEntry)

	opts := []grpclogrus.Option{
		grpclogrus.WithLevels(func(code codes.Code) logrus.Level {
			if code == codes.OK {
				return logrus.InfoLevel
			}
			return logrus.ErrorLevel
		}),

		grpclogrus.WithDurationField(func(duration time.Duration) (key string, value interface{}) {
			return "grpc.time_ms", duration.Milliseconds()
		}),
	}
	srvMetrics := grpcprom.NewServerMetrics(
		grpcprom.WithServerHandlingTimeHistogram(
			grpcprom.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)
	reg := prometheus.NewRegistry()
	reg.MustRegister(srvMetrics)
	exemplarFromContext := func(ctx context.Context) prometheus.Labels {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return prometheus.Labels{"traceID": span.TraceID().String()}
		}
		return nil
	}

	logger := log.NewLogfmtLogger(os.Stderr)
	rpcLogger := log.With(logger, "service", "gRPC/server", "component", "user_service")
	logTraceID := func(ctx context.Context) logging.Fields {
		if span := trace.SpanContextFromContext(ctx); span.IsSampled() {
			return logging.Fields{"traceID", span.TraceID().String()}
		}
		return nil
	}
	panicsTotal := promauto.With(reg).NewCounter(prometheus.CounterOpts{
		Name: "grpc_req_panics_recovered_total",
		Help: "Total number of gRPC requests recovered from internal panic.",
	})
	grpcPanicRecoveryHandler := func(p any) (err error) {
		panicsTotal.Inc()
		level.Error(rpcLogger).Log("msg", "recovered from panic", "panic", p, "stack", debug.Stack())
		return status.Errorf(codes.Internal, "%s", p)
	}

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpcmiddleware.ChainStreamServer(
			otelgrpc.StreamServerInterceptor(),
			srvMetrics.StreamServerInterceptor(grpcprom.WithExemplarFromContext(exemplarFromContext)),
			logging.StreamServerInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID)),
			grpcctxtags.StreamServerInterceptor(),
			grpclogrus.StreamServerInterceptor(logEntry, opts...),
			recovery.StreamServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		)),
		grpc.ChainUnaryInterceptor(grpcmiddleware.ChainUnaryServer(
			otelgrpc.UnaryServerInterceptor(),
			srvMetrics.UnaryServerInterceptor(grpcprom.WithExemplarFromContext(exemplarFromContext)),
			logging.UnaryServerInterceptor(interceptorLogger(rpcLogger), logging.WithFieldsFromContext(logTraceID)),
			grpcctxtags.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpclogrus.UnaryServerInterceptor(logEntry, opts...),
			recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(grpcPanicRecoveryHandler)),
		)),
	)

	userServer, err := rpc_api.NewGrpcUserServer(s.config, s.userService)
	if err != nil {
		logger.Log(err)
	}
	pb.RegisterUserServiceServer(grpcServer, userServer)
	reflection.Register(grpcServer)
	tracing.RegisterMetricsEndpoint(s.probing.Port)
	listener, err := net.Listen("tcp", s.config.Addr)

	if err != nil {
		logger.Log(err)
		//s.logger.Error().Err(err)
	}

	logger.Log("Starting gRPC server on: %s", s.config.Addr)
	if err = grpcServer.Serve(listener); err != nil {
		logger.Log("serve:", err)
		//s.logger.Error().Err(err)
	}
	grpcServer.GracefulStop()
}

func interceptorLogger(l log.Logger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		largs := append([]any{"msg", msg}, fields...)
		switch lvl {
		case logging.LevelDebug:
			_ = level.Debug(l).Log(largs...)
		case logging.LevelInfo:
			_ = level.Info(l).Log(largs...)
		case logging.LevelWarn:
			_ = level.Warn(l).Log(largs...)
		case logging.LevelError:
			_ = level.Error(l).Log(largs...)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
