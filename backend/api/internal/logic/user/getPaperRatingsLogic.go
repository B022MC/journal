// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPaperRatingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPaperRatingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaperRatingsLogic {
	return &GetPaperRatingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPaperRatingsLogic) GetPaperRatings(req *types.PageReq) (resp *types.PaperRatingsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
