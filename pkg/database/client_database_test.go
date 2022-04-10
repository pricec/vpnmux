package database_test

import (
	"context"
	"os"
	"testing"

	"github.com/pricec/vpnmux/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestUpdateClient(t *testing.T) {
	ctx := context.Background()

	f, err := os.CreateTemp("", "")
	assert.Nil(t, err, "unexpected error creating temporary file")
	f.Close()
	defer os.Remove(f.Name())

	db, err := database.New(ctx, f.Name())
	assert.Nil(t, err, "unexpected error creating database")
	assert.NotNil(t, db, "unexpected nil database returned")

	client, err := db.Clients.Put(ctx, &database.Client{
		Name:    "test",
		Address: "blah",
	})
	assert.Nil(t, err)
	assert.Equal(t, "test", client.Name)
	assert.Equal(t, "blah", client.Address)

	client.Name = "name2"

	err = db.Clients.Update(ctx, client)
	assert.Nil(t, err)

	client, err = db.Clients.Get(ctx, client.ID)
	assert.Nil(t, err)
	assert.Equal(t, "name2", client.Name)
	assert.Equal(t, "blah", client.Address)
}

func TestClients(t *testing.T) {
	ctx := context.Background()

	f, err := os.CreateTemp("", "")
	assert.Nil(t, err, "unexpected error creating temporary file")
	f.Close()
	defer os.Remove(f.Name())

	db, err := database.New(ctx, f.Name())
	assert.Nil(t, err, "unexpected error creating database")
	assert.NotNil(t, db, "unexpected nil database returned")

	clients, err := db.Clients.List(ctx)
	assert.Nil(t, err)
	assert.Empty(t, clients)

	client, err := db.Clients.Put(ctx, &database.Client{
		Name:    "test",
		Address: "1.1.1.1",
	})
	assert.Nil(t, err)
	assert.NotNil(t, client)

	client2, err := db.Clients.Get(ctx, "test")
	assert.Equal(t, database.ErrNotFound, err)
	assert.Nil(t, client2)

	client, err = db.Clients.Get(ctx, client.ID)
	assert.Nil(t, err)
	assert.NotNil(t, client)

	clients, err = db.Clients.List(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(clients))

	err = db.Clients.Delete(ctx, client.ID)
	assert.Nil(t, err)

	clients, err = db.Clients.List(ctx)
	assert.Nil(t, err)
	assert.Empty(t, clients)
}
