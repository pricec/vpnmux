package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/pricec/vpnmux/pkg/api/v1"
	"github.com/pricec/vpnmux/pkg/config"
)

type HealthHandler struct{}

func (h HealthHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type Server struct {
	server          *http.Server
	shutdownTimeout time.Duration
}

type ServerOptions struct {
	Config *config.Config
}

func NewServer(ctx context.Context, opts ServerOptions) (*Server, error) {
	r := mux.NewRouter()
	r.Handle("/healthz", HealthHandler{}).Methods("GET")
	v1.RegisterHandlers(ctx, r.PathPrefix("/v1").Subrouter(), opts.Config)

	s := &Server{
		server: &http.Server{
			// TODO: make listen interface configurable
			Addr:    fmt.Sprintf(":%d", opts.Config.ListenPort),
			Handler: r,
		},
		shutdownTimeout: opts.Config.ShutdownTimeout,
	}
	go s.start()
	return s, nil
}

func (s *Server) start() {
	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("error stopping server: %v", err)
	}
}

func (s *Server) Close(ctx context.Context) error {
	ctx2, cancel := context.WithTimeout(ctx, s.shutdownTimeout)
	defer cancel()

	err := s.server.Shutdown(ctx2)
	if errors.Is(err, context.DeadlineExceeded) {
		err = s.server.Close()
	}
	return err
}
