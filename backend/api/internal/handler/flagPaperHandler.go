package handler

import (
	"net/http"

	"journal/api/internal/logic"
	"journal/api/internal/svc"
	"journal/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func FlagPaperHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FlagReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		var idReq types.IdReq
		if err := httpx.Parse(r, &idReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewFlagPaperLogic(r.Context(), svcCtx)
		resp, err := l.FlagPaper(idReq.Id, &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
