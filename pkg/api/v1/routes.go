package v1

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pricec/vpnmux/pkg/config"
	"github.com/pricec/vpnmux/pkg/database"
	"github.com/pricec/vpnmux/pkg/reconciler"
)

func RegisterHandlers(ctx context.Context, r *mux.Router, cfg *config.Config) {
	db, err := database.New(ctx, cfg.DBPath)
	if err != nil {
		log.Panicf("error opening database: %v", err)
	}

	rec, err := reconciler.New(ctx, reconciler.Options{
		DB: db,
		Network: reconciler.NetworkReconcilerOptions{
			VPNImage:        cfg.VPNImage,
			LocalSubnetCIDR: cfg.LocalSubnetCIDR,
		},
		Forwarding: reconciler.ForwardingOptions{
			LANInterface: cfg.LANInterface,
			WANInterface: cfg.WANInterface,
		},
	})
	if err != nil {
		log.Panicf("error creating reconciler: %v", err)
	}

	mgr := &Manager{
		db:  db,
		rec: rec,
	}

	// TODO: PATCH routes are currently disabled because of the cascading
	// impact of changes (credential -> config -> network).

	r.HandleFunc("/credential", mgr.ListCredentials).Methods("GET")
	r.HandleFunc("/credential", mgr.CreateCredential).Methods("POST")
	r.HandleFunc("/credential/{id}", mgr.GetCredential).Methods("GET")
	//r.HandleFunc("/credential/{id}", mgr.UpdateCredential).Methods("PATCH")
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

	r.HandleFunc("/client", mgr.ListClients).Methods("GET")
	r.HandleFunc("/client", mgr.CreateClient).Methods("POST")
	r.HandleFunc("/client/{id}", mgr.GetClient).Methods("GET")
	r.HandleFunc("/client/{id}", mgr.UpdateClient).Methods("PATCH")
	r.HandleFunc("/client/{id}", mgr.DeleteClient).Methods("DELETE")

	r.HandleFunc("/client/{id}/network", mgr.GetClientNetwork).Methods("GET")
	r.HandleFunc("/client/{id}/network", mgr.UnsetClientNetwork).Methods("DELETE")
	r.HandleFunc("/client/{id}/network/{network}", mgr.SetClientNetwork).Methods("POST")
}

type Manager struct {
	db  *database.Database
	rec *reconciler.Reconciler
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
