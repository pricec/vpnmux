package reconciler

import (
	"context"
	"fmt"

	"github.com/pricec/vpnmux/pkg/database"
	"github.com/pricec/vpnmux/pkg/network"
)

type ClientNetworkReconciler struct {
	db *database.Database
}

func (r *ClientNetworkReconciler) Update(ctx context.Context, cfg *database.ClientNetwork) (*database.ClientNetwork, error) {
	return nil, fmt.Errorf("TODO: implement client-network update reconciler")
}

func NewClientNetworkReconciler(ctx context.Context, db *database.Database) (*ClientNetworkReconciler, error) {
	nets, err := db.ClientNetworks.List(ctx)
	if err != nil {
		return nil, err
	}

	r := &ClientNetworkReconciler{db: db}
	for _, net := range nets {
		if _, _, err := r.check(ctx, net.ClientID); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *ClientNetworkReconciler) check(ctx context.Context, id string) (*database.ClientNetwork, *network.Client, error) {
	net, err := r.db.ClientNetworks.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	c, err := r.db.Clients.Get(ctx, net.ClientID)
	if err != nil {
		return nil, nil, err
	}

	client, err := network.NewClient(c.Address)
	if err != nil {
		return nil, nil, err
	}

	if net.NetworkID != "" {
		dockerNet, err := network.NewFromID(net.NetworkID)
		if err != nil {
			return nil, nil, err
		}

		if err := client.SetRouteTable(dockerNet.Container.RouteTableID); err != nil {
			return nil, nil, err
		}
	} else {
		if err := client.ClearRoutes(); err != nil {
			return nil, nil, err
		}
	}

	// TODO: check if settings are equal?
	return net, client, nil
}

func (r *ClientNetworkReconciler) Get(ctx context.Context, id string) (*database.ClientNetwork, *network.Client, error) {
	return r.check(ctx, id)
}

func (r *ClientNetworkReconciler) Create(ctx context.Context, cn *database.ClientNetwork) (*database.ClientNetwork, error) {
	net, err := r.db.ClientNetworks.Put(ctx, cn)
	if err != nil {
		return nil, err
	}

	if _, _, err := r.check(ctx, net.ClientID); err != nil {
		// TODO: clean up
		return nil, err

	}

	return net, nil
}

func (r *ClientNetworkReconciler) Delete(ctx context.Context, id string) error {
	_, client, err := r.check(ctx, id)
	if err != nil {
		return err
	}

	if err := r.db.ClientNetworks.Delete(ctx, id); err != nil {
		return err
	}

	if err := client.ClearRoutes(); err != nil {
		return err
	}

	return nil
}
