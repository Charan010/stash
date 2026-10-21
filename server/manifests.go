package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Charan010/stash/hash"
	objecthash "github.com/Charan010/stash/hash"
	"github.com/Charan010/stash/manifest"
	"github.com/Charan010/stash/store"
)

type CommitManifestRequest struct {
	Size   uint64   `json:"size"`
	Chunks []string `json:"chunks"`
}

type commitManifestResponse struct {
	ObjectID string `json:"object_id"`
}

func (s *Server) handleCommitManifest(w http.ResponseWriter, r *http.Request) {

	var req CommitManifestRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid JSON body")
		return
	}

	chunks := make([]objecthash.Hash, 0, len(req.Chunks))

	for _, raw := range req.Chunks {

		hash, err := hash.Parse(raw)

		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_HASH", "invalid chunk hash")
			return
		}

		chunks = append(chunks, hash)
	}

	m := manifest.New(req.Size, chunks)

	objectID, err := s.store.CommitManifest(m)

	if err != nil {

		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusConflict, "MISSING_CHUNK", "manifest references a missing chunk")
			return
		}

		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to commit manifest")
		return
	}

	writeJSON(w, http.StatusCreated, commitManifestResponse{
		ObjectID: objectID.String(),
	})

}

func (s *Server) handleGetObject(w http.ResponseWriter, r *http.Request) {

	rawID := r.PathValue("id")

	id, err := objecthash.Parse(rawID)

	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_HASH", "invalid object id")
		return
	}

	m, err := s.store.GetManifest(id)

	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "object not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to load manifest")
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatUint(m.Size, 10))

	w.WriteHeader(http.StatusOK)

	_ = s.store.RestoreManifest(m, w)
}
