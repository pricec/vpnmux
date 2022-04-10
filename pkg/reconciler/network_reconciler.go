package reconciler

import (
	"context"
	"fmt"

	"github.com/pricec/vpnmux/pkg/database"
	"github.com/pricec/vpnmux/pkg/network"
	"github.com/pricec/vpnmux/pkg/openvpn"
)

type NetworkReconciler struct {
	db *database.Database
}

func (r *NetworkReconciler) Update(ctx context.Context, cfg *database.Network) (*database.Network, error) {
	return nil, fmt.Errorf("TODO: implement network update reconciler")
}

func NewNetworkReconciler(ctx context.Context, db *database.Database) (*NetworkReconciler, error) {
	nets, err := db.Networks.List(ctx)
	if err != nil {
		return nil, err
	}

	r := &NetworkReconciler{db: db}
	for _, net := range nets {
		if _, _, err := r.check(ctx, net.ID); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *NetworkReconciler) check(ctx context.Context, id string) (*database.Network, *network.Network, error) {
	net, err := r.db.Networks.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	dockerNet, err := network.NewFromID(net.ID)
	if err != nil {
		return nil, nil, err
	}

	// TODO: check if settings are equal?
	return net, dockerNet, nil
}

func (r *NetworkReconciler) Get(ctx context.Context, id string) (*database.Network, *network.Network, error) {
	return r.check(ctx, id)
}

func (r *NetworkReconciler) create(ctx context.Context, name string, cfg *openvpn.Config) (*database.Network, error) {
	net, err := r.db.Networks.Put(ctx, &database.Network{
		Name:     name,
		ConfigID: cfg.ID,
	})
	if err != nil {
		return nil, err
	}

	_, err = network.New(net.ID, cfg)
	if err != nil {
		// TODO: clean up database
		return nil, err
	}

	if _, _, err := r.check(ctx, net.ID); err != nil {
		// TODO: clean up
		return nil, err

	}

	return net, nil
}

func (r *NetworkReconciler) Delete(ctx context.Context, id string) error {
	net, dockerNet, err := r.check(ctx, id)
	if err != nil {
		return err
	}

	if err := r.db.Networks.Delete(ctx, net.ID); err != nil {
		return err
	}

	if err := dockerNet.Close(); err != nil {
		return err
	}
	return nil
}
