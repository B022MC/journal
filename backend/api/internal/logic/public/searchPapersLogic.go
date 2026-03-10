// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package public

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchPapersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchPapersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchPapersLogic {
	return &SearchPapersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchPapersLogic) SearchPapers(req *types.SearchPapersReq) (resp *types.ListPapersResp, err error) {
	// todo: add your logic here and delete this line

	return
}
