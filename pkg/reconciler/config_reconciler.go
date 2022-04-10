package reconciler

import (
	"context"
	"fmt"

	"github.com/pricec/vpnmux/pkg/database"
	"github.com/pricec/vpnmux/pkg/openvpn"
)

type ConfigReconciler struct {
	db *database.Database
}

func (r *ConfigReconciler) Update(ctx context.Context, cfg *database.Config) (*database.Config, error) {
	return nil, fmt.Errorf("TODO: implement config update reconciler")
}

func NewConfigReconciler(ctx context.Context, db *database.Database) (*ConfigReconciler, error) {
	configs, err := db.Configs.List(ctx)
	if err != nil {
		return nil, err
	}

	r := &ConfigReconciler{db: db}
	for _, config := range configs {
		if _, _, err := r.check(ctx, config.ID); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *ConfigReconciler) check(ctx context.Context, id string) (*database.Config, *openvpn.Config, error) {
	cfg, err := r.db.Configs.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	diskCfg, err := openvpn.NewConfigFromID(id)
	if err != nil {
		return nil, nil, err
	}

	// TODO: check if settings are equal?
	return cfg, diskCfg, nil
}

func (r *ConfigReconciler) Get(ctx context.Context, id string) (*database.Config, *openvpn.Config, error) {
	return r.check(ctx, id)
}

func (r *ConfigReconciler) Create(ctx context.Context, c *database.Config) (*database.Config, error) {
	cfg, err := r.db.Configs.Put(ctx, c)
	if err != nil {
		return nil, err
	}

	userCred, err := r.db.Credentials.Get(ctx, cfg.UserCred)
	if err != nil {
		return nil, err
	}

	passCred, err := r.db.Credentials.Get(ctx, cfg.PassCred)
	if err != nil {
		return nil, err
	}

	caCred, err := r.db.Credentials.Get(ctx, cfg.CACred)
	if err != nil {
		return nil, err
	}

	ovpnCred, err := r.db.Credentials.Get(ctx, cfg.OVPNCred)
	if err != nil {
		return nil, err
	}

	_, err = openvpn.NewConfig(cfg.ID, openvpn.ConfigOptions{
		Host:    cfg.Host,
		User:    userCred.Value,
		Pass:    passCred.Value,
		CACert:  caCred.Value,
		TLSCert: ovpnCred.Value,
	})
	if err != nil {
		// TODO: clean up database
		return nil, err
	}

	if _, _, err := r.check(ctx, cfg.ID); err != nil {
		// TODO: clean up
		return nil, err

	}

	return cfg, nil
}

func (r *ConfigReconciler) Delete(ctx context.Context, id string) error {
	cfg, diskCfg, err := r.check(ctx, id)
	if err != nil {
		return err
	}

	if err := r.db.Configs.Delete(ctx, cfg.ID); err != nil {
		return err
	}

	if err := diskCfg.Close(); err != nil {
		return err
	}
	return nil
}
