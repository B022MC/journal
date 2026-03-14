package handler

import (
	"net/http"

	"journal/api/internal/logic"
	"journal/api/internal/svc"
	"journal/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetPaperRatingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pageReq types.PageReq
		if err := httpx.Parse(r, &pageReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		var idReq types.IdReq
		if err := httpx.ParsePath(r, &idReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetPaperRatingsLogic(r.Context(), svcCtx)
		resp, err := l.GetPaperRatings(idReq.Id, &pageReq)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
