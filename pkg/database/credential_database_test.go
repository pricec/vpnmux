package database_test

import (
	"context"
	"os"
	"testing"

	"github.com/pricec/vpnmux/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestUpdateCredential(t *testing.T) {
	ctx := context.Background()

	f, err := os.CreateTemp("", "")
	assert.Nil(t, err, "unexpected error creating temporary file")
	f.Close()
	defer os.Remove(f.Name())

	db, err := database.New(ctx, f.Name())
	assert.Nil(t, err, "unexpected error creating database")
	assert.NotNil(t, db, "unexpected nil database returned")

	c, err := db.Credentials.Put(ctx, "name", "value")
	assert.Nil(t, err)
	assert.Equal(t, "name", c.Name)
	assert.Equal(t, "value", c.Value)

	c.Name = "name2"
	c.Value = "value2"

	err = db.Credentials.Update(ctx, c)
	assert.Nil(t, err)

	c, err = db.Credentials.Get(ctx, c.ID)
	assert.Nil(t, err)
	assert.Equal(t, "name2", c.Name)
	assert.Equal(t, "value2", c.Value)
}

func TestCredentials(t *testing.T) {
	ctx := context.Background()

	f, err := os.CreateTemp("", "")
	assert.Nil(t, err, "unexpected error creating temporary file")
	f.Close()
	defer os.Remove(f.Name())

	db, err := database.New(ctx, f.Name())
	assert.Nil(t, err, "unexpected error creating database")
	assert.NotNil(t, db, "unexpected nil database returned")

	creds, err := db.Credentials.List(ctx)
	assert.Nil(t, err, "error listing credentials")
	assert.Empty(t, creds, "found credentials in an empty database")

	cred, err := db.Credentials.Put(ctx, "test", "Test")
	assert.Nil(t, err, "unexpected error creating credential")
	assert.NotNil(t, cred, "got nil credential from Put")

	c, err := db.Credentials.Get(ctx, "test")
	assert.Equal(t, database.ErrNotFound, err)
	assert.Nil(t, c, "expected nil cred on nonexistent cred")

	c, err = db.Credentials.Get(ctx, cred.ID)
	assert.Nil(t, err, "unexpected error getting cred")
	assert.NotNil(t, c, "expected non-nil cred")

	creds, err = db.Credentials.List(ctx)
	assert.Nil(t, err, "unexpected error listing credentials")
	assert.Equal(t, 1, len(creds), "unexpected number of credentials returned")

	err = db.Credentials.Delete(ctx, cred.ID)
	assert.Nil(t, err, "unexpected error deleting credential")

	creds, err = db.Credentials.List(ctx)
	assert.Nil(t, err, "error listing credentials")
	assert.Empty(t, creds, "found credentials in an empty database")
}
