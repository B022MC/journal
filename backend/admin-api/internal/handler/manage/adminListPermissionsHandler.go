// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package manage

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"journal/admin-api/internal/logic/manage"
	"journal/admin-api/internal/svc"
)

func AdminListPermissionsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := manage.NewAdminListPermissionsLogic(r.Context(), svcCtx)
		resp, err := l.AdminListPermissions()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
