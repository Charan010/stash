package store

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/Charan010/stash/checksum"
	objecthash "github.com/Charan010/stash/hash"
	"github.com/Charan010/stash/index"
	"github.com/Charan010/stash/manifest"
	"github.com/Charan010/stash/pack"
	"github.com/Charan010/stash/packfile"
)

var ErrNotFound = errors.New("store: object not found")

type Store struct {
	index *index.Index
	log   *index.Log

	packs *packfile.Manager

	activePackID uint64
	writer       *pack.Writer

	manifestDir string

	putMu sync.Mutex
}

func New(index *index.Index, log *index.Log, packs *packfile.Manager, activePackID uint64,
	writer *pack.Writer, manifestDir string) *Store {

	return &Store{
		index:        index,
		log:          log,
		packs:        packs,
		activePackID: activePackID,
		writer:       writer,
		manifestDir:  manifestDir,
	}
}

func (s *Store) Has(hash objecthash.Hash) bool {

	_, ok := s.index.Get(hash)
	return ok
}

func (s *Store) Put(data []byte) (objecthash.Hash, error) {

	hash := objecthash.Sum(data)

	s.putMu.Lock()
	defer s.putMu.Unlock()

	if _, ok := s.index.Get(hash); ok {
		return hash, nil
	}

	obj := pack.Object{
		Header: pack.ObjectHeader{
			Hash:       hash,
			RawSize:    uint64(len(data)),
			StoredSize: uint64(len(data)),
			CRC32:      checksum.CRC32(data),
		},

		Data: data,
	}

	offset, err := s.writer.WriteObject(obj)
	if err != nil {
		return objecthash.Hash{}, fmt.Errorf("store: write object: %w", err)
	}

	loc := index.Location{PackID: s.activePackID, Offset: offset}

	record := index.Record{
		Hash:   hash,
		PackID: loc.PackID,
		Offset: loc.Offset,
	}

	if err := s.log.Append(record); err != nil {
		return objecthash.Hash{}, fmt.Errorf("store: append index log: %w", err)
	}

	s.index.Put(hash, loc)
	return hash, nil

}

func (s *Store) Get(hash objecthash.Hash) ([]byte, error) {

	loc, ok := s.index.Get(hash)

	if !ok {
		return nil, ErrNotFound
	}

	file, err := s.packs.Open(loc.PackID)
	if err != nil {
		return nil, fmt.Errorf("store: open pack %d: %w", loc.PackID, err)
	}

	obj, err := pack.ReadObjectAt(file, int64(loc.Offset))

	if err != nil {
		return nil, fmt.Errorf("store: read object %s: %w", hash, err)
	}

	if obj.Header.Hash != hash {
		return nil, fmt.Errorf("store: hash mismatch: requested=%s stored=%s", hash, obj.Header.Hash)
	}

	return obj.Data, nil

}

func (s *Store) CommitManifest(m manifest.Manifest) (objecthash.Hash, error) {

	if err := m.Validate(); err != nil {
		return objecthash.Hash{}, fmt.Errorf("store: invalid manifest: %w", err)
	}

	for i, chunkHash := range m.Chunks {
		if !s.Has(chunkHash) {
			return objecthash.Hash{}, fmt.Errorf("store: manifest chunk %d (%s): %w", i, chunkHash, ErrNotFound)
		}
	}

	id := m.ID()

	path := filepath.Join(s.manifestDir, id.String()+".manifest")

	if err := manifest.WriteAtomic(path, m); err != nil {
		return objecthash.Hash{}, fmt.Errorf("store: commit manifest %s: %w", id, err)
	}

	return id, nil

}

func (s *Store) GetManifest(id objecthash.Hash) (manifest.Manifest, error) {

	path := filepath.Join(s.manifestDir, id.String()+".manifest")

	m, err := manifest.Read(path)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return manifest.Manifest{}, ErrNotFound
		}

		return manifest.Manifest{}, fmt.Errorf("store: mismatch hash match: requested = %s actual = %s", id, m.ID())
	}

	return m, nil
}

func (s *Store) Restore(id objecthash.Hash, w io.Writer) error {

	m, err := s.GetManifest(id)

	if err != nil {
		return fmt.Errorf("store: load manifest %s : %w", id, err)
	}

	var written uint64

	for i, chunkHash := range m.Chunks {
		data, err := s.Get(chunkHash)

		if err != nil {
			return fmt.Errorf("store: restore chunk %d (%s): %w", i, chunkHash, err)
		}

		if err := writeAll(w, data); err != nil {
			return fmt.Errorf("store: write restored chunk %d: %w", i, err)
		}

		written += uint64(len(data))
	}

	if written != m.Size {
		return fmt.Errorf("store: restored size mismatch: expected=%d actual=%d", m.Size, written)
	}

	return nil

}

func (s *Store) MissingHashes(hashes []objecthash.Hash) []objecthash.Hash {

	var missing []objecthash.Hash

	for _, hash := range hashes {
		if !s.Has(hash) {
			missing = append(missing, hash)
		}
	}

	return missing
}

func writeAll(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return err
		}

		if n == 0 {
			return io.ErrShortWrite
		}

		data = data[n:]
	}

	return nil
}
