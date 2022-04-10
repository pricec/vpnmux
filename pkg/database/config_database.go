package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type ConfigDatabase struct {
	db *sql.DB
}

type Config struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	UserCred string `json:"user_cred,omitempty"`
	PassCred string `json:"pass_cred,omitempty"`
	CACred   string `json:"ca_cred,omitempty"`
	OVPNCred string `json:"ovpn_cred,omitempty"`
}

func (d *ConfigDatabase) List(ctx context.Context) ([]*Config, error) {
	rows, err := d.db.QueryContext(ctx, "SELECT id, name, host FROM config")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs = make([]*Config, 0)
	for rows.Next() {
		cfg := &Config{}
		if err := rows.Scan(&cfg.ID, &cfg.Name, &cfg.Host); err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}

func (d *ConfigDatabase) Get(ctx context.Context, id string) (*Config, error) {
	row := d.db.QueryRowContext(ctx, "SELECT id, name, host, user_c, pass_c, ca_c, ovpn_c FROM config WHERE id = ?", id)
	cfg := &Config{}
	err := row.Scan(&cfg.ID, &cfg.Name, &cfg.Host, &cfg.UserCred, &cfg.PassCred, &cfg.CACred, &cfg.OVPNCred)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ErrNotFound
	case err == nil:
		return cfg, nil
	default:
		return nil, err
	}
}

func (d *ConfigDatabase) Put(ctx context.Context, cfg *Config) (*Config, error) {
	id := uuid.New()
	_, err := d.db.ExecContext(ctx, "INSERT INTO config(id, name, host, user_c, pass_c, ca_c, ovpn_c) VALUES(?, ?, ?, ?, ?, ?, ?)", id, cfg.Name, cfg.Host, cfg.UserCred, cfg.PassCred, cfg.CACred, cfg.OVPNCred)
	if err != nil {
		return nil, err
	}

	cfg.ID = id.String()
	return cfg, nil
}

func (d *ConfigDatabase) Update(ctx context.Context, cfg *Config) error {
	result, err := d.db.ExecContext(ctx, "UPDATE config SET name = ?, host = ?, user_c = ?, pass_c = ?, ca_c = ?, ovpn_c = ? WHERE id = ?", cfg.Name, cfg.Host, cfg.UserCred, cfg.PassCred, cfg.CACred, cfg.OVPNCred, cfg.ID)
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

func (d *ConfigDatabase) Delete(ctx context.Context, id string) error {
	result, err := d.db.ExecContext(ctx, "DELETE FROM config WHERE id = ?", id)
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
