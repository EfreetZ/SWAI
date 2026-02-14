package handler

import (
	"encoding/json"
	"net/http"
)

const (
	ErrCodeSuccess       = 0
	ErrCodeBadRequest    = 10001
	ErrCodeUnauthorized  = 10002
	ErrCodeForbidden     = 10003
	ErrCodeNotFound      = 10004
	ErrCodeInternal      = 10005
	ErrCodeUserExists    = 20001
	ErrCodeWrongPassword = 20002
)

// Response 统一 JSON 响应结构。
type Response struct {
	Code    int    `json:"code"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

// Success 返回成功响应。
func Success(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, Response{Code: ErrCodeSuccess, Data: data, Message: "success"})
}

// Error 返回错误响应。
func Error(w http.ResponseWriter, httpCode, bizCode int, message string) {
	writeJSON(w, httpCode, Response{Code: bizCode, Data: nil, Message: message})
}

func writeJSON(w http.ResponseWriter, httpCode int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	_ = json.NewEncoder(w).Encode(resp)
}
