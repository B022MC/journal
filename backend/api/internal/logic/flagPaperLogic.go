package logic

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/common/consts"

	"github.com/zeromicro/go-zero/core/logx"
)

type FlagPaperLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFlagPaperLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FlagPaperLogic {
	return &FlagPaperLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FlagPaperLogic) FlagPaper(paperId int64, req *types.FlagReq) (*types.FlagActionResp, error) {
	return submitFlag(l.ctx, l.svcCtx, consts.FlagTargetPaper, paperId, req)
}
