package pack

import (
	"fmt"

	"github.com/Charan010/stash/binio"
	objecthash "github.com/Charan010/stash/hash"
)

const (
	MagicSize        = 8
	HeaderSize       = MagicSize + 4 + 8
	ObjectHeaderSize = objecthash.Size + 8 + 8 + 4
)

func EncodeHeader(h Header) []byte {
	w := binio.NewWriter(HeaderSize)

	w.Bytes([]byte(Magic))
	w.Uint32(h.Version)
	w.Uint64(h.ObjectCount)

	return w.Data()
}

func DecodeHeader(data []byte) (Header, error) {
	var h Header

	if len(data) < HeaderSize {
		return h, fmt.Errorf(
			"pack: invalid header: need %d bytes, got %d",
			HeaderSize,
			len(data),
		)
	}

	r := binio.NewReader(data[:HeaderSize])

	magic, err := r.Bytes(MagicSize)
	if err != nil {
		return h, fmt.Errorf("pack: read magic: %w", err)
	}

	if string(magic) != Magic {
		return h, fmt.Errorf(
			"pack: invalid magic: got %q",
			string(magic),
		)
	}

	h.Version, err = r.Uint32()
	if err != nil {
		return h, fmt.Errorf("pack: read version: %w", err)
	}

	if h.Version != Version {
		return h, fmt.Errorf(
			"pack: unsupported version: got %d want %d",
			h.Version,
			Version,
		)
	}

	h.ObjectCount, err = r.Uint64()
	if err != nil {
		return h, fmt.Errorf("pack: read object count: %w", err)
	}

	return h, nil
}

func EncodeObjectHeader(h ObjectHeader) []byte {
	w := binio.NewWriter(ObjectHeaderSize)

	w.Bytes(h.Hash[:])
	w.Uint64(h.RawSize)
	w.Uint64(h.StoredSize)
	w.Uint32(h.CRC32)

	return w.Data()
}

func DecodeObjectHeader(data []byte) (ObjectHeader, error) {
	var h ObjectHeader

	if len(data) < ObjectHeaderSize {
		return h, fmt.Errorf(
			"pack: invalid object header: need %d bytes, got %d",
			ObjectHeaderSize,
			len(data),
		)
	}

	r := binio.NewReader(data[:ObjectHeaderSize])

	hashBytes, err := r.Bytes(objecthash.Size)
	if err != nil {
		return h, fmt.Errorf("pack: read object hash: %w", err)
	}

	copy(h.Hash[:], hashBytes)

	h.RawSize, err = r.Uint64()
	if err != nil {
		return h, fmt.Errorf("pack: read raw size: %w", err)
	}

	h.StoredSize, err = r.Uint64()
	if err != nil {
		return h, fmt.Errorf("pack: read stored size: %w", err)
	}

	h.CRC32, err = r.Uint32()
	if err != nil {
		return h, fmt.Errorf("pack: read crc32: %w", err)
	}

	return h, nil
}
