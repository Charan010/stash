package checksum

import "hash/crc32"

func CRC32(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}
