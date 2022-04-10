package database

import (
	"context"
	"database/sql"
	"errors"
)

type ClientNetworkDatabase struct {
	db *sql.DB
}

type ClientNetwork struct {
	ClientID  string `json:"client_id"`
	NetworkID string `json:"network_id"`
}

func (d *ClientNetworkDatabase) List(ctx context.Context) ([]*ClientNetwork, error) {
	rows, err := d.db.QueryContext(ctx, "SELECT client_id, network_id FROM client_network")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cns = make([]*ClientNetwork, 0)
	for rows.Next() {
		cn := &ClientNetwork{}
		if err := rows.Scan(&cn.ClientID, &cn.NetworkID); err != nil {
			return nil, err
		}
		cns = append(cns, cn)
	}
	return cns, nil
}

func (d *ClientNetworkDatabase) Get(ctx context.Context, id string) (*ClientNetwork, error) {
	row := d.db.QueryRowContext(ctx, "SELECT client_id, network_id FROM client_network WHERE client_id = ?", id)
	cn := &ClientNetwork{}
	err := row.Scan(&cn.ClientID, &cn.NetworkID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ErrNotFound
	case err == nil:
		return cn, nil
	default:
		return nil, err
	}
}

func (d *ClientNetworkDatabase) Put(ctx context.Context, cn *ClientNetwork) (*ClientNetwork, error) {
	_, err := d.db.ExecContext(ctx, "INSERT INTO client_network(client_id, network_id) VALUES(?, ?)", cn.ClientID, cn.NetworkID)
	if err != nil {
		return nil, err
	}
	return cn, nil
}

func (d *ClientNetworkDatabase) Update(ctx context.Context, cn *ClientNetwork) error {
	result, err := d.db.ExecContext(ctx, "UPDATE client_network SET client_id = ?, network_id = ? WHERE client_id = ?", cn.ClientID, cn.NetworkID, cn.ClientID)
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

func (d *ClientNetworkDatabase) Delete(ctx context.Context, id string) error {
	result, err := d.db.ExecContext(ctx, "DELETE FROM client_network WHERE client_id = ?", id)
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
