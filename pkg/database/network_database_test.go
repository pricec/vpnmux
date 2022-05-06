package database_test

import (
	"context"
	"testing"

	"github.com/pricec/vpnmux/pkg/database"
	"github.com/stretchr/testify/require"
)

func TestUpdateNetwork(t *testing.T) {
	ctx := context.Background()
	h, err := NewHarness(ctx, HarnessOptions{
		NumNetworks: 1,
	})
	require.Nil(t, err)

	net, err := h.DB.Networks.Get(ctx, h.Networks[0].ID)
	require.Nil(t, err)
	require.NotNil(t, net)
	net.Name = "name2"

	err = h.DB.Networks.Update(ctx, net)
	require.Nil(t, err)

	net, err = h.DB.Networks.Get(ctx, net.ID)
	require.Nil(t, err)
	require.Equal(t, "name2", net.Name)
}

func TestNetworks(t *testing.T) {
	ctx := context.Background()
	h, err := NewHarness(ctx, HarnessOptions{
		NumNetworks: 1,
	})
	require.Nil(t, err)

	nets, err := h.DB.Networks.List(ctx)
	require.Nil(t, err)
	require.Equal(t, 1, len(nets))

	net2, err := h.DB.Networks.Get(ctx, "test")
	require.Equal(t, database.ErrNotFound, err)
	require.Nil(t, net2)

	net, err := h.DB.Networks.Get(ctx, h.Networks[0].ID)
	require.Nil(t, err)
	require.NotNil(t, net)

	nets, err = h.DB.Networks.List(ctx)
	require.Nil(t, err)
	require.Equal(t, 1, len(nets))

	err = h.DB.Networks.Delete(ctx, net.ID)
	require.Nil(t, err)

	nets, err = h.DB.Networks.List(ctx)
	require.Nil(t, err)
	require.Empty(t, nets)
}
