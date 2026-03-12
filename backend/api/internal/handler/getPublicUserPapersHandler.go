// Code scaffolded by goctl. Safe to edit.

package handler

import (
	"net/http"

	"journal/api/internal/logic"
	"journal/api/internal/svc"
	"journal/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetPublicUserPapersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserPageReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetPublicUserPapersLogic(r.Context(), svcCtx)
		resp, err := l.GetPublicUserPapers(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
