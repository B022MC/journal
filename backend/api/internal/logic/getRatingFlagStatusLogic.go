package logic

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/common/consts"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetRatingFlagStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRatingFlagStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRatingFlagStatusLogic {
	return &GetRatingFlagStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRatingFlagStatusLogic) GetRatingFlagStatus(ratingId int64) (*types.FlagStatusResp, error) {
	return getFlagStatus(l.ctx, l.svcCtx, consts.FlagTargetRating, ratingId)
}
