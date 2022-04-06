package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pricec/vpnmux/pkg/api"
)

func main() {
	sigCh := make(chan os.Signal, 1)
	doneCh := make(chan struct{})
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		signal.Reset()
		close(sigCh)
		close(doneCh)
	}()

	server, err := api.NewServer(api.ServerOptions{
		ShutdownTimeout: 10 * time.Second,
		ListenPort:      8080,
	})
	if err != nil {
		log.Fatalf("error setting up API server: %v", err)
	}
	defer func() {
		err := server.Close(context.Background())
		if err != nil {
			log.Printf("error closing server: %v", err)
		}
	}()

	/*
		rt, err := network.NewVPNRoutingTable(100, c.IPAddress)
		if err != nil {
			log.Fatalf("error setting up routing table: %v", err)
		}

		if err := rt.AddSource("192.168.1.60"); err != nil {
			log.Printf("error adding source to routing table: %v", err)
		}
	*/

	<-doneCh
}
