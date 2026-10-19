package pack

import (
	"fmt"
	"io"

	"github.com/Charan010/stash/checksum"
)

type Writer struct {
	w      io.Writer
	offset uint64
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: w,
	}
}

func NewWriterAt(w io.Writer, offset uint64) *Writer {
	return &Writer{
		w:      w,
		offset: offset,
	}
}

func (pw *Writer) Offset() uint64 {
	return pw.offset
}

func (pw *Writer) WriteHeader(h Header) error {
	data := EncodeHeader(h)

	n, err := writeAll(pw.w, data)
	if err != nil {
		return fmt.Errorf("pack: write header: %w", err)
	}

	pw.offset += uint64(n)

	return nil
}

func (pw *Writer) WriteObject(obj Object) (uint64, error) {
	actualSize := uint64(len(obj.Data))

	if obj.Header.StoredSize != actualSize {
		return 0, fmt.Errorf("pack: stored size mismatch: header=%d data=%d", obj.Header.StoredSize, actualSize)
	}

	actualCRC := checksum.CRC32(obj.Data)

	if obj.Header.CRC32 != actualCRC {
		return 0, fmt.Errorf("pack: crc32 mismatch before write: header=%08x actual=%08x", obj.Header.CRC32, actualCRC)
	}

	objectOffset := pw.offset

	header := EncodeObjectHeader(obj.Header)

	n, err := writeAll(pw.w, header)
	if err != nil {
		return 0, fmt.Errorf("pack: write object header: %w", err)
	}

	pw.offset += uint64(n)

	n, err = writeAll(pw.w, obj.Data)
	if err != nil {
		return 0, fmt.Errorf("pack: write object data: %w", err)
	}

	pw.offset += uint64(n)

	return objectOffset, nil
}

func writeAll(w io.Writer, data []byte) (int, error) {
	total := 0

	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return total, err
		}

		if n == 0 {
			return total, io.ErrShortWrite
		}

		total += n
		data = data[n:]
	}

	return total, nil
}
