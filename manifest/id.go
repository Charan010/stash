package manifest

import (
	"github.com/Charan010/stash/binio"
	objecthash "github.com/Charan010/stash/hash"
)

func (m Manifest) ID() objecthash.Hash {

	w := binio.NewWriter(8 + 8 + len(m.Chunks)*objecthash.Size)

	w.Uint64(m.Size)
	w.Uint64(uint64(len(m.Chunks)))

	for _, chunkHash := range m.Chunks {
		w.Bytes(chunkHash[:])
	}

	return objecthash.Sum(w.Data())

}
