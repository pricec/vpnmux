package database_test

import (
	"context"
	"testing"

	"github.com/pricec/vpnmux/pkg/database"
	"github.com/stretchr/testify/require"
)

func TestUpdateClientNetwork(t *testing.T) {
	ctx := context.Background()

	h, err := NewHarness(ctx, HarnessOptions{
		NumClients:  1,
		NumNetworks: 2,
	})
	require.Nil(t, err)
	defer h.Close()

	c, err := h.DB.ClientNetworks.Put(ctx, &database.ClientNetwork{
		ClientID:  h.Clients[0].ID,
		NetworkID: h.Networks[0].ID,
	})
	require.Nil(t, err)
	require.Equal(t, h.Clients[0].ID, c.ClientID)
	require.Equal(t, h.Networks[0].ID, c.NetworkID)

	c.NetworkID = h.Networks[1].ID

	err = h.DB.ClientNetworks.Update(ctx, c)
	require.Nil(t, err)

	c, err = h.DB.ClientNetworks.Get(ctx, c.ClientID)
	require.Nil(t, err)
	require.Equal(t, h.Clients[0].ID, c.ClientID)
	require.Equal(t, h.Networks[1].ID, c.NetworkID)
}

func TestClientNetwork(t *testing.T) {
	ctx := context.Background()
	h, err := NewHarness(ctx, HarnessOptions{
		NumClients:  1,
		NumNetworks: 1,
	})
	require.Nil(t, err)
	defer h.Close()

	cns, err := h.DB.ClientNetworks.List(ctx)
	require.Nil(t, err)
	require.Empty(t, cns)

	cn, err := h.DB.ClientNetworks.Put(ctx, &database.ClientNetwork{
		ClientID:  h.Clients[0].ID,
		NetworkID: h.Networks[0].ID,
	})
	require.Nil(t, err)
	require.NotNil(t, cn)

	c, err := h.DB.ClientNetworks.Get(ctx, "test")
	require.Equal(t, database.ErrNotFound, err)
	require.Nil(t, c)

	c, err = h.DB.ClientNetworks.Get(ctx, cn.ClientID)
	require.Nil(t, err)
	require.NotNil(t, c)

	cns, err = h.DB.ClientNetworks.List(ctx)
	require.Nil(t, err)
	require.Equal(t, 1, len(cns))

	err = h.DB.ClientNetworks.Delete(ctx, cn.ClientID)
	require.Nil(t, err)

	cns, err = h.DB.ClientNetworks.List(ctx)
	require.Nil(t, err)
	require.Empty(t, cns)
}
