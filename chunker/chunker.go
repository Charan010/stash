package chunker

import (
	"io"

	fastcdc "github.com/jotfs/fastcdc-go"
)

const (
	MinChunkSize = 256 << 10
	AvgChunkSize = 1 << 20
	MaxChunkSize = 4 << 20
)

type Config struct {
	MinSize int
	AvgSize int
	MaxSize int
}

func DefaultConfig() Config {
	return Config{
		MinSize: MinChunkSize,
		AvgSize: AvgChunkSize,
		MaxSize: MaxChunkSize,
	}
}

type Chunker struct {
	inner *fastcdc.Chunker
	err   error
}

func NewReader(r io.Reader, cfg Config) *Chunker {
	inner, err := fastcdc.NewChunker(
		r,
		fastcdc.Options{
			MinSize:     cfg.MinSize,
			AverageSize: cfg.AvgSize,
			MaxSize:     cfg.MaxSize,
		},
	)

	return &Chunker{
		inner: inner,
		err:   err,
	}
}

func (c *Chunker) Next() ([]byte, bool) {
	if c.err != nil {
		return nil, false
	}

	chunk, err := c.inner.Next()

	if err == io.EOF {
		return nil, false
	}

	if err != nil {
		c.err = err
		return nil, false
	}

	/*
		fastCDC might reuse this buffer for next chunk. So we need to copy the chunk
		and return to the user.
	*/
	data := make([]byte, len(chunk.Data))
	copy(data, chunk.Data)

	return data, true
}

func (c *Chunker) Err() error {
	return c.err
}
