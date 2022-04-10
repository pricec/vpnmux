package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pricec/vpnmux/pkg/database"
	"github.com/pricec/vpnmux/pkg/openvpn"
)

func (m *Manager) WriteConfig(ctx context.Context, id string) error {
	cfg, err := m.db.Configs.Get(ctx, id)
	if err != nil {
		return err
	}

	userCred, err := m.db.Credentials.Get(ctx, cfg.UserCred)
	if err != nil {
		return err
	}

	passCred, err := m.db.Credentials.Get(ctx, cfg.PassCred)
	if err != nil {
		return err
	}

	caCred, err := m.db.Credentials.Get(ctx, cfg.CACred)
	if err != nil {
		return err
	}

	ovpnCred, err := m.db.Credentials.Get(ctx, cfg.OVPNCred)
	if err != nil {
		return err
	}

	_, err = openvpn.NewConfig2(fmt.Sprintf("/var/lib/vpnmux/openvpn/%s", cfg.ID), openvpn.ConfigOptions{
		Host:    cfg.Host,
		User:    userCred.Value,
		Pass:    passCred.Value,
		CACert:  caCred.Value,
		TLSCert: ovpnCred.Value,
	})
	return err
}

func (m *Manager) ListConfigs(w http.ResponseWriter, r *http.Request) {
	creds, err := m.db.Configs.List(r.Context())
	check(w, creds, err, ErrorDatabase)
}

func (m *Manager) CreateConfig(w http.ResponseWriter, r *http.Request) {
	c := &database.Config{}
	if err := json.NewDecoder(r.Body).Decode(c); err != nil {
		check(w, nil, err, errDecode(err))
		return
	}
	cfg, err := m.db.Configs.Put(r.Context(), c)
	check(w, cfg, err, ErrorDatabase)

	m.WriteConfig(r.Context(), cfg.ID)
}

func (m *Manager) GetConfig(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var alt Error
	cfg, err := m.db.Configs.Get(r.Context(), id)
	switch err {
	case database.ErrNotFound:
		alt = ErrorNotFound
	default:
		alt = ErrorDatabase
	}
	check(w, cfg, err, alt)
}

func (m *Manager) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	c := &database.Config{}
	if err := json.NewDecoder(r.Body).Decode(c); err != nil {
		check(w, nil, err, errDecode(err))
		return
	}
	c.ID = id

	var alt Error
	err := m.db.Configs.Update(r.Context(), c)
	switch err {
	case database.ErrNotFound:
		alt = ErrorNotFound
	default:
		alt = ErrorDatabase
	}
	check(w, ErrorOK, err, alt)
}

func (m *Manager) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var alt Error
	err := m.db.Configs.Delete(r.Context(), id)
	switch err {
	case database.ErrNotFound:
		alt = ErrorNotFound
	default:
		alt = ErrorDatabase
	}
	check(w, ErrorOK, err, alt)
}
