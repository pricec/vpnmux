package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type ClientDatabase struct {
	db *sql.DB
}

type Client struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

func (d *ClientDatabase) List(ctx context.Context) ([]*Client, error) {
	rows, err := d.db.QueryContext(ctx, "SELECT id, name, address FROM client")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients = make([]*Client, 0)
	for rows.Next() {
		client := &Client{}
		if err := rows.Scan(&client.ID, &client.Name, &client.Address); err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	return clients, nil
}

func (d *ClientDatabase) Get(ctx context.Context, id string) (*Client, error) {
	row := d.db.QueryRowContext(ctx, "SELECT id, name, address FROM client WHERE id = ?", id)
	client := &Client{}
	err := row.Scan(&client.ID, &client.Name, &client.Address)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ErrNotFound
	case err == nil:
		return client, nil
	default:
		return nil, err
	}
}

func (d *ClientDatabase) Put(ctx context.Context, client *Client) (*Client, error) {
	id := uuid.New()
	_, err := d.db.ExecContext(ctx, "INSERT INTO client(id, name, address) VALUES(?, ?, ?)", id, client.Name, client.Address)
	if err != nil {
		return nil, err
	}

	client.ID = id.String()
	return client, nil
}

func (d *ClientDatabase) Update(ctx context.Context, client *Client) error {
	result, err := d.db.ExecContext(ctx, "UPDATE client SET name = ?, address = ? WHERE id = ?", client.Name, client.Address, client.ID)
	if err == nil {
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows != int64(1) {
			return ErrNotFound
		}
	}
	return err
}

func (d *ClientDatabase) Delete(ctx context.Context, id string) error {
	result, err := d.db.ExecContext(ctx, "DELETE FROM client WHERE id = ?", id)
	if err == nil {
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows != int64(1) {
			return ErrNotFound
		}
	}
	return err
}
