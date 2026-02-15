package protocol

import "encoding/json"

// Codec 编解码器接口。
type Codec interface {
	Encode(v interface{}) ([]byte, error)
	Decode(data []byte, v interface{}) error
	Name() string
}

// JSONCodec JSON 编解码。
type JSONCodec struct{}

func (c *JSONCodec) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (c *JSONCodec) Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (c *JSONCodec) Name() string {
	return "json"
}

// GetCodec 获取编解码器。
func GetCodec(t CodecType) Codec {
	switch t {
	case JSON:
		return &JSONCodec{}
	default:
		return &JSONCodec{}
	}
}
