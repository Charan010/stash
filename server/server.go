package server

import (
	"net/http"

	"github.com/Charan010/stash/store"
)

type Server struct {
	store *store.Store
	mux   *http.ServeMux
}

func New(s *store.Store) *Server {

	srv := &Server{
		store: s,
		mux:   http.NewServeMux(),
	}

	srv.routes()

	return srv
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("POST /v1/chunks/missing", s.handleMissingChunks)
	s.mux.HandleFunc("PUT /v1/chunks/{hash}", s.handlePutChunk)
	s.mux.HandleFunc("POST /v1/manifests", s.handleCommitManifest)
	s.mux.HandleFunc("GET /v1/chunks/{hash}", s.handleGetChunk)
	s.mux.HandleFunc("GEt /v1/objects/{hash}", s.handleGetObject)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
