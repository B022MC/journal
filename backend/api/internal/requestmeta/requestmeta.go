package requestmeta

import (
	"net"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RequestIP(r *http.Request) string {
	raw := strings.TrimSpace(httpx.GetRemoteAddr(r))
	if raw == "" {
		return ""
	}

	if strings.Contains(raw, ",") {
		raw = strings.TrimSpace(strings.Split(raw, ",")[0])
	}

	host, _, err := net.SplitHostPort(raw)
	if err == nil {
		return host
	}

	return raw
}

func UserAgent(r *http.Request) string {
	return strings.TrimSpace(r.UserAgent())
}
