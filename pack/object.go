package pack

import objecthash "github.com/Charan010/stash/hash"

type ObjectHeader struct {
	Hash       objecthash.Hash
	RawSize    uint64
	StoredSize uint64
	CRC32      uint32
}

type Object struct {
	Header ObjectHeader
	Data   []byte
}
