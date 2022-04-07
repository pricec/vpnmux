package network

import (
	"database/sql"
	"os"
	"path"

	_ "modernc.org/sqlite"
)

var schema = []string{
	`
CREATE TABLE IF NOT EXISTS client (
  address VARCHAR(64) PRIMARY KEY,
  network VARCHAR(128) NULL
);
`,
}

const dbPath = "/var/lib/vpnmux/network.db"

type Database struct {
	path string
	db   *sql.DB
}

func NewDatabase() (*Database, error) {
	err := os.MkdirAll(path.Dir(dbPath), 0700)
	if err != nil {
		return nil, err
	}

	d := &Database{
		path: dbPath,
	}
	return d, d.setup()
}

func (d *Database) setup() error {
	var err error
	d.db, err = sql.Open("sqlite", d.path)
	if err != nil {
		return err
	}

	for _, createTable := range schema {
		_, err = d.db.Exec(createTable)
		if err != nil {
			d.db.Close()
			return err
		}
	}
	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

type Client struct {
	Address string
	Network string
}

func (d *Database) GetClients() ([]Client, error) {
	rows, err := d.db.Query("SELECT address, network FROM client;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients = make([]Client, 0)
	for rows.Next() {
		var addr, network sql.NullString
		if err := rows.Scan(&addr, &network); err != nil {
			return nil, err
		}
		clients = append(clients, Client{
			Address: addr.String,
			Network: network.String,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return clients, nil
}

func (d *Database) PutClient(address string) error {
	_, err := d.db.Exec("INSERT INTO client(address) VALUES(?);", address)
	return err
}

func (d *Database) SetClientNetwork(address string, network string) error {
	_, err := d.db.Exec("UPDATE client SET network = ? WHERE address = ?;", network, address)
	return err
}

func (d *Database) DeleteClient(address string) error {
	_, err := d.db.Exec("DELETE FROM client WHERE address = ?;", address)
	return err
}
