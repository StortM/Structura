package api

import (
	"github.com/gorilla/mux"
)

// NewRouter creates a new router with all API routes
func NewRouter() *mux.Router {
	r := mux.NewRouter()

	// Main page
	r.HandleFunc("/", GetHomePage).Methods("GET")

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Schema management
	api.HandleFunc("/schemas", ListSchemas).Methods("GET")
	api.HandleFunc("/schemas", ImportSchema).Methods("POST")
	api.HandleFunc("/schemas/{id}", GetSchema).Methods("GET")
	api.HandleFunc("/schemas/{id}", DeleteSchema).Methods("DELETE")
	
	// Layout
	api.HandleFunc("/schemas/{id}/layout", ApplyLayout).Methods("POST")
	
	// Table manipulation
	api.HandleFunc("/schemas/{id}/tables/{tableId}/position", UpdateTablePosition).Methods("PUT")
	
		return r
	}