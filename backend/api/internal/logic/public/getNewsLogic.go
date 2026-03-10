// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package public

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNewsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNewsLogic {
	return &GetNewsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNewsLogic) GetNews(req *types.IdReq) (resp *types.NewsItem, err error) {
	// todo: add your logic here and delete this line

	return
}
