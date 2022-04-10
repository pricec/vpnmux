package reconciler

import (
	"context"

	"github.com/pricec/vpnmux/pkg/database"
)

type Reconciler struct {
	db       *database.Database
	Configs  *ConfigReconciler
	Networks *NetworkReconciler
}

func New(ctx context.Context, db *database.Database) (*Reconciler, error) {
	configs, err := NewConfigReconciler(ctx, db)
	if err != nil {
		return nil, err
	}

	networks, err := NewNetworkReconciler(ctx, db)
	if err != nil {
		return nil, err
	}

	return &Reconciler{
		db:       db,
		Configs:  configs,
		Networks: networks,
	}, nil
}

func (r *Reconciler) CreateNetwork(ctx context.Context, n *database.Network) (*database.Network, error) {
	_, cfg, err := r.Configs.Get(ctx, n.ConfigID)
	if err != nil {
		return nil, err
	}
	return r.Networks.create(ctx, n.Name, cfg)
}
