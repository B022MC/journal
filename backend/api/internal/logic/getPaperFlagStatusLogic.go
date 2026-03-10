package logic

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/common/consts"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPaperFlagStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPaperFlagStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaperFlagStatusLogic {
	return &GetPaperFlagStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPaperFlagStatusLogic) GetPaperFlagStatus(paperId int64) (*types.FlagStatusResp, error) {
	return getFlagStatus(l.ctx, l.svcCtx, consts.FlagTargetPaper, paperId)
}
