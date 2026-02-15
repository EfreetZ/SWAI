package proxy

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Forward 转发请求到目标地址。
func Forward(w http.ResponseWriter, r *http.Request, targetAddr string, stripPrefix string) error {
	target, err := url.Parse(targetAddr)
	if err != nil {
		return err
	}
	path := r.URL.Path
	if stripPrefix != "" {
		path = strings.TrimPrefix(path, stripPrefix)
		if path == "" {
			path = "/"
		}
	}
	forwardURL := target.String() + path
	if r.URL.RawQuery != "" {
		forwardURL += "?" + r.URL.RawQuery
	}
	bodyBytes, _ := io.ReadAll(r.Body)
	_ = r.Body.Close()
	req, err := http.NewRequestWithContext(r.Context(), r.Method, forwardURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return err
	}
	req.Header = r.Header.Clone()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	for k, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
	return nil
}
