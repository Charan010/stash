package index

import (
	"fmt"

	"github.com/Charan010/stash/binio"
	"github.com/Charan010/stash/checksum"
	objecthash "github.com/Charan010/stash/hash"
)

/*
	Record length.

	Each record is state of index (in memory map) persisted to disk to recover from crashes.

	field 		size
	Hash		32 bytes
	PackID		8 bytes
	Offset		8 bytes
	CRC32		4 bytes
*/

const (
	RecordPayloadSize = objecthash.Size + 8 + 8
	RecordSize        = RecordPayloadSize + 4
)

type Record struct {
	Hash   objecthash.Hash
	PackID uint64
	Offset uint64
}

func EncodeRecord(r Record) []byte {
	w := binio.NewWriter(RecordSize)

	w.Bytes(r.Hash[:])
	w.Uint64(r.PackID)
	w.Uint64(r.Offset)

	payload := w.Data()
	crc := checksum.CRC32(payload)

	w.Uint32(crc)

	return w.Data()
}

func DecodeRecord(data []byte) (Record, error) {
	var r Record

	if len(data) < RecordSize {
		return r, fmt.Errorf("index: invalid record: need %d bytes, got %d", RecordSize, len(data))
	}

	payload := data[:RecordPayloadSize]

	expectedCRC, err := binio.NewReader(data[RecordPayloadSize:RecordSize]).Uint32()

	if err != nil {
		return r, fmt.Errorf("index: read record crc: %w", err)
	}

	actualCRC := checksum.CRC32(payload)

	if actualCRC != expectedCRC {
		return r, fmt.Errorf("index: record crc mismatch: expected=%08x actual=%08x", expectedCRC, actualCRC)
	}

	reader := binio.NewReader(payload)

	hashBytes, err := reader.Bytes(objecthash.Size)
	if err != nil {
		return r, fmt.Errorf("index: read hash: %w", err)
	}

	copy(r.Hash[:], hashBytes)

	r.PackID, err = reader.Uint64()
	if err != nil {
		return r, fmt.Errorf("index: read pack id: %w", err)
	}

	r.Offset, err = reader.Uint64()
	if err != nil {
		return r, fmt.Errorf("index: read offset: %w", err)
	}

	return r, nil
}
