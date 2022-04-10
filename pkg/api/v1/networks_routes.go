package v1

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pricec/vpnmux/pkg/database"
)

func (m *Manager) ListNetworks(w http.ResponseWriter, r *http.Request) {
	creds, err := m.db.Networks.List(r.Context())
	check(w, creds, err, ErrorDatabase)
}

func (m *Manager) CreateNetwork(w http.ResponseWriter, r *http.Request) {
	n := &database.Network{}
	if err := json.NewDecoder(r.Body).Decode(n); err != nil {
		check(w, nil, err, errDecode(err))
		return
	}
	net, err := m.db.Networks.Put(r.Context(), n)
	check(w, net, err, ErrorDatabase)

	log.Printf("error deploying network: %v", m.DeployNetwork(r.Context(), n))
}

func (m *Manager) GetNetwork(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var alt Error
	net, err := m.db.Networks.Get(r.Context(), id)
	switch err {
	case database.ErrNotFound:
		alt = ErrorNotFound
	default:
		alt = ErrorDatabase
	}
	check(w, net, err, alt)
}

func (m *Manager) UpdateNetwork(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	net := &database.Network{}
	if err := json.NewDecoder(r.Body).Decode(net); err != nil {
		check(w, nil, err, errDecode(err))
		return
	}
	net.ID = id

	var alt Error
	err := m.db.Networks.Update(r.Context(), net)
	switch err {
	case database.ErrNotFound:
		alt = ErrorNotFound
	default:
		alt = ErrorDatabase
	}
	check(w, ErrorOK, err, alt)
}

func (m *Manager) DeleteNetwork(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var alt Error
	err := m.db.Networks.Delete(r.Context(), id)
	switch err {
	case database.ErrNotFound:
		alt = ErrorNotFound
	default:
		alt = ErrorDatabase
	}
	check(w, ErrorOK, err, alt)
}
