package logic

import (
	"context"

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

func (l *UserRatingsLogic) UserRatings(req *types.PageReq) (*types.UserRatingsResp, error) {
	userId := currentUserID(l.ctx)
	rpcResp, err := l.svcCtx.RatingRpc.GetUserRatings(l.ctx, &rating.UserRatingsReq{
		UserId:   userId,
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}

	items := make([]types.RatingItem, 0, len(rpcResp.Items))
	for _, item := range rpcResp.Items {
		items = append(items, types.RatingItem{
			Id:        item.Id,
			PaperId:   item.PaperId,
			UserId:    item.UserId,
			Username:  item.Username,
			Nickname:  item.Nickname,
			Score:     item.Score,
			Comment:   item.Comment,
			CreatedAt: item.CreatedAt,
		})
	}

	return &types.UserRatingsResp{
		Items: items,
		Total: rpcResp.Total,
	}, nil
}
