package v1

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pricec/vpnmux/pkg/database"
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
