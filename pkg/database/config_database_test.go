package database_test

import (
	"context"
	"os"
	"testing"

	"github.com/pricec/vpnmux/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestUpdateConfig(t *testing.T) {
	ctx := context.Background()

	f, err := os.CreateTemp("", "")
	assert.Nil(t, err, "unexpected error creating temporary file")
	f.Close()
	defer os.Remove(f.Name())

	db, err := database.New(ctx, f.Name())
	assert.Nil(t, err, "unexpected error creating database")
	assert.NotNil(t, db, "unexpected nil database returned")

	cred, err := db.Credentials.Put(ctx, "name", "value")
	assert.Nil(t, err)

	c, err := db.Configs.Put(ctx, &database.Config{
		Name:     "test",
		Host:     "test_host",
		UserCred: cred.ID,
		PassCred: cred.ID,
		CACred:   cred.ID,
		OVPNCred: cred.ID,
	})
	assert.Nil(t, err)
	assert.Equal(t, "test", c.Name)
	assert.Equal(t, "test_host", c.Host)

	c.Name = "test2"
	c.Host = "test_host2"
	cred2, err := db.Credentials.Put(ctx, "name", "value")
	assert.Nil(t, err)
	c.UserCred = cred2.ID
	c.PassCred = cred2.ID

	err = db.Configs.Update(ctx, c)
	assert.Nil(t, err)

	c, err = db.Configs.Get(ctx, c.ID)
	assert.Nil(t, err)
	assert.Equal(t, "test2", c.Name)
	assert.Equal(t, "test_host2", c.Host)
	assert.Equal(t, cred2.ID, c.UserCred)
	assert.Equal(t, cred2.ID, c.PassCred)
	assert.Equal(t, cred.ID, c.CACred)
	assert.Equal(t, cred.ID, c.OVPNCred)
}

func TestConfig(t *testing.T) {
	ctx := context.Background()

	f, err := os.CreateTemp("", "")
	assert.Nil(t, err, "unexpected error creating temporary file")
	f.Close()
	defer os.Remove(f.Name())

	db, err := database.New(ctx, f.Name())
	assert.Nil(t, err, "unexpected error creating database")
	assert.NotNil(t, db, "unexpected nil database returned")

	_, err = db.Credentials.Put(ctx, "name", "value")
	assert.Nil(t, err)

	cfgs, err := db.Configs.List(ctx)
	assert.Nil(t, err)
	assert.Empty(t, cfgs)

	cfg, err := db.Configs.Put(ctx, &database.Config{
		Name: "test",
		Host: "blah",
	})
	assert.Nil(t, err)
	assert.NotNil(t, cfg)

	c, err := db.Configs.Get(ctx, "test")
	assert.Equal(t, database.ErrNotFound, err)
	assert.Nil(t, c)

	c, err = db.Configs.Get(ctx, cfg.ID)
	assert.Nil(t, err)
	assert.NotNil(t, c)

	cfgs, err = db.Configs.List(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(cfgs))

	err = db.Configs.Delete(ctx, cfg.ID)
	assert.Nil(t, err, "unexpected error deleting credential")

	cfgs, err = db.Configs.List(ctx)
	assert.Nil(t, err, "error listing credentials")
	assert.Empty(t, cfgs, "found credentials in an empty database")
}
