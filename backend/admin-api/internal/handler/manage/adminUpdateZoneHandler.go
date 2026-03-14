package manage

import (
	"net/http"

	"journal/admin-api/internal/logic/manage"
	"journal/admin-api/internal/svc"
	"journal/admin-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func AdminUpdateZoneHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateZoneReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		var idReq types.IdReq
		if err := httpx.ParsePath(r, &idReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := manage.NewAdminUpdateZoneLogic(r.Context(), svcCtx)
		resp, err := l.AdminUpdateZone(idReq.Id, &req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
