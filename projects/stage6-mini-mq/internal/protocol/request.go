package protocol

type RequestType uint16

const (
	ProduceRequest RequestType = iota
	FetchRequest
	OffsetCommitRequest
	OffsetFetchRequest
	MetadataRequest
	JoinGroupRequest
	HeartbeatRequest
)

// Request 通用请求。
type Request struct {
	APIKey        RequestType
	CorrelationID int32
	ClientID      string
	Body          []byte
}
