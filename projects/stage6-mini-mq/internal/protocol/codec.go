package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

// EncodeRequest 编码请求。
func EncodeRequest(req *Request) ([]byte, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	buf := bytes.NewBuffer(nil)
	_ = binary.Write(buf, binary.BigEndian, uint16(req.APIKey))
	_ = binary.Write(buf, binary.BigEndian, req.CorrelationID)
	clientIDBytes := []byte(req.ClientID)
	_ = binary.Write(buf, binary.BigEndian, uint16(len(clientIDBytes)))
	_, _ = buf.Write(clientIDBytes)
	_ = binary.Write(buf, binary.BigEndian, uint32(len(req.Body)))
	_, _ = buf.Write(req.Body)

	payload := buf.Bytes()
	result := bytes.NewBuffer(nil)
	_ = binary.Write(result, binary.BigEndian, uint32(len(payload)))
	_, _ = result.Write(payload)
	return result.Bytes(), nil
}

// DecodeRequest 解码请求。
func DecodeRequest(reader io.Reader) (*Request, error) {
	var total uint32
	if err := binary.Read(reader, binary.BigEndian, &total); err != nil {
		return nil, err
	}
	payload := make([]byte, total)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}
	buf := bytes.NewReader(payload)
	request := &Request{}
	var api uint16
	if err := binary.Read(buf, binary.BigEndian, &api); err != nil {
		return nil, err
	}
	request.APIKey = RequestType(api)
	if err := binary.Read(buf, binary.BigEndian, &request.CorrelationID); err != nil {
		return nil, err
	}
	var clientIDLen uint16
	if err := binary.Read(buf, binary.BigEndian, &clientIDLen); err != nil {
		return nil, err
	}
	clientID := make([]byte, clientIDLen)
	if _, err := io.ReadFull(buf, clientID); err != nil {
		return nil, err
	}
	request.ClientID = string(clientID)
	var bodyLen uint32
	if err := binary.Read(buf, binary.BigEndian, &bodyLen); err != nil {
		return nil, err
	}
	request.Body = make([]byte, bodyLen)
	if _, err := io.ReadFull(buf, request.Body); err != nil {
		return nil, err
	}
	return request, nil
}

// EncodeResponse 编码响应。
func EncodeResponse(resp *Response) ([]byte, error) {
	if resp == nil {
		return nil, errors.New("response is nil")
	}
	buf := bytes.NewBuffer(nil)
	_ = binary.Write(buf, binary.BigEndian, resp.CorrelationID)
	_ = binary.Write(buf, binary.BigEndian, resp.ErrorCode)
	_ = binary.Write(buf, binary.BigEndian, uint32(len(resp.Body)))
	_, _ = buf.Write(resp.Body)
	payload := buf.Bytes()
	result := bytes.NewBuffer(nil)
	_ = binary.Write(result, binary.BigEndian, uint32(len(payload)))
	_, _ = result.Write(payload)
	return result.Bytes(), nil
}

// DecodeResponse 解码响应。
func DecodeResponse(reader io.Reader) (*Response, error) {
	var total uint32
	if err := binary.Read(reader, binary.BigEndian, &total); err != nil {
		return nil, err
	}
	payload := make([]byte, total)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return nil, err
	}
	buf := bytes.NewReader(payload)
	resp := &Response{}
	if err := binary.Read(buf, binary.BigEndian, &resp.CorrelationID); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.BigEndian, &resp.ErrorCode); err != nil {
		return nil, err
	}
	var bodyLen uint32
	if err := binary.Read(buf, binary.BigEndian, &bodyLen); err != nil {
		return nil, err
	}
	resp.Body = make([]byte, bodyLen)
	if _, err := io.ReadFull(buf, resp.Body); err != nil {
		return nil, err
	}
	return resp, nil
}
