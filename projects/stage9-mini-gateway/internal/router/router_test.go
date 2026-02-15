package router

import (
	"net/http/httptest"
	"testing"

	"github.com/EfreetZ/SWAI/projects/stage9-mini-gateway/internal/config"
)

func TestMatch(t *testing.T) {
	r := NewRouter([]config.RouteConfig{{Name: "a", Prefix: "/api", Methods: []string{"GET"}}})
	req := httptest.NewRequest("GET", "/api/users", nil)
	_, ok := r.Match(req)
	if !ok {
		t.Fatal("expected matched route")
	}
}
