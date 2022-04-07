package v0

import (
	"log"

	"github.com/gorilla/mux"
	"github.com/pricec/vpnmux/pkg/network"
)

func RegisterHandlers(r *mux.Router) {
	db, err := network.NewDatabase()
	if err != nil {
		log.Panicf("failed to instantiate database: %v", err)
	}

	mgr, err := NewManager(db)
	if err != nil {
		log.Panicf("failed to set up manager: %v", err)
	}

	r.HandleFunc("/network", mgr.ListNetworks).Methods("GET")
	r.HandleFunc("/network", mgr.CreateNetwork).Methods("POST")
	r.HandleFunc("/network/{id}", mgr.DescribeNetwork).Methods("GET")
	r.HandleFunc("/network/{id}", mgr.DeleteNetwork).Methods("DELETE")

	r.HandleFunc("/client", mgr.ListClients).Methods("GET")
	r.HandleFunc("/client", mgr.CreateClient).Methods("POST")
	r.HandleFunc("/client/{addr}", mgr.DeleteClient).Methods("DELETE")

	r.HandleFunc("/client/{addr}/network", mgr.ListClientNetwork).Methods("GET")
	r.HandleFunc("/client/{addr}/network/{network}", mgr.AssignClientNetwork).Methods("POST")
	r.HandleFunc("/client/{addr}/network", mgr.DeleteClientNetwork).Methods("DELETE")
}
