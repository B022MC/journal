// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package public

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListNewsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListNewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListNewsLogic {
	return &ListNewsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListNewsLogic) ListNews(req *types.PageReq) (resp *types.ListNewsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
