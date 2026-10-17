package pack

const (
	Magic   = "STASHPAK"
	Version = uint32(1)
)

type Header struct {
	Version     uint32
	ObjectCount uint64
}
