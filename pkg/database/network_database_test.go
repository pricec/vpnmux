package database_test

import (
	"context"
	"os"
	"testing"

	"github.com/pricec/vpnmux/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestUpdateNetwork(t *testing.T) {
	ctx := context.Background()

	f, err := os.CreateTemp("", "")
	assert.Nil(t, err, "unexpected error creating temporary file")
	f.Close()
	defer os.Remove(f.Name())

	db, err := database.New(ctx, f.Name())
	assert.Nil(t, err, "unexpected error creating database")
	assert.NotNil(t, db, "unexpected nil database returned")

	net, err := db.Networks.Put(ctx, &database.Network{
		Name:     "test",
		ConfigID: "blah",
	})
	assert.Nil(t, err)
	assert.Equal(t, "test", net.Name)

	net.Name = "name2"

	err = db.Networks.Update(ctx, net)
	assert.Nil(t, err)

	net, err = db.Networks.Get(ctx, net.ID)
	assert.Nil(t, err)
	assert.Equal(t, "name2", net.Name)
}

func TestNetworks(t *testing.T) {
	ctx := context.Background()

	f, err := os.CreateTemp("", "")
	assert.Nil(t, err, "unexpected error creating temporary file")
	f.Close()
	defer os.Remove(f.Name())

	db, err := database.New(ctx, f.Name())
	assert.Nil(t, err, "unexpected error creating database")
	assert.NotNil(t, db, "unexpected nil database returned")

	nets, err := db.Networks.List(ctx)
	assert.Nil(t, err)
	assert.Empty(t, nets)

	net, err := db.Networks.Put(ctx, &database.Network{
		Name:     "test",
		ConfigID: "blah",
	})
	assert.Nil(t, err)
	assert.NotNil(t, net)

	net2, err := db.Networks.Get(ctx, "test")
	assert.Equal(t, database.ErrNotFound, err)
	assert.Nil(t, net2)

	net, err = db.Networks.Get(ctx, net.ID)
	assert.Nil(t, err)
	assert.NotNil(t, net)

	nets, err = db.Networks.List(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(nets))

	err = db.Networks.Delete(ctx, net.ID)
	assert.Nil(t, err)

	nets, err = db.Networks.List(ctx)
	assert.Nil(t, err)
	assert.Empty(t, nets)
}
