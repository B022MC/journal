package logic

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/common/consts"

	"github.com/zeromicro/go-zero/core/logx"
)

type FlagRatingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFlagRatingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FlagRatingLogic {
	return &FlagRatingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FlagRatingLogic) FlagRating(ratingId int64, req *types.FlagReq) (*types.FlagActionResp, error) {
	return submitFlag(l.ctx, l.svcCtx, consts.FlagTargetRating, ratingId, req)
}
