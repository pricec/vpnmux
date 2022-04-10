package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type NetworkDatabase struct {
	db *sql.DB
}

type Network struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ConfigID string `json:"config_id"`
}

func (d *NetworkDatabase) List(ctx context.Context) ([]*Network, error) {
	rows, err := d.db.QueryContext(ctx, "SELECT id, name, config FROM network")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var networks = make([]*Network, 0)
	for rows.Next() {
		net := &Network{}
		if err := rows.Scan(&net.ID, &net.Name, &net.ConfigID); err != nil {
			return nil, err
		}
		networks = append(networks, net)
	}
	return networks, nil
}

func (d *NetworkDatabase) Get(ctx context.Context, id string) (*Network, error) {
	row := d.db.QueryRowContext(ctx, "SELECT id, name, config FROM network WHERE id = ?", id)
	net := &Network{}
	err := row.Scan(&net.ID, &net.Name, &net.ConfigID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ErrNotFound
	case err == nil:
		return net, nil
	default:
		return nil, err
	}
}

func (d *NetworkDatabase) Put(ctx context.Context, net *Network) (*Network, error) {
	id := uuid.New()
	_, err := d.db.ExecContext(ctx, "INSERT INTO network(id, name, config) VALUES(?, ?, ?)", id, net.Name, net.ConfigID)
	if err != nil {
		return nil, err
	}

	net.ID = id.String()
	return net, nil
}

func (d *NetworkDatabase) Update(ctx context.Context, net *Network) error {
	result, err := d.db.ExecContext(ctx, "UPDATE network SET name = ?, config = ? WHERE id = ?", net.Name, net.ConfigID, net.ID)
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

func (d *NetworkDatabase) Delete(ctx context.Context, id string) error {
	result, err := d.db.ExecContext(ctx, "DELETE FROM network WHERE id = ?", id)
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
