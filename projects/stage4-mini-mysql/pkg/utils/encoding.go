package utils

import "encoding/binary"

// Uint32ToBytes 将 uint32 编码为字节。
func Uint32ToBytes(v uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, v)
	return buf
}

// BytesToUint32 将字节解析为 uint32。
func BytesToUint32(data []byte) uint32 {
	if len(data) < 4 {
		return 0
	}
	return binary.BigEndian.Uint32(data)
}
