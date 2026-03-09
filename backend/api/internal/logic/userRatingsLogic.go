package logic

import (
	"context"
	"encoding/json"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/rating/rating"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserRatingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserRatingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserRatingsLogic {
	return &UserRatingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserRatingsLogic) UserRatings(req *types.PageReq) error {
	userId, _ := l.ctx.Value("userId").(json.Number).Int64()
	_, _ = l.svcCtx.RatingRpc.GetUserRatings(l.ctx, &rating.UserRatingsReq{
		UserId:   userId,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	return nil
}
