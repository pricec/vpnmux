package database

import (
	"context"
	"database/sql"
	"errors"
)

type DNSDatabase struct {
	db *sql.DB
}

type DNSRoute struct {
	ID        string `json:"-"`
	NetworkID string `json:"network_id"`
}

var EmptyRoute = &DNSRoute{
	ID:        "0",
	NetworkID: "none",
}

func (d *DNSDatabase) Get(ctx context.Context) (*DNSRoute, error) {
	row := d.db.QueryRowContext(ctx, "SELECT id, network_id FROM dns_route WHERE id = 0")
	route := &DNSRoute{}
	err := row.Scan(&route.ID, &route.NetworkID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return EmptyRoute, nil
	case err == nil:
		return route, nil
	default:
		return nil, err
	}
}

func (d *DNSDatabase) Put(ctx context.Context, route *DNSRoute) (*DNSRoute, error) {
	_, err := d.db.ExecContext(ctx, "INSERT INTO dns_route(id, network_id) VALUES(0, ?)", route.NetworkID)
	if err != nil {
		return nil, err
	}
	route.ID = "0"
	return route, nil
}

func (d *DNSDatabase) Update(ctx context.Context, route *DNSRoute) error {
	result, err := d.db.ExecContext(ctx, "UPDATE dns_route SET network_id = ? WHERE id = 0", route.NetworkID)
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

func (d *DNSDatabase) Delete(ctx context.Context) error {
	result, err := d.db.ExecContext(ctx, "DELETE FROM dns_route WHERE id = 0")
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
