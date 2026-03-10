package manage

import (
	"net/http"

	"journal/admin-api/internal/logic/manage"
	"journal/admin-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func AdminListKeywordRulesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := manage.NewAdminListKeywordRulesLogic(r.Context(), svcCtx)
		resp, err := l.AdminListKeywordRules()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
