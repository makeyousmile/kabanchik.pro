package handlers

import (
	"net/http"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/health", health)
}
