package index

import (
	"fmt"
	"io"
	"os"

	"github.com/Charan010/stash/pack"
)

func RebuildFromPack(index *Index, PackID uint64, file *os.File) error {

	headerBuf := make([]byte, pack.HeaderSize)

	if _, err := file.ReadAt(headerBuf, 0); err != nil {
		return fmt.Errorf("index: read pack header : %w\n", err)
	}

	header, err := pack.DecodeHeader(headerBuf)
	if err != nil {
		return fmt.Errorf("index: decode pack header: %w", err)
	}

	offset := uint64(pack.HeaderSize)
	for i := uint64(0); i < header.ObjectCount; i++ {

		objectHeaderBuf := make([]byte, pack.ObjectHeaderSize)

		_, err := file.ReadAt(objectHeaderBuf, int64(offset))

		if err != nil {
			if err == io.EOF {
				return fmt.Errorf("index: unexpected end of pack at object %d", i)
			}

			return fmt.Errorf("index: read object header at offset %d: %w", offset, err)
		}

		objectHeader, err := pack.DecodeObjectHeader(objectHeaderBuf)
		if err != nil {
			return fmt.Errorf("index: decode object header at offset %d: %w", offset, err)
		}

		index.Put(objectHeader.Hash, Location{PackID: PackID, Offset: offset})

		offset += uint64(pack.ObjectHeaderSize)
		offset += objectHeader.StoredSize
	}

	return nil

}
