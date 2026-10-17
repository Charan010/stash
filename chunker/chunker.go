package chunker

const (
	DefaultWindow        = 48
	DefaultBase   uint64 = 257
	DefaultMask   uint64 = 0x1FFF

	MinChunkSize = 2 << 10
	MaxChunkSize = 1 << 20
)

type Config struct {
	Window int
	Base   uint64
	Mask   uint64

	MinSize int
	MaxSize int
}

func DefaultConfig() Config {
	return Config{
		Window:  DefaultWindow,
		Base:    DefaultBase,
		Mask:    DefaultMask,
		MinSize: MinChunkSize,
		MaxSize: MaxChunkSize,
	}
}

type Chunker struct {
	data []byte
	pos  int

	window  int
	base    uint64
	mask    uint64
	minSize int
	maxSize int

	highestPower uint64
}

func New(data []byte, cfg Config) *Chunker {

	if cfg.Window <= 0 {
		panic("chunking: window must be >= 0")
	}

	if cfg.MaxSize < cfg.MinSize {
		panic("chunking: max chunk size must be >= min chunk size")
	}

	highestPower := uint64(1)

	for i := 1; i < cfg.Window; i++ {
		highestPower *= cfg.Base
	}

	return &Chunker{
		data:         data,
		window:       cfg.Window,
		base:         cfg.Base,
		mask:         cfg.Mask,
		minSize:      cfg.MinSize,
		maxSize:      cfg.MaxSize,
		highestPower: highestPower,
	}
}

func (c *Chunker) Next() ([]byte, bool) {
	if c.pos >= len(c.data) {
		return nil, false
	}

	start := c.pos

	var hash uint64

	for c.pos < len(c.data) {
		incoming := c.data[c.pos]
		c.pos++

		windowLen := c.pos - start

		if windowLen <= c.window {
			hash = hash*c.base + uint64(incoming)
		} else {
			outgoing := c.data[c.pos-c.window-1]

			hash -= uint64(outgoing) * c.highestPower
			hash *= c.base
			hash += uint64(incoming)
		}

		if windowLen >= c.minSize && ((hash&c.mask) == 0 || windowLen >= c.maxSize) {
			break
		}
	}

	return c.data[start:c.pos], true
}
