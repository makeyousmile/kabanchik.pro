package handlers

import "net/http"

// Deprecated: use API.Register.
func Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/health", health)
}
