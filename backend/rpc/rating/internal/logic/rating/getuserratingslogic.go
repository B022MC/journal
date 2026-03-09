package ratinglogic

import (
	"context"

	"journal/rpc/rating/internal/svc"
	"journal/rpc/rating/rating"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserRatingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserRatingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserRatingsLogic {
	return &GetUserRatingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserRatingsLogic) GetUserRatings(in *rating.UserRatingsReq) (*rating.UserRatingsResp, error) {
	page := int(in.Page)
	pageSize := int(in.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := l.svcCtx.RatingModel.ListByUser(l.ctx, in.UserId, page, pageSize)
	if err != nil {
		return nil, err
	}

	ratingItems := make([]*rating.RatingItem, 0, len(items))
	for _, r := range items {
		ratingItems = append(ratingItems, &rating.RatingItem{
			Id:        r.Id,
			PaperId:   r.PaperId,
			UserId:    r.UserId,
			Username:  r.Username,
			Nickname:  r.Nickname,
			Score:     r.Score,
			Comment:   r.Comment,
			CreatedAt: r.CreatedAt.Unix(),
		})
	}

	return &rating.UserRatingsResp{
		Items: ratingItems,
		Total: total,
	}, nil
}
