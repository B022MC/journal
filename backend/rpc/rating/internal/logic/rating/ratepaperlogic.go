package ratinglogic

import (
	"context"
	"errors"

	"journal/common/contribution"
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

	reviewer, err := l.svcCtx.UserModel.FindByIdPrimary(l.ctx, in.UserId)
	if err != nil {
		return nil, errors.New("reviewer not found")
	}

	hadRated, err := l.svcCtx.RatingModel.HasRated(l.ctx, in.PaperId, in.UserId)
	if err != nil {
		return nil, err
	}

	// Upsert rating
	_, err = l.svcCtx.RatingModel.Upsert(l.ctx, &model.Rating{
		PaperId:        in.PaperId,
		UserId:         in.UserId,
		Score:          in.Score,
		Comment:        in.Comment,
		ReviewerWeight: contribution.ReviewerWeightForContribution(reviewer.ContributionScore),
	})
	if err != nil {
		return nil, err
	}

	if !hadRated {
		if err := l.svcCtx.UserModel.IncrReviewCount30d(l.ctx, in.UserId); err != nil {
			return nil, err
		}
	}
	if err := l.svcCtx.UserModel.UpdateLastActive(l.ctx, in.UserId); err != nil {
		return nil, err
	}

	// Recalculate paper stats
	avgScore, count, controversy, err := l.svcCtx.RatingModel.GetPaperRatingStats(l.ctx, in.PaperId)
	if err != nil {
		return nil, err
	}
	weightedStats, err := l.svcCtx.RatingModel.GetWeightedRatingStats(l.ctx, in.PaperId)
	if err != nil {
		return nil, err
	}

	if err := l.svcCtx.PaperModel.UpdateScoresV2(
		l.ctx,
		in.PaperId,
		avgScore,
		count,
		p.ViewCount,
		controversy,
		weightedStats.WeightedAvg,
		weightedStats.AvgReviewerAuth,
		p.CreatedAt,
	); err != nil {
		return nil, err
	}

	contribManager := contribution.NewManager(l.svcCtx.UserModel, l.svcCtx.PaperModel, l.svcCtx.RatingModel)
	if _, _, err := contribManager.SyncUser(l.ctx, in.UserId); err != nil {
		return nil, err
	}
	if _, _, err := contribManager.SyncUser(l.ctx, p.AuthorId); err != nil {
		return nil, err
	}

	newShitScore := model.CalcShitScoreV2(
		weightedStats.WeightedAvg,
		count,
		p.ViewCount,
		controversy,
		weightedStats.AvgReviewerAuth,
		p.CreatedAt,
	)

	return &rating.RateResp{
		Success:        true,
		Message:        "rating submitted",
		NewAvgRating:   avgScore,
		NewRatingCount: count,
		NewShitScore:   newShitScore,
	}, nil
}
