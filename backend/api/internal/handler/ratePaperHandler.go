package handler

import (
	"context"
	"net/http"
	"strconv"

	"journal/api/internal/logic"
	"journal/api/internal/requestmeta"
	"journal/api/internal/svc"
	"journal/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RatePaperHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RatePaperReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// Extract paper ID from path /papers/:id/rate
		var idReq types.IdReq
		if err := httpx.Parse(r, &idReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		_ = strconv.FormatInt(idReq.Id, 10) // ensure it's valid

		ctx := context.WithValue(r.Context(), "requestIP", requestmeta.RequestIP(r))
		ctx = context.WithValue(ctx, "requestUserAgent", requestmeta.UserAgent(r))

		l := logic.NewRatePaperLogic(ctx, svcCtx)
		resp, err := l.RatePaper(&req, idReq.Id)
		if err != nil {
			httpx.ErrorCtx(ctx, w, err)
		} else {
			httpx.OkJsonCtx(ctx, w, resp)
		}
	}
}
