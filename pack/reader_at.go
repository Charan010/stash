package pack

import (
	"fmt"
	"io"

	"github.com/Charan010/stash/checksum"
	objecthash "github.com/Charan010/stash/hash"
)

func ReadObjectAt(r io.ReaderAt, offset int64) (Object, error) {

	headerBuf := make([]byte, ObjectHeaderSize)

	if _, err := r.ReadAt(headerBuf, offset); err != nil {
		return Object{}, fmt.Errorf("pack: read object header at offset %d : %w", offset, err)
	}

	h, err := DecodeObjectHeader(headerBuf)
	if err != nil {
		return Object{}, fmt.Errorf("pack: decode object header at offset %d: %w", offset, err)
	}

	if h.StoredSize > uint64(^uint(0)>>1) {
		return Object{}, fmt.Errorf("pack: object too large: %d bytes", h.StoredSize)
	}

	data := make([]byte, int(h.StoredSize))

	dataOffset := offset + int64(ObjectHeaderSize)

	if _, err := r.ReadAt(data, dataOffset); err != nil {
		return Object{}, fmt.Errorf("pack: read object data at offset %d: %w", dataOffset, err)
	}

	actualCRC := checksum.CRC32(data)

	if actualCRC != h.CRC32 {
		return Object{}, fmt.Errorf("pack: crc32 mismatch: expected=%08x actual=%08x", h.CRC32, actualCRC)
	}

	actualHash := objecthash.Sum(data)

	if actualHash != h.Hash {
		return Object{}, fmt.Errorf("pack: object hash mismatch: expected=%s actual=%s", h.Hash, actualHash)
	}

	return Object{
		Header: h,
		Data:   data,
	}, nil

}
