package manifest

import (
	"fmt"

	objecthash "github.com/Charan010/stash/hash"
)

type Manifest struct {
	Size   uint64
	Chunks []objecthash.Hash
}

func New(size uint64, chunks []objecthash.Hash) Manifest {
	copied := make([]objecthash.Hash, len(chunks))
	copy(copied, chunks)

	return Manifest{
		Size:   size,
		Chunks: copied,
	}
}

func (m Manifest) ChunkCount() int {
	return len(m.Chunks)
}

func (m Manifest) Validate() error {
	if m.Size == 0 && len(m.Chunks) != 0 {
		return fmt.Errorf("zero-size object cannot contain chunks")
	}

	if m.Size > 0 && len(m.Chunks) == 0 {
		return fmt.Errorf("non-empty object has no chunks")
	}

	for i, hash := range m.Chunks {
		if hash.IsZero() {
			return fmt.Errorf("chunk %d has zero hash", i)
		}
	}

	return nil
}
