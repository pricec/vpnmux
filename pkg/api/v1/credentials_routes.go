package v1

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pricec/vpnmux/pkg/database"
)

func (m *Manager) ListCredentials(w http.ResponseWriter, r *http.Request) {
	creds, err := m.db.Credentials.List(r.Context())
	check(w, creds, err, ErrorDatabase)
}

func (m *Manager) CreateCredential(w http.ResponseWriter, r *http.Request) {
	c := &database.Credential{}
	if err := json.NewDecoder(r.Body).Decode(c); err != nil {
		check(w, nil, err, errDecode(err))
		return
	}
	cred, err := m.db.Credentials.Put(r.Context(), c.Name, c.Value)
	check(w, cred, err, ErrorDatabase)
}

func (m *Manager) GetCredential(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var alt Error
	cred, err := m.db.Credentials.Get(r.Context(), id)
	switch err {
	case database.ErrNotFound:
		alt = ErrorNotFound
	default:
		alt = ErrorDatabase
	}
	check(w, cred, err, alt)
}

func (m *Manager) UpdateCredential(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	c := &database.Credential{}
	if err := json.NewDecoder(r.Body).Decode(c); err != nil {
		check(w, nil, err, errDecode(err))
		return
	}
	c.ID = id

	var alt Error
	err := m.db.Credentials.Update(r.Context(), c)
	switch err {
	case database.ErrNotFound:
		alt = ErrorNotFound
	default:
		alt = ErrorDatabase
	}
	check(w, ErrorOK, err, alt)
}

func (m *Manager) DeleteCredential(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var alt Error
	err := m.db.Credentials.Delete(r.Context(), id)
	switch err {
	case database.ErrNotFound:
		alt = ErrorNotFound
	default:
		alt = ErrorDatabase
	}
	check(w, ErrorOK, err, alt)
}
