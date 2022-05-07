package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pricec/vpnmux/pkg/api"
	"github.com/pricec/vpnmux/pkg/config"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("error reading configuration: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	doneCh := make(chan struct{})
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
		signal.Reset()
		close(sigCh)
		close(doneCh)
	}()

	server, err := api.NewServer(ctx, api.ServerOptions{
		Config: cfg,
	})
	if err != nil {
		log.Fatalf("error setting up API server: %v", err)
	}
	defer func() {
		err := server.Close(ctx)
		if err != nil {
			log.Printf("error closing server: %v", err)
		}
	}()

	<-doneCh
}
