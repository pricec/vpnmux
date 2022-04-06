package v0

import (
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterHandlers(r *mux.Router) {
	r.HandleFunc("/test", test).Methods("GET")
}

func test(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
