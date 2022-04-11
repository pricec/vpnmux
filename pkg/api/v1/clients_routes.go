package v1

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pricec/vpnmux/pkg/database"
)

func (m *Manager) ListClients(w http.ResponseWriter, r *http.Request) {
	creds, err := m.db.Clients.List(r.Context())
	check(w, creds, err, ErrorDatabase)
}

func (m *Manager) CreateClient(w http.ResponseWriter, r *http.Request) {
	c := &database.Client{}
	if err := json.NewDecoder(r.Body).Decode(c); err != nil {
		check(w, nil, err, errDecode(err))
		return
	}

	cfg, err := m.rec.Clients.Create(r.Context(), c)
	check(w, cfg, err, ErrorDatabase)
}

func (m *Manager) GetClient(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var alt Error
	client, _, err := m.rec.Clients.Get(r.Context(), id)
	switch err {
	case database.ErrNotFound:
		alt = ErrorNotFound
	default:
		alt = ErrorDatabase
	}
	check(w, client, err, alt)
}

func (m *Manager) UpdateClient(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	c := &database.Client{}
	if err := json.NewDecoder(r.Body).Decode(c); err != nil {
		check(w, nil, err, errDecode(err))
		return
	}
	c.ID = id

	var alt Error
	client, err := m.rec.Clients.Update(r.Context(), c)
	switch err {
	case database.ErrNotFound:
		alt = ErrorNotFound
	default:
		alt = ErrorDatabase
	}
	check(w, client, err, alt)
}

func (m *Manager) DeleteClient(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var alt Error
	err := m.rec.Clients.Delete(r.Context(), id)
	switch err {
	case database.ErrNotFound:
		alt = ErrorNotFound
	default:
		alt = ErrorDatabase
	}
	check(w, ErrorOK, err, alt)
}

func (m *Manager) GetClientNetwork(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var alt Error
	cn, _, err := m.rec.ClientNetworks.Get(r.Context(), id)
	switch err {
	case database.ErrNotFound:
		alt = ErrorNotFound
	default:
		alt = ErrorDatabase
	}
	check(w, cn, err, alt)
}

func (m *Manager) SetClientNetwork(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	network := mux.Vars(r)["network"]

	var alt Error
	cn, err := m.rec.ClientNetworks.Create(r.Context(), &database.ClientNetwork{
		ClientID:  id,
		NetworkID: network,
	})
	switch err {
	case database.ErrNotFound:
		alt = ErrorNotFound
	default:
		alt = ErrorDatabase
	}
	check(w, cn, err, alt)
}

func (m *Manager) UnsetClientNetwork(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var alt Error
	err := m.rec.ClientNetworks.Delete(r.Context(), id)
	switch err {
	case database.ErrNotFound:
		alt = ErrorNotFound
	default:
		alt = ErrorDatabase
	}
	check(w, ErrorOK, err, alt)
}
