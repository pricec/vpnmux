package reconciler

import (
	"context"

	"github.com/pricec/vpnmux/pkg/database"
)

type Options struct {
	DB      *database.Database
	Network NetworkReconcilerOptions
}

type Reconciler struct {
	db             *database.Database
	Configs        *ConfigReconciler
	Networks       *NetworkReconciler
	Clients        *ClientReconciler
	ClientNetworks *ClientNetworkReconciler
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

	clients, err := NewClientReconciler(ctx, opts.DB)
	if err != nil {
		return nil, err
	}

	clientNetworks, err := NewClientNetworkReconciler(ctx, opts.DB)
	if err != nil {
		return nil, err
	}

	return &Reconciler{
		db:             opts.DB,
		Configs:        configs,
		Networks:       networks,
		Clients:        clients,
		ClientNetworks: clientNetworks,
	}, nil
}

func (r *Reconciler) CreateNetwork(ctx context.Context, n *database.Network) (*database.Network, error) {
	_, cfg, err := r.Configs.Get(ctx, n.ConfigID)
	if err != nil {
		return nil, err
	}
	return r.Networks.create(ctx, n.Name, cfg)
}
