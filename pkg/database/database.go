package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path"

	_ "modernc.org/sqlite"
)

var (
	ErrNotFound = fmt.Errorf("object not found")
)

type Database struct {
	db          *sql.DB
	Credentials *CredentialDatabase
	Configs     *ConfigDatabase
}

func New(ctx context.Context, dbPath string) (*Database, error) {
	if err := os.MkdirAll(path.Dir(dbPath), 0700); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	for _, statement := range schema {
		_, err = db.ExecContext(ctx, statement)
		if err != nil {
			log.Panicf("error applying schema: %v", err)
		}
	}

	return &Database{
		db: db,
		Credentials: &CredentialDatabase{
			db: db,
		},
		Configs: &ConfigDatabase{
			db: db,
		},
	}, nil
}