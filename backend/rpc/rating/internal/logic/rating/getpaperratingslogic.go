package ratinglogic

import (
	"context"

	"journal/rpc/rating/internal/svc"
	"journal/rpc/rating/rating"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPaperRatingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPaperRatingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaperRatingsLogic {
	return &GetPaperRatingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPaperRatingsLogic) GetPaperRatings(in *rating.PaperRatingsReq) (*rating.PaperRatingsResp, error) {
	page := int(in.Page)
	pageSize := int(in.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := l.svcCtx.RatingModel.ListByPaper(l.ctx, in.PaperId, page, pageSize)
	if err != nil {
		return nil, err
	}

	avgScore, _, _, _ := l.svcCtx.RatingModel.GetPaperRatingStats(l.ctx, in.PaperId)

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

	return &rating.PaperRatingsResp{
		Items:    ratingItems,
		Total:    total,
		AvgScore: avgScore,
	}, nil
}
