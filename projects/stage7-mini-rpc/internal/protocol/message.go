package protocol

// RPCRequest RPC 请求。
type RPCRequest struct {
	ServiceName string            `json:"service_name"`
	MethodName  string            `json:"method_name"`
	Args        []byte            `json:"args"`
	Metadata    map[string]string `json:"metadata"`
}

// RPCResponse RPC 响应。
type RPCResponse struct {
	RequestID uint64            `json:"request_id"`
	Error     string            `json:"error"`
	Data      []byte            `json:"data"`
	Metadata  map[string]string `json:"metadata"`
}
