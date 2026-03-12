// Code scaffolded by goctl. Safe to edit.

package handler

import (
	"net/http"

	"journal/api/internal/logic"
	"journal/api/internal/svc"
	"journal/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetPublicUserProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.IdReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetPublicUserProfileLogic(r.Context(), svcCtx)
		resp, err := l.GetPublicUserProfile(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
