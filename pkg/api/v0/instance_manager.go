package v0

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pricec/vpnmux/pkg/network"
)

type InstanceManager struct {
	sync.Mutex
	Instances map[uuid.UUID]*network.VPNInstance
}

func (m *InstanceManager) List(w http.ResponseWriter, r *http.Request) {
	m.Lock()
	defer m.Unlock()

	ids := make([]string, 0, len(m.Instances))
	for id, _ := range m.Instances {
		ids = append(ids, id.String())
	}

	if err := json.NewEncoder(w).Encode(ids); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// try ve-car.prod.surfshark.com_tcp.ovpn
type createInstanceRequest struct {
	Config string `json:"config"`
}

func (m *InstanceManager) Create(w http.ResponseWriter, r *http.Request) {
	m.Lock()
	defer m.Unlock()

	id := uuid.New()
	body := &createInstanceRequest{}
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Printf("error decoding body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	instance, err := network.NewVPNInstance(id.String(), body.Config)
	if err != nil {
		log.Printf("error creating instance: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	m.Instances[id] = instance
	json.NewEncoder(w).Encode(id.String())
}

func (m *InstanceManager) Delete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	m.Lock()
	defer m.Unlock()
	instance, ok := m.Instances[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := instance.Close(); err != nil {
		log.Printf("error deleting instance: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	delete(m.Instances, id)
}
