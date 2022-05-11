package reconciler

import (
	"context"
	"fmt"

	"github.com/pricec/vpnmux/pkg/database"
	"github.com/pricec/vpnmux/pkg/network"
)

type ClientReconciler struct {
	db         *database.Database
	forwarding ForwardingOptions
}

func (r *ClientReconciler) Update(ctx context.Context, cfg *database.Client) (*database.Client, error) {
	return nil, fmt.Errorf("TODO: implement client update reconciler")
}

func NewClientReconciler(ctx context.Context, db *database.Database, forwarding ForwardingOptions) (*ClientReconciler, error) {
	clients, err := db.Clients.List(ctx)
	if err != nil {
		return nil, err
	}

	r := &ClientReconciler{
		db:         db,
		forwarding: forwarding,
	}
	for _, client := range clients {
		if _, _, err := r.check(ctx, client.ID); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *ClientReconciler) check(ctx context.Context, id string) (*database.Client, *network.Client, error) {
	client, err := r.db.Clients.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	networkClient, err := network.NewClient(client.Address, r.forwarding.LANInterface, r.forwarding.WANInterface)
	if err != nil {
		return nil, nil, err
	}

	return client, networkClient, nil
}

func (r *ClientReconciler) Get(ctx context.Context, id string) (*database.Client, *network.Client, error) {
	return r.check(ctx, id)
}

func (r *ClientReconciler) Create(ctx context.Context, c *database.Client) (*database.Client, error) {
	client, err := r.db.Clients.Put(ctx, c)
	if err != nil {
		return nil, err
	}

	_, err = network.NewClient(client.Address, r.forwarding.LANInterface, r.forwarding.WANInterface)
	if err != nil {
		// TODO: clean up database
		return nil, err
	}

	if _, _, err := r.check(ctx, client.ID); err != nil {
		// TODO: clean up
		return nil, err

	}

	return client, nil
}

func (r *ClientReconciler) Delete(ctx context.Context, id string) error {
	client, networkClient, err := r.check(ctx, id)
	if err != nil {
		return err
	}

	if err := r.db.Clients.Delete(ctx, client.ID); err != nil {
		return err
	}

	if err := networkClient.Close(); err != nil {
		return err
	}
	return nil
}
