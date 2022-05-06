package database_test

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/pricec/vpnmux/pkg/database"
)

type Harness struct {
	DBPath   string
	DB       *database.Database
	Configs  []*database.Config
	Networks []*database.Network
	Clients  []*database.Client
}

type HarnessOptions struct {
	NumClients  int
	NumNetworks int
}

func NewHarness(ctx context.Context, opts HarnessOptions) (*Harness, error) {
	h := &Harness{}
	err := h.populate(ctx, opts)
	if err != nil {
		h.Close()
		return nil, err
	}
	return h, nil
}

func (h *Harness) populate(ctx context.Context, opts HarnessOptions) error {
	f, err := os.CreateTemp("", "")
	if err != nil {
		return err
	}
	h.DBPath = f.Name()
	f.Close()

	h.DB, err = database.New(ctx, h.DBPath)
	if err != nil {
		return err
	}

	h.Configs = make([]*database.Config, opts.NumNetworks)
	h.Networks = make([]*database.Network, opts.NumNetworks)
	for i := 0; i < opts.NumNetworks; i += 1 {
		// 4 credentials
		credNames := []string{"user", "pass", "ca", "tls"}
		creds := make(map[string]string)
		for _, name := range credNames {
			value := uuid.New().String()

			c, err := h.DB.Credentials.Put(ctx, name, value)
			if err != nil {
				return err
			}
			creds[name] = c.ID
		}

		// 1 config
		// TODO: random address
		h.Configs[i], err = h.DB.Configs.Put(ctx, &database.Config{
			Name:     fmt.Sprintf("Test Config %d", i),
			Host:     "1.1.1.1",
			UserCred: creds["user"],
			PassCred: creds["pass"],
			CACred:   creds["ca"],
			OVPNCred: creds["tls"],
		})
		if err != nil {
			return err
		}

		// 1 network
		h.Networks[i], err = h.DB.Networks.Put(ctx, &database.Network{
			Name:     fmt.Sprintf("Test Network %d", i),
			ConfigID: h.Configs[i].ID,
		})
		if err != nil {
			return err
		}
	}

	h.Clients = make([]*database.Client, opts.NumClients)
	for i := 0; i < opts.NumClients; i += 1 {
		// 1 client
		// TODO: random address
		h.Clients[i], err = h.DB.Clients.Put(ctx, &database.Client{
			Name:    fmt.Sprintf("Test Client %d", i),
			Address: "1.2.3.4",
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *Harness) Close() error {
	var result error

	if h.DBPath != "" {
		result = multierror.Append(result, os.Remove(h.DBPath))
	}

	return result
}
