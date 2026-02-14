package protocol

// Response 通用响应。
type Response struct {
	CorrelationID int32
	ErrorCode     int16
	Body          []byte
}
