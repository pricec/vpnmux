package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type CredentialDatabase struct {
	db *sql.DB
}

type Credential struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}

func (d *CredentialDatabase) List(ctx context.Context) ([]*Credential, error) {
	rows, err := d.db.QueryContext(ctx, "SELECT id, name FROM credential")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credentials = make([]*Credential, 0)
	for rows.Next() {
		cred := &Credential{}
		if err := rows.Scan(&cred.ID, &cred.Name); err != nil {
			return nil, err
		}
		credentials = append(credentials, cred)
	}
	return credentials, nil
}

func (d *CredentialDatabase) Get(ctx context.Context, id string) (*Credential, error) {
	row := d.db.QueryRowContext(ctx, "SELECT id, name, value FROM credential WHERE id = ?", id)
	cred := &Credential{}
	err := row.Scan(&cred.ID, &cred.Name, &cred.Value)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, ErrNotFound
	case err == nil:
		return cred, nil
	default:
		return nil, err
	}
}

func (d *CredentialDatabase) Put(ctx context.Context, name, value string) (*Credential, error) {
	id := uuid.New()
	_, err := d.db.ExecContext(ctx, "INSERT INTO credential(id, name, value) VALUES(?, ?, ?)", id, name, value)
	if err != nil {
		return nil, err
	}
	return &Credential{
		ID:    id.String(),
		Name:  name,
		Value: value,
	}, nil
}

func (d *CredentialDatabase) Update(ctx context.Context, cred *Credential) error {
	result, err := d.db.ExecContext(ctx, "UPDATE credential SET name = ?, value = ? WHERE id = ?", cred.Name, cred.Value, cred.ID)
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

func (d *CredentialDatabase) Delete(ctx context.Context, id string) error {
	result, err := d.db.ExecContext(ctx, "DELETE FROM credential WHERE id = ?", id)
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
