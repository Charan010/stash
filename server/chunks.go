package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/Charan010/stash/hash"
	objecthash "github.com/Charan010/stash/hash"
	"github.com/Charan010/stash/store"
)

type missingChunksRequest struct {
	Hashes []string `json:"hashes"`
}

type missingChunksResponse struct {
	Missing []string `json:"missing"`
}

const maxMissingHashes = 10_000

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

func (s *Server) handleGetChunk(w http.ResponseWriter, r *http.Request) {

	rawHash := r.PathValue("hash")

	hash, err := hash.Parse(rawHash)

	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_HASH", "invalid chunk hash")
		return
	}

	data, err := s.store.Get(hash)

	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "chunk not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "INTERNAL", "unable to fetch chunk")
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(data); err != nil {
		return
	}
}

func (s *Server) handlePutChunk(w http.ResponseWriter, r *http.Request) {

	rawHash := r.PathValue("hash")

	claimedHash, err := hash.Parse(rawHash)

	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_HASH", "invalid chunk hash")
		return
	}

	const maxChunkSize = 4 << 20 // 4MB

	r.Body = http.MaxBytesReader(w, r.Body, maxChunkSize)

	data, err := io.ReadAll(r.Body)

	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "failed to read chunk body")
		return
	}

	actualHash := objecthash.Sum(data)

	if actualHash != claimedHash {
		writeError(w, http.StatusBadRequest, "HASH_MISMATCH", "chunk body does not match claimed hash")
		return
	}

	storedHash, err := s.store.Put(data)

	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "failed to store chunk")
		return
	}

	if storedHash != claimedHash {
		writeError(w, http.StatusInternalServerError, "INTERNAL", "stored chunk hash mismatch")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
