package httpserver

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/app"
	"github.com/EfreetZ/SWAI/projects/stage11-mini-ecommerce/internal/model"
)

// Server HTTP 接口层。
type Server struct {
	app *app.App
	mux *http.ServeMux
}

// New 创建 HTTP 服务。
func New(a *app.App) *Server {
	s := &Server{app: a, mux: http.NewServeMux()}
	s.mux.HandleFunc("/healthz", s.health)
	s.mux.HandleFunc("/api/orders", s.createOrder)
	return s
}

func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": "ok", "message": "success"})
}

type createOrderReq struct {
	UserID int64             `json:"user_id"`
	Items  []model.OrderItem `json:"items"`
}

func (s *Server) createOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{"code": 405, "data": nil, "message": "method not allowed"})
		return
	}
	defer func() { _ = r.Body.Close() }()
	var req createOrderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"code": 400, "data": nil, "message": "invalid request"})
		return
	}
	ctx := r.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	o, p, err := s.app.CreateOrderAndPay(ctx, req.UserID, req.Items)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"code": 500, "data": nil, "message": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"code": 0, "data": map[string]interface{}{"order": o, "payment": p}, "message": "success"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
