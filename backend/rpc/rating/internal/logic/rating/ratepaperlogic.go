package ratinglogic

import (
	"context"
	"errors"

	"journal/model"
	"journal/rpc/rating/internal/svc"
	"journal/rpc/rating/rating"

	"github.com/zeromicro/go-zero/core/logx"
)

type RatePaperLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRatePaperLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RatePaperLogic {
	return &RatePaperLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RatePaperLogic) RatePaper(in *rating.RatePaperReq) (*rating.RateResp, error) {
	if in.Score < 1 || in.Score > 10 {
		return nil, errors.New("score must be between 1 and 10")
	}

	// Check paper exists and user isn't the author
	p, err := l.svcCtx.PaperModel.FindById(l.ctx, in.PaperId)
	if err != nil {
		return nil, errors.New("paper not found")
	}
	if p.AuthorId == in.UserId {
		return nil, errors.New("cannot rate your own paper")
	}

	// Upsert rating
	_, err = l.svcCtx.RatingModel.Upsert(l.ctx, &model.Rating{
		PaperId: in.PaperId,
		UserId:  in.UserId,
		Score:   in.Score,
		Comment: in.Comment,
	})
	if err != nil {
		return nil, err
	}

	// Recalculate paper stats
	avgScore, count, controversy, err := l.svcCtx.RatingModel.GetPaperRatingStats(l.ctx, in.PaperId)
	if err != nil {
		return nil, err
	}

	// Update paper scores
	err = l.svcCtx.PaperModel.UpdateScores(l.ctx, in.PaperId, avgScore, count, controversy)
	if err != nil {
		return nil, err
	}

	newShitScore := model.CalcShitScore(avgScore, count, p.ViewCount, controversy)

	return &rating.RateResp{
		Success:        true,
		Message:        "rating submitted",
		NewAvgRating:   avgScore,
		NewRatingCount: count,
		NewShitScore:   newShitScore,
	}, nil
}
