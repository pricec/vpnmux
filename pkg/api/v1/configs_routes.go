package v1

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pricec/vpnmux/pkg/database"
)

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

	cfg, err := m.rec.Configs.Create(r.Context(), c)
	check(w, cfg, err, ErrorDatabase)
}

func (m *Manager) GetConfig(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var alt Error
	cfg, _, err := m.rec.Configs.Get(r.Context(), id)
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
	cfg, err := m.rec.Configs.Update(r.Context(), c)
	switch err {
	case database.ErrNotFound:
		alt = ErrorNotFound
	default:
		alt = ErrorDatabase
	}
	check(w, cfg, err, alt)
}

func (m *Manager) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var alt Error
	err := m.rec.Configs.Delete(r.Context(), id)
	switch err {
	case database.ErrNotFound:
		alt = ErrorNotFound
	default:
		alt = ErrorDatabase
	}
	check(w, ErrorOK, err, alt)
}
