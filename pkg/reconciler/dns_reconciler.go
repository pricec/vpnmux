package reconciler

import (
	"context"
	"fmt"

	"github.com/pricec/vpnmux/pkg/database"
	"github.com/pricec/vpnmux/pkg/network"
)

type DNSReconciler struct {
	db     *database.Database
	router *network.DNSRouter
}

func (r *DNSReconciler) Update(ctx context.Context, cfg *database.DNSRoute) (*database.DNSRoute, error) {
	return nil, fmt.Errorf("TODO: implement DNS update reconciler")
}

func NewDNSReconciler(ctx context.Context, db *database.Database, mark, localSubnet string) (*DNSReconciler, error) {
	router, err := network.NewDNSRouter(ctx, mark, localSubnet)
	if err != nil {
		return nil, err
	}

	r := &DNSReconciler{
		db:     db,
		router: router,
	}

	if _, err := r.check(ctx); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *DNSReconciler) check(ctx context.Context) (*database.DNSRoute, error) {
	route, err := r.db.DNS.Get(ctx)
	if route != database.EmptyRoute {
		net, err := r.db.Networks.Get(ctx, route.NetworkID)
		if err != nil {
			return nil, err
		}

		dockerNet, err := network.NewFromID(net.ID)
		if err != nil {
			return nil, err
		}

		if err := r.router.Route(dockerNet.Container.IPAddress); err != nil {
			return nil, err
		}
	}
	return route, err
}

func (r *DNSReconciler) Get(ctx context.Context) (*database.DNSRoute, error) {
	return r.check(ctx)
}

func (r *DNSReconciler) Create(ctx context.Context, networkID string) (*database.DNSRoute, error) {
	route, err := r.db.DNS.Put(ctx, &database.DNSRoute{
		NetworkID: networkID,
	})
	if err != nil {
		return nil, err
	}

	if _, err := r.check(ctx); err != nil {
		// TODO: clean up database
		return nil, err

	}

	return route, nil
}

func (r *DNSReconciler) Delete(ctx context.Context) error {
	if err := r.db.DNS.Delete(ctx); err != nil {
		return err
	}

	if err := r.router.Clear(); err != nil {
		return err
	}
	return nil
}
