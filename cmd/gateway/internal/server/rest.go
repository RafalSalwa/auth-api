package server

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/RafalSalwa/interview-app-srv/cmd/gateway/config"
	"github.com/RafalSalwa/interview-app-srv/pkg/logger"
	"github.com/RafalSalwa/interview-app-srv/pkg/tracing"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"net/http"
	_ "net/http/pprof"
)

type Server struct {
	srv *http.Server
	log *logger.Logger
	cfg *config.Config
}

func NewServer(cfg *config.Config, r *mux.Router, l *logger.Logger) *Server {

	tlsConf := new(tls.Config)
	s := &http.Server{
		Addr:         cfg.Http.Addr,
		Handler:      r,
		ReadTimeout:  cfg.Http.TimeoutRead,
		WriteTimeout: cfg.Http.TimeoutWrite,
		IdleTimeout:  cfg.Http.TimeoutIdle,
		TLSConfig:    tlsConf,
	}

	return &Server{
		srv: s,
		log: l,
		cfg: cfg,
	}
}

func (srv *Server) ServeHTTP() {
	go func() {
		srv.log.Info().Msgf("Starting server server on: %v [auth method: %s]", srv.srv.Addr, srv.cfg.Auth.AuthMethod)
		if srv.cfg.App.Env == "dev" {
			if err := srv.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				srv.log.Error().Err(err).Msg("server:Listen")
			}
		} else {
			if err := srv.srv.ListenAndServeTLS(
				"/etc/ssl/certs/server.crt",
				"/etc/ssl/private/server.key"); err != nil && !errors.Is(err, http.ErrServerClosed) {
				srv.log.Error().Err(err).Msg("server:ListenTLS")
			}
		}
	}()

	if srv.cfg.Jaeger.Enable {
		tp, err := tracing.NewJaegerTracer(*srv.cfg.Jaeger)
		if err != nil {
			srv.log.Error().Err(err).Msg("server:jaeger:register")
		}
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	}
}

func (srv *Server) Shutdown() {
	closed := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), srv.srv.IdleTimeout)
	defer cancel()

	if err := srv.srv.Shutdown(ctx); err != nil {
		srv.log.Error().Err(err).Msg("server:shutdown")
	}

	close(closed)
}
