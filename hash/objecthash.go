package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

const Size = sha256.Size

type Hash [Size]byte

func Sum(data []byte) Hash {
	return sha256.Sum256(data)
}

func Parse(s string) (Hash, error) {
	var h Hash

	if len(s) != Size*2 {
		return h, fmt.Errorf("objecthash: expected %d hex characters, got %d", Size*2, len(s))
	}

	n, err := hex.Decode(h[:], []byte(s))
	if err != nil {
		return Hash{}, fmt.Errorf("objecthash: invalid hash: %w", err)
	}

	if n != Size {
		return Hash{}, errors.New("objecthash: unexpected decoded hash length")
	}

	return h, nil
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func (h Hash) Bytes() []byte {
	out := make([]byte, Size)
	copy(out, h[:])

	return out
}

func (h Hash) IsZero() bool {
	return h == Hash{}
}
