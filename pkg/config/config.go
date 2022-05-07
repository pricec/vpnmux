package config

import (
	"time"

	env "github.com/caarlos0/env/v6"
)

type Config struct {
	DBPath          string        `env:"VPNMUX_DB_PATH" envDefault:"/var/lib/vpnmux/v1.db"`
	VPNImage        string        `env:"VPNMUX_IMAGE" envDefault:"pricec/openvpn-client"`
	LocalSubnetCIDR string        `env:"VPNMUX_SUBNET_CIDR,notEmpty"`
	ShutdownTimeout time.Duration `env:"VPNMUX_SHUTDOWN_TIMEOUT" envDefault:"10s"`
	ListenPort      uint16        `env:"VPNMUX_LISTEN_PORT" envDefault:"8080"`
}

func New() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
