package server

import (
	"encoding/json"
	"net/http"

	objecthash "github.com/Charan010/stash/hash"
	"github.com/Charan010/stash/store"
)

type Server struct {
	store *store.Store
	mux   *http.ServeMux
}

type missingChunksRequest struct {
	Hashes []string `json:"hashes"`
}

type missingChunksResponse struct {
	Missing []string `json:"missing"`
}

const maxMissingHashes = 10_000

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
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleMissingChunks(w http.ResponseWriter, r *http.Request) {

	var req missingChunksRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid json body")
		return
	}

	if len(req.Hashes) > maxMissingHashes {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "too many hashes in one request")
		return
	}

	hashes := make([]objecthash.Hash, 0, len(req.Hashes))

	for _, rawHash := range req.Hashes {

		hash, err := objecthash.Parse(rawHash)

		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_HASH", "invalid chunk hash")
			return
		}

		hashes = append(hashes, hash)
	}

	missing := s.store.MissingHashes(hashes)

	response := missingChunksResponse{Missing: make([]string, 0, len(missing))}

	for _, hash := range missing {
		response.Missing = append(response.Missing, hash.String())
	}

	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, value any) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(value)
}
