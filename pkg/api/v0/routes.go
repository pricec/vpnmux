package v0

import (
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pricec/vpnmux/pkg/network"
)

func RegisterHandlers(r *mux.Router) {
	im := &InstanceManager{
		Instances: make(map[uuid.UUID]*network.VPNInstance),
	}
	r.HandleFunc("/network", im.List).Methods("GET")
	r.HandleFunc("/network", im.Create).Methods("POST")
	r.HandleFunc("/network/{id}", im.Delete).Methods("DELETE")
}
