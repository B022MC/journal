package manage

import (
	"net/http"

	"journal/admin-api/internal/logic/manage"
	"journal/admin-api/internal/svc"
	"journal/admin-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func AdminDeleteKeywordRuleHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.IdReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := manage.NewAdminDeleteKeywordRuleLogic(r.Context(), svcCtx)
		resp, err := l.AdminDeleteKeywordRule(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
