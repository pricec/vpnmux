package reconciler

import (
	"context"

	"github.com/pricec/vpnmux/pkg/database"
)

type ForwardingOptions struct {
	LANInterface string
	WANInterface string
	DNSMark      string
}

type Options struct {
	DB         *database.Database
	Network    NetworkReconcilerOptions
	Forwarding ForwardingOptions
}

type Reconciler struct {
	db             *database.Database
	Configs        *ConfigReconciler
	Networks       *NetworkReconciler
	Clients        *ClientReconciler
	ClientNetworks *ClientNetworkReconciler
	DNS            *DNSReconciler
}

func New(ctx context.Context, opts Options) (*Reconciler, error) {
	configs, err := NewConfigReconciler(ctx, opts.DB)
	if err != nil {
		return nil, err
	}

	networks, err := NewNetworkReconciler(ctx, opts.DB, opts.Network)
	if err != nil {
		return nil, err
	}

	clients, err := NewClientReconciler(ctx, opts.DB, opts.Forwarding)
	if err != nil {
		return nil, err
	}

	clientNetworks, err := NewClientNetworkReconciler(ctx, opts.DB, opts.Forwarding)
	if err != nil {
		return nil, err
	}

	dns, err := NewDNSReconciler(ctx, opts.DB, opts.Forwarding.DNSMark)
	if err != nil {
		return nil, err
	}

	return &Reconciler{
		db:             opts.DB,
		Configs:        configs,
		Networks:       networks,
		Clients:        clients,
		ClientNetworks: clientNetworks,
		DNS:            dns,
	}, nil
}

func (r *Reconciler) CreateNetwork(ctx context.Context, n *database.Network) (*database.Network, error) {
	_, cfg, err := r.Configs.Get(ctx, n.ConfigID)
	if err != nil {
		return nil, err
	}
	return r.Networks.create(ctx, n.Name, cfg)
}
