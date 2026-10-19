package manifest

import (
	"fmt"
	"io"
	"os"

	"github.com/Charan010/stash/binio"
	objecthash "github.com/Charan010/stash/hash"
)

func Read(path string) (Manifest, error) {
	file, err := os.Open(path)
	if err != nil {

		return Manifest{}, fmt.Errorf("manifest: open: %w", err)
	}
	defer file.Close()

	return ReadFrom(file)
}

func ReadFrom(r io.Reader) (Manifest, error) {
	var m Manifest

	header := make([]byte, headerSize)

	if _, err := io.ReadFull(r, header); err != nil {
		return m, fmt.Errorf("manifest: read header: %w", err)
	}

	br := binio.NewReader(header)

	magicBytes, err := br.Bytes(8)
	if err != nil {
		return m, fmt.Errorf("manifest: read magic: %w", err)
	}

	if string(magicBytes) != magic {
		return m, fmt.Errorf("manifest: invalid magic: got %q", string(magicBytes))
	}

	gotVersion, err := br.Uint32()
	if err != nil {
		return m, fmt.Errorf("manifest: read version: %w", err)
	}

	if gotVersion != version {
		return m, fmt.Errorf("manifest: unsupported version: got %d want %d", gotVersion, version)
	}

	m.Size, err = br.Uint64()
	if err != nil {
		return m, fmt.Errorf("manifest: read size: %w", err)
	}

	chunkCount, err := br.Uint64()
	if err != nil {
		return m, fmt.Errorf("manifest: read chunk count: %w", err)
	}

	const maxChunks = uint64(100_000_000)

	if chunkCount > maxChunks {
		return m, fmt.Errorf("manifest: unreasonable chunk count: %d", chunkCount)
	}

	m.Chunks = make([]objecthash.Hash, int(chunkCount))

	for i := range m.Chunks {
		if _, err := io.ReadFull(r, m.Chunks[i][:]); err != nil {
			return Manifest{}, fmt.Errorf("manifest: read chunk %d: %w", i, err)
		}
	}

	if err := m.Validate(); err != nil {
		return Manifest{}, fmt.Errorf("manifest: invalid: %w", err)
	}

	return m, nil
}
