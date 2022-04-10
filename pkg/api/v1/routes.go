package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pricec/vpnmux/pkg/database"
	"github.com/pricec/vpnmux/pkg/network"
	"github.com/pricec/vpnmux/pkg/openvpn"
)

func RegisterHandlers(ctx context.Context, r *mux.Router, dbPath string) {
	db, err := database.New(ctx, dbPath)
	if err != nil {
		log.Panicf("error opening database: %v", err)
	}

	mgr := &Manager{
		db: db,
	}

	r.HandleFunc("/credential", mgr.ListCredentials).Methods("GET")
	r.HandleFunc("/credential", mgr.CreateCredential).Methods("POST")
	r.HandleFunc("/credential/{id}", mgr.GetCredential).Methods("GET")
	r.HandleFunc("/credential/{id}", mgr.UpdateCredential).Methods("PATCH")
	r.HandleFunc("/credential/{id}", mgr.DeleteCredential).Methods("DELETE")

	r.HandleFunc("/config", mgr.ListConfigs).Methods("GET")
	r.HandleFunc("/config", mgr.CreateConfig).Methods("POST")
	r.HandleFunc("/config/{id}", mgr.GetConfig).Methods("GET")
	r.HandleFunc("/config/{id}", mgr.UpdateConfig).Methods("PATCH")
	r.HandleFunc("/config/{id}", mgr.DeleteConfig).Methods("DELETE")

	r.HandleFunc("/network", mgr.ListNetworks).Methods("GET")
	r.HandleFunc("/network", mgr.CreateNetwork).Methods("POST")
	r.HandleFunc("/network/{id}", mgr.GetNetwork).Methods("GET")
	r.HandleFunc("/network/{id}", mgr.UpdateNetwork).Methods("PATCH")
	r.HandleFunc("/network/{id}", mgr.DeleteNetwork).Methods("DELETE")
}

type Manager struct {
	db *database.Database
}

type Error struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

var (
	ErrorOK       = Error{Code: http.StatusOK, Description: "OK"}
	ErrorNotFound = Error{Code: http.StatusNotFound, Description: "Resource not found"}
	ErrorDatabase = Error{Code: http.StatusInternalServerError, Description: "database error"}
)

func errDecode(err error) Error {
	return Error{
		Code:        http.StatusBadRequest,
		Description: err.Error(),
	}
}

// if err is not nil, log it and respond with alt. Otherwise, respond
// with result.
func check(w http.ResponseWriter, result interface{}, err error, alt Error) {
	if err != nil {
		log.Printf("error: %v", err)
		w.WriteHeader(alt.Code)
		if err := json.NewEncoder(w).Encode(alt); err != nil {
			log.Printf("error encoding response: %v", err)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("error encoding response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (m *Manager) Config(ctx context.Context, id string) (*openvpn.Config, error) {
	cfg, err := m.db.Configs.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	userCred, err := m.db.Credentials.Get(ctx, cfg.UserCred)
	if err != nil {
		return nil, err
	}

	passCred, err := m.db.Credentials.Get(ctx, cfg.PassCred)
	if err != nil {
		return nil, err
	}

	caCred, err := m.db.Credentials.Get(ctx, cfg.CACred)
	if err != nil {
		return nil, err
	}

	ovpnCred, err := m.db.Credentials.Get(ctx, cfg.OVPNCred)
	if err != nil {
		return nil, err
	}

	return openvpn.NewConfig2(fmt.Sprintf("/var/lib/vpnmux/openvpn/%s", cfg.ID), openvpn.ConfigOptions{
		Host:    cfg.Host,
		User:    userCred.Value,
		Pass:    passCred.Value,
		CACert:  caCred.Value,
		TLSCert: ovpnCred.Value,
	})
}

func (m *Manager) DeployNetwork(ctx context.Context, n *database.Network) error {
	cfg, err := m.Config(ctx, n.ConfigID)
	if err != nil {
		return err
	}

	// TODO: clean up network?
	net, err := network.New(n.ID)
	if err != nil {
		return err
	}

	_, err = network.NewContainer(net.Name, cfg)
	return err
}
