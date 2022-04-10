package database_test

import (
	"context"
	"os"
	"testing"

	"github.com/pricec/vpnmux/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestUpdateClientNetwork(t *testing.T) {
	ctx := context.Background()

	f, err := os.CreateTemp("", "")
	assert.Nil(t, err, "unexpected error creating temporary file")
	f.Close()
	defer os.Remove(f.Name())

	db, err := database.New(ctx, f.Name())
	assert.Nil(t, err, "unexpected error creating database")
	assert.NotNil(t, db, "unexpected nil database returned")

	c, err := db.ClientNetworks.Put(ctx, &database.ClientNetwork{
		ClientID:  "client",
		NetworkID: "network",
	})
	assert.Nil(t, err)
	assert.Equal(t, "client", c.ClientID)
	assert.Equal(t, "network", c.NetworkID)

	c.NetworkID = "network2"

	err = db.ClientNetworks.Update(ctx, c)
	assert.Nil(t, err)

	c, err = db.ClientNetworks.Get(ctx, c.ClientID)
	assert.Nil(t, err)
	assert.Equal(t, "client", c.ClientID)
	assert.Equal(t, "network2", c.NetworkID)
}

func TestClientNetwork(t *testing.T) {
	ctx := context.Background()

	f, err := os.CreateTemp("", "")
	assert.Nil(t, err, "unexpected error creating temporary file")
	f.Close()
	defer os.Remove(f.Name())

	db, err := database.New(ctx, f.Name())
	assert.Nil(t, err, "unexpected error creating database")
	assert.NotNil(t, db, "unexpected nil database returned")

	cns, err := db.ClientNetworks.List(ctx)
	assert.Nil(t, err)
	assert.Empty(t, cns)

	cn, err := db.ClientNetworks.Put(ctx, &database.ClientNetwork{
		ClientID:  "client",
		NetworkID: "network",
	})
	assert.Nil(t, err)
	assert.NotNil(t, cn)

	c, err := db.ClientNetworks.Get(ctx, "test")
	assert.Equal(t, database.ErrNotFound, err)
	assert.Nil(t, c)

	c, err = db.ClientNetworks.Get(ctx, cn.ClientID)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	cns, err = db.ClientNetworks.List(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(cns))

	err = db.ClientNetworks.Delete(ctx, cn.ClientID)
	assert.Nil(t, err)

	cns, err = db.ClientNetworks.List(ctx)
	assert.Nil(t, err)
	assert.Empty(t, cns)
}
