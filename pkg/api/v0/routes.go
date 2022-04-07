package v0

import (
	"log"

	"github.com/gorilla/mux"
	"github.com/pricec/vpnmux/pkg/network"
)

func RegisterHandlers(r *mux.Router) {
	instances, err := network.RecoverVPNInstances()
	if err != nil {
		log.Panicf("failed to recover instances: %v", err)
	}

	im := &InstanceManager{
		Instances: instances,
	}
	r.HandleFunc("/network", im.List).Methods("GET")
	r.HandleFunc("/network", im.Create).Methods("POST")
	r.HandleFunc("/network/{id}", im.Delete).Methods("DELETE")

	db, err := network.NewDatabase()
	if err != nil {
		log.Panicf("failed to instantiate database: %v", err)
	}

	cm, err := NewClientManager(db)
	if err != nil {
		log.Panicf("failed to create client manager: %v", err)
	}
	r.HandleFunc("/client", cm.List).Methods("GET")
	r.HandleFunc("/client", cm.Create).Methods("POST")
	r.HandleFunc("/client/{addr}", cm.Delete).Methods("DELETE")
}
