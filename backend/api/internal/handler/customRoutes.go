package handler

import (
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func RegisterCustomHandlers(server *rest.Server) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		httpx.OkJsonCtx(r.Context(), w, map[string]string{
			"service":   "journal-api",
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/api/health",
			Handler: handler,
		},
		{
			Method:  http.MethodGet,
			Path:    "/healthz",
			Handler: handler,
		},
	})
}
