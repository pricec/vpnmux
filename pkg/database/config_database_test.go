package database_test

import (
	"context"
	"os"
	"testing"

	"github.com/pricec/vpnmux/pkg/database"
	"github.com/stretchr/testify/require"
)

func TestUpdateConfig(t *testing.T) {
	ctx := context.Background()

	f, err := os.CreateTemp("", "")
	require.Nil(t, err, "unexpected error creating temporary file")
	f.Close()
	defer os.Remove(f.Name())

	db, err := database.New(ctx, f.Name())
	require.Nil(t, err, "unexpected error creating database")
	require.NotNil(t, db, "unexpected nil database returned")

	cred, err := db.Credentials.Put(ctx, "name", "value")
	require.Nil(t, err)

	c, err := db.Configs.Put(ctx, &database.Config{
		Name:     "test",
		Host:     "test_host",
		UserCred: cred.ID,
		PassCred: cred.ID,
		CACred:   cred.ID,
		OVPNCred: cred.ID,
	})
	require.Nil(t, err)
	require.Equal(t, "test", c.Name)
	require.Equal(t, "test_host", c.Host)

	c.Name = "test2"
	c.Host = "test_host2"
	cred2, err := db.Credentials.Put(ctx, "name", "value")
	require.Nil(t, err)
	c.UserCred = cred2.ID
	c.PassCred = cred2.ID

	err = db.Configs.Update(ctx, c)
	require.Nil(t, err)

	c, err = db.Configs.Get(ctx, c.ID)
	require.Nil(t, err)
	require.Equal(t, "test2", c.Name)
	require.Equal(t, "test_host2", c.Host)
	require.Equal(t, cred2.ID, c.UserCred)
	require.Equal(t, cred2.ID, c.PassCred)
	require.Equal(t, cred.ID, c.CACred)
	require.Equal(t, cred.ID, c.OVPNCred)
}

func TestConfig(t *testing.T) {
	ctx := context.Background()
	h, err := NewHarness(ctx, HarnessOptions{
		NumClients:  0,
		NumNetworks: 1,
	})
	require.Nil(t, err)
	defer h.Close()

	c, err := h.DB.Configs.Get(ctx, "test")
	require.Equal(t, database.ErrNotFound, err)
	require.Nil(t, c)

	c, err = h.DB.Configs.Get(ctx, h.Configs[0].ID)
	require.Nil(t, err)
	require.NotNil(t, c)

	cfgs, err := h.DB.Configs.List(ctx)
	require.Nil(t, err)
	require.Equal(t, 1, len(cfgs))

	err = h.DB.Networks.Delete(ctx, h.Networks[0].ID)
	require.Nil(t, err)

	err = h.DB.Configs.Delete(ctx, h.Configs[0].ID)
	require.Nil(t, err)

	cfgs, err = h.DB.Configs.List(ctx)
	require.Nil(t, err)
	require.Empty(t, cfgs)
}
