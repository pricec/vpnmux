package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/pricec/vpnmux/pkg/api/v0"
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
	ShutdownTimeout time.Duration
	ListenPort      uint16
}

func NewServer(opts ServerOptions) (*Server, error) {
	r := mux.NewRouter()
	r.Handle("/healthz", HealthHandler{}).Methods("GET")
	v0.RegisterHandlers(r.PathPrefix("/v0").Subrouter())

	s := &Server{
		server: &http.Server{
			// TODO: make listen interface configurable
			Addr:    fmt.Sprintf(":%d", opts.ListenPort),
			Handler: r,
		},
		shutdownTimeout: opts.ShutdownTimeout,
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
