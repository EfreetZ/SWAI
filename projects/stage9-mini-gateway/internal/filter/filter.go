package filter

import "net/http"

// Handler 网关处理器。
type Handler func(http.ResponseWriter, *http.Request)

// Filter 过滤器。
type Filter func(next Handler) Handler

// Chain 构建过滤链。
func Chain(final Handler, filters ...Filter) Handler {
	h := final
	for i := len(filters) - 1; i >= 0; i-- {
		h = filters[i](h)
	}
	return h
}
