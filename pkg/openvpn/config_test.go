package openvpn_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/pricec/vpnmux/pkg/openvpn"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	id := uuid.New().String()
	cfg, err := openvpn.NewConfig(id, openvpn.ConfigOptions{
		Host:    "host",
		User:    "username",
		Pass:    "password",
		CACert:  "CA",
		TLSCert: "openvpn private key",
	})
	assert.Nilf(t, err, "unexpected error: %v", err)
	assert.NotNil(t, cfg)

	cfg2, err := openvpn.NewConfigFromID(id)
	assert.Nil(t, err)
	assert.NotNil(t, cfg2)

	assert.Equal(t, cfg.ID, cfg2.ID)
	assert.Equal(t, cfg.Dir, cfg2.Dir)
	assert.Nil(t, cfg.Close())
}
