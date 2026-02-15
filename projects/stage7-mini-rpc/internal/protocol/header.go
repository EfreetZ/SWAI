package protocol

import "encoding/binary"

const MagicNumber uint16 = 0x4D52

const HeaderSize = 17

type MessageType byte

const (
	Request MessageType = iota
	Response
	Heartbeat
)

type CodecType byte

const (
	JSON CodecType = iota
)

// Header RPC 报文头。
type Header struct {
	Magic         uint16
	Version       byte
	Type          MessageType
	Codec         CodecType
	RequestID     uint64
	PayloadLength uint32
}

// EncodeHeader 编码报文头。
func EncodeHeader(h Header) []byte {
	buf := make([]byte, HeaderSize)
	binary.BigEndian.PutUint16(buf[0:2], h.Magic)
	buf[2] = h.Version
	buf[3] = byte(h.Type)
	buf[4] = byte(h.Codec)
	binary.BigEndian.PutUint64(buf[5:13], h.RequestID)
	binary.BigEndian.PutUint32(buf[13:17], h.PayloadLength)
	return buf
}

// DecodeHeader 解码报文头。
func DecodeHeader(buf []byte) Header {
	return Header{
		Magic:         binary.BigEndian.Uint16(buf[0:2]),
		Version:       buf[2],
		Type:          MessageType(buf[3]),
		Codec:         CodecType(buf[4]),
		RequestID:     binary.BigEndian.Uint64(buf[5:13]),
		PayloadLength: binary.BigEndian.Uint32(buf[13:17]),
	}
}
