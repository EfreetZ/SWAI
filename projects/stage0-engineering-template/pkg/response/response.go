// Package response 提供标准化 HTTP JSON 响应格式
// 所有接口统一返回 { code, data, message } 结构
package response

import (
	"encoding/json"
	"net/http"
)

// Response 标准 JSON 响应结构
type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// Success 返回成功响应，HTTP 200 + code=0
func Success(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, &Response{
		Code:    0,
		Data:    data,
		Message: "success",
	})
}

// Error 返回错误响应，自定义 HTTP 状态码和业务错误码
func Error(w http.ResponseWriter, httpCode int, bizCode int, msg string) {
	writeJSON(w, httpCode, &Response{
		Code:    bizCode,
		Data:    nil,
		Message: msg,
	})
}

// writeJSON 将响应序列化为 JSON 并写入 ResponseWriter
func writeJSON(w http.ResponseWriter, httpCode int, resp *Response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpCode)
	_ = json.NewEncoder(w).Encode(resp)
}

// 业务错误码定义
const (
	CodeSuccess      = 0
	CodeBadRequest   = 10001
	CodeUnauthorized = 10002
	CodeForbidden    = 10003
	CodeNotFound     = 10004
	CodeInternal     = 10005
)
