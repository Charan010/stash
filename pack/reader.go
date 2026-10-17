package pack

import (
	"fmt"
	"io"

	"github.com/Charan010/stash/checksum"
	objecthash "github.com/Charan010/stash/hash"
)

type Reader struct {
	r io.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		r: r,
	}
}

func (pr *Reader) ReadHeader() (Header, error) {
	buf := make([]byte, HeaderSize)

	if _, err := io.ReadFull(pr.r, buf); err != nil {
		return Header{}, fmt.Errorf("pack: read header: %w", err)
	}

	h, err := DecodeHeader(buf)
	if err != nil {
		return Header{}, fmt.Errorf("pack: decode header: %w", err)
	}

	return h, nil
}

func (pr *Reader) ReadObject() (Object, error) {
	headerBuf := make([]byte, ObjectHeaderSize)

	if _, err := io.ReadFull(pr.r, headerBuf); err != nil {
		return Object{}, fmt.Errorf("pack: read object header: %w", err)
	}

	h, err := DecodeObjectHeader(headerBuf)
	if err != nil {
		return Object{}, fmt.Errorf("pack: decode object header: %w", err)
	}

	if h.StoredSize > uint64(^uint(0)>>1) {
		return Object{}, fmt.Errorf("pack: object too large: %d bytes", h.StoredSize)
	}

	data := make([]byte, int(h.StoredSize))

	if _, err := io.ReadFull(pr.r, data); err != nil {
		return Object{}, fmt.Errorf("pack: read object data: %w", err)
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
