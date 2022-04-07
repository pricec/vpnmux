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

type ClientNetwork struct {
	address string
	network string
	rtID    int
}

type Manager struct {
	sync.Mutex
	db          *network.Database
	clients     map[string]*network.VPNClient
	networks    map[uuid.UUID]*network.VPNInstance
	assignments map[string]*ClientNetwork
}

func NewManager(db *network.Database) (*Manager, error) {
	mgr := &Manager{
		db:          db,
		clients:     make(map[string]*network.VPNClient),
		assignments: make(map[string]*ClientNetwork),
	}

	if err := mgr.restore(); err != nil {
		return nil, err
	}

	return mgr, nil
}

func (m *Manager) ListClientNetwork(w http.ResponseWriter, r *http.Request) {

}

func (m *Manager) AssignClientNetwork(w http.ResponseWriter, r *http.Request) {

}

func (m *Manager) DeleteClientNetwork(w http.ResponseWriter, r *http.Request) {

}

func (m *Manager) restore() error {
	networks, err := network.RecoverVPNInstances()
	if err != nil {
		return err
	}
	m.networks = networks

	clients, err := m.db.GetClients()
	if err != nil {
		return err
	}

	for _, client := range clients {
		client, err := network.NewVPNClient(client.Address)
		if err != nil {
			return err
		}

		m.clients[client.Address] = client
	}
	return nil
}

func (m *Manager) ListClients(w http.ResponseWriter, r *http.Request) {
	m.Lock()
	defer m.Unlock()

	ids := make([]string, 0, len(m.clients))
	for id, _ := range m.clients {
		ids = append(ids, id)
	}

	if err := json.NewEncoder(w).Encode(ids); err != nil {
		log.Printf("error encoding response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type createClientRequest struct {
	Address string `json:"address"`
}

func (m *Manager) CreateClient(w http.ResponseWriter, r *http.Request) {
	body := &createClientRequest{}
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Printf("error decoding body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	m.Lock()
	defer m.Unlock()

	if _, ok := m.clients[body.Address]; ok {
		log.Printf("attempt to create duplicate client %v", body.Address)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := m.db.PutClient(body.Address); err != nil {
		log.Printf("error create client in db: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	client, err := network.NewVPNClient(body.Address)
	if err != nil {
		// TODO: clean up database?
		log.Printf("error creating client: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	m.clients[body.Address] = client
}

func (m *Manager) DeleteClient(w http.ResponseWriter, r *http.Request) {
	addr := mux.Vars(r)["addr"]

	m.Lock()
	defer m.Unlock()

	client, ok := m.clients[addr]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := m.db.DeleteClient(addr); err != nil {
		log.Printf("error deleting client from database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := client.Close(); err != nil {
		log.Printf("error cleaning up client: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	delete(m.clients, addr)
}

func (m *Manager) ListNetworks(w http.ResponseWriter, r *http.Request) {
	m.Lock()
	defer m.Unlock()

	ids := make([]string, 0, len(m.networks))
	for id, _ := range m.networks {
		ids = append(ids, id.String())
	}

	if err := json.NewEncoder(w).Encode(ids); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// try ve-car.prod.surfshark.com_tcp.ovpn
// try us-chi.prod.surfshark.com_udp.ovpn
type createNetworkRequest struct {
	Config string `json:"config"`
}

func (m *Manager) CreateNetwork(w http.ResponseWriter, r *http.Request) {
	m.Lock()
	defer m.Unlock()

	id := uuid.New()
	body := &createNetworkRequest{}
	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		log.Printf("error decoding body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	network, err := network.NewVPNInstance(id.String(), body.Config)
	if err != nil {
		log.Printf("error creating instance: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	m.networks[id] = network
	json.NewEncoder(w).Encode(id.String())
}

func (m *Manager) DeleteNetwork(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := uuid.Parse(params["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	m.Lock()
	defer m.Unlock()
	network, ok := m.networks[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := network.Close(); err != nil {
		log.Printf("error deleting instance: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	delete(m.networks, id)
}
