package v0

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/pricec/vpnmux/pkg/network"
)

type ClientManager struct {
	sync.Mutex
	db      *network.Database
	clients map[string]*network.VPNClient
}

func NewClientManager(db *network.Database) (*ClientManager, error) {
	// Start by restoring existing clients
	mgr := &ClientManager{
		db:      db,
		clients: make(map[string]*network.VPNClient),
	}

	if err := mgr.restore(); err != nil {
		return nil, err
	}

	return mgr, nil
}

func (m *ClientManager) restore() error {
	clients, err := m.db.GetClients()
	if err != nil {
		return err
	}

	for _, addr := range clients {
		client, err := network.NewVPNClient(addr)
		if err != nil {
			return err
		}

		m.clients[addr] = client
	}
	return nil
}

func (m *ClientManager) List(w http.ResponseWriter, r *http.Request) {
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

func (m *ClientManager) Create(w http.ResponseWriter, r *http.Request) {
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

func (m *ClientManager) Delete(w http.ResponseWriter, r *http.Request) {
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
