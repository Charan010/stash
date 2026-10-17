package binio

import "encoding/binary"

type Writer struct {
	buf []byte
}

func NewWriter(capacity int) *Writer {
	if capacity < 0 {
		capacity = 0
	}

	return &Writer{
		buf: make([]byte, 0, capacity),
	}
}

func (w *Writer) Len() int {
	return len(w.buf)
}

func (w *Writer) Uint8(v uint8) {
	w.buf = append(w.buf, v)
}

func (w *Writer) Uint16(v uint16) {
	var buf [2]byte

	binary.LittleEndian.PutUint16(buf[:], v)

	w.buf = append(w.buf, buf[:]...)
}

func (w *Writer) Uint32(v uint32) {
	var buf [4]byte

	binary.LittleEndian.PutUint32(buf[:], v)

	w.buf = append(w.buf, buf[:]...)
}

func (w *Writer) Uint64(v uint64) {
	var buf [8]byte

	binary.LittleEndian.PutUint64(buf[:], v)

	w.buf = append(w.buf, buf[:]...)
}

func (w *Writer) Int8(v int8) {
	w.Uint8(uint8(v))
}

func (w *Writer) Int16(v int16) {
	w.Uint16(uint16(v))
}

func (w *Writer) Int32(v int32) {
	w.Uint32(uint32(v))
}

func (w *Writer) Int64(v int64) {
	w.Uint64(uint64(v))
}

func (w *Writer) Bytes(data []byte) {
	w.buf = append(w.buf, data...)
}

func (w *Writer) Data() []byte {
	return w.buf
}

func (w *Writer) Copy() []byte {
	out := make([]byte, len(w.buf))
	copy(out, w.buf)

	return out
}

func (w *Writer) Reset() {
	w.buf = w.buf[:0]
}
