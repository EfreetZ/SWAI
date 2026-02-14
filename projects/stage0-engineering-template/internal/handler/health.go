// Package handler 提供 HTTP 请求处理器
package handler

import (
	"net/http"

	"github.com/EfreetZ/SWAI/projects/stage0-engineering-template/pkg/response"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

// NewHealthHandler 创建健康检查处理器实例
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// healthResponse 健康检查响应数据
type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// Check 健康检查端点：GET /health
// 返回服务运行状态，供 Docker healthcheck / 负载均衡器探活使用
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, response.CodeBadRequest, "method not allowed")
		return
	}

	response.Success(w, healthResponse{
		Status:  "ok",
		Version: "0.1.0",
	})
}
