// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package public

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"journal/api/internal/logic/public"
	"journal/api/internal/svc"
	"journal/api/internal/types"
)

func SearchPapersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SearchPapersReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := public.NewSearchPapersLogic(r.Context(), svcCtx)
		resp, err := l.SearchPapers(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
