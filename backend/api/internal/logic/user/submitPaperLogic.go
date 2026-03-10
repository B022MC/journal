// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitPaperLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSubmitPaperLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitPaperLogic {
	return &SubmitPaperLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SubmitPaperLogic) SubmitPaper(req *types.SubmitPaperReq) (resp *types.SubmitPaperResp, err error) {
	// todo: add your logic here and delete this line

	return
}
