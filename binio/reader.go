package binio

import (
	"encoding/binary"
	"fmt"
)

type Reader struct {
	data []byte
	pos  int
}

func NewReader(data []byte) *Reader {
	return &Reader{
		data: data,
	}
}

func (r *Reader) Position() int {
	return r.pos
}

func (r *Reader) Remaining() int {
	return len(r.data) - r.pos
}

func (r *Reader) Bytes(n int) ([]byte, error) {
	if n < 0 {
		return nil, fmt.Errorf("binio: negative byte count: %d", n)
	}

	if r.Remaining() < n {
		return nil, fmt.Errorf(
			"binio: truncated input: need %d bytes, have %d",
			n,
			r.Remaining(),
		)
	}

	start := r.pos
	r.pos += n

	return r.data[start:r.pos], nil
}

func (r *Reader) Uint8() (uint8, error) {
	b, err := r.Bytes(1)
	if err != nil {
		return 0, err
	}

	return b[0], nil
}

func (r *Reader) Uint16() (uint16, error) {
	b, err := r.Bytes(2)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint16(b), nil
}

func (r *Reader) Uint32() (uint32, error) {
	b, err := r.Bytes(4)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint32(b), nil
}

func (r *Reader) Uint64() (uint64, error) {
	b, err := r.Bytes(8)
	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(b), nil
}

func (r *Reader) Int8() (int8, error) {
	v, err := r.Uint8()
	if err != nil {
		return 0, err
	}

	return int8(v), nil
}

func (r *Reader) Int16() (int16, error) {
	v, err := r.Uint16()
	if err != nil {
		return 0, err
	}

	return int16(v), nil
}

func (r *Reader) Int32() (int32, error) {
	v, err := r.Uint32()
	if err != nil {
		return 0, err
	}

	return int32(v), nil
}

func (r *Reader) Int64() (int64, error) {
	v, err := r.Uint64()
	if err != nil {
		return 0, err
	}

	return int64(v), nil
}

func (r *Reader) Skip(n int) error {
	_, err := r.Bytes(n)
	return err
}
