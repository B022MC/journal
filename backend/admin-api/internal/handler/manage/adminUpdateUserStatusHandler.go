// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package manage

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"journal/admin-api/internal/logic/manage"
	"journal/admin-api/internal/svc"
	"journal/admin-api/internal/types"
)

func AdminUpdateUserStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateUserStatusReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := manage.NewAdminUpdateUserStatusLogic(r.Context(), svcCtx)
		resp, err := l.AdminUpdateUserStatus(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
