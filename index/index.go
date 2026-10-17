package index

import (
	"sync"

	objecthash "github.com/Charan010/stash/hash"
)

/*

	Index basically answers given a hash, where does the actual content or data exist
	on the disk.

	hash -> (packfileID, offset)

	rest of the metadata is self described in the packfiles.
*/

type Location struct {
	PackID uint64
	Offset uint64
}

type Index struct {
	mu      sync.RWMutex
	entries map[objecthash.Hash]Location
}

func New() *Index {
	return &Index{
		entries: make(map[objecthash.Hash]Location),
	}
}

func (index *Index) Put(hash objecthash.Hash, loc Location) {

	index.mu.Lock()
	index.entries[hash] = loc
	index.mu.Unlock()
}

func (index *Index) Get(hash objecthash.Hash) (Location, bool) {

	index.mu.RLock()
	loc, ok := index.entries[hash]
	index.mu.RUnlock()

	return loc, ok
}

func (idx *Index) Delete(hash objecthash.Hash) {
	idx.mu.Lock()
	delete(idx.entries, hash)
	idx.mu.Unlock()
}

func (idx *Index) Len() int {
	idx.mu.RLock()
	n := len(idx.entries)
	idx.mu.RUnlock()

	return n
}

func (idx *Index) Reset() {
	idx.mu.Lock()
	idx.entries = make(map[objecthash.Hash]Location)
	idx.mu.Unlock()
}
