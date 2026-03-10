package flagging

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strconv"

	apicache "journal/api/internal/cache"
	"journal/common/consts"
	"journal/common/degradation"
	"journal/model"
)

var (
	ErrTargetNotFound = errors.New("target not found")
	ErrAlreadyFlagged = errors.New("already flagged")
	ErrSelfFlag       = errors.New("self flag")
	ErrInvalidReason  = errors.New("invalid flag reason")
)

type TargetStatus struct {
	Exists           bool
	TargetType       string
	TargetId         int64
	FlagCount        int32
	PendingCount     int32
	WeightedSum      float64
	Quorum           int32
	DegradationLevel int32
}

type Service struct {
	flagModel         *model.FlagModel
	paperModel        *model.PaperModel
	ratingModel       *model.RatingModel
	userModel         *model.UserModel
	engine            *degradation.Engine
	cache             *apicache.Service
	postFlagQueue     *PostFlagQueue
	postFlagProcessor *PostFlagProcessor
}

func NewService(
	flagModel *model.FlagModel,
	paperModel *model.PaperModel,
	ratingModel *model.RatingModel,
	userModel *model.UserModel,
	engine *degradation.Engine,
	cache *apicache.Service,
	postFlagQueue *PostFlagQueue,
	postFlagProcessor *PostFlagProcessor,
) *Service {
	return &Service{
		flagModel:         flagModel,
		paperModel:        paperModel,
		ratingModel:       ratingModel,
		userModel:         userModel,
		engine:            engine,
		cache:             cache,
		postFlagQueue:     postFlagQueue,
		postFlagProcessor: postFlagProcessor,
	}
}

func (s *Service) SubmitFlag(ctx context.Context, targetType string, targetId, reporterId int64, reason, detail string) (int64, *TargetStatus, error) {
	if !isValidReason(reason) {
		return 0, nil, ErrInvalidReason
	}

	reporterScore, err := s.loadContributionScore(ctx, reporterId)
	if err != nil {
		return 0, nil, err
	}

	ownerID, err := s.loadTargetOwner(ctx, targetType, targetId)
	if err != nil {
		return 0, nil, err
	}
	if ownerID == reporterId {
		return 0, nil, ErrSelfFlag
	}

	flagged, err := s.flagModel.HasFlagged(ctx, targetType, targetId, reporterId)
	if err != nil {
		return 0, nil, err
	}
	if flagged {
		return 0, nil, ErrAlreadyFlagged
	}

	flag := &model.Flag{
		TargetType:           targetType,
		TargetId:             targetId,
		ReporterId:           reporterId,
		Reason:               reason,
		Detail:               detail,
		ReporterContribution: reporterScore,
	}
	flagID, err := s.engine.RecordFlag(ctx, flag)
	if err != nil {
		return 0, nil, err
	}

	_ = s.userModel.UpdateLastActive(ctx, reporterId)
	if s.cache != nil {
		_ = s.cache.DeleteFlagStatus(ctx, targetType, targetId)
		if targetType == consts.FlagTargetPaper {
			_ = s.cache.DeletePaperModeration(ctx, targetId)
		}
	}

	postFlagEvent := PostFlagEvent{
		FlagId:     flagID,
		TargetType: targetType,
		TargetId:   targetId,
		ReporterId: reporterId,
	}
	if s.postFlagProcessor == nil {
		if _, err := s.engine.EvaluateDegradation(ctx, targetType, targetId); err != nil {
			return 0, nil, err
		}
	} else if s.postFlagQueue == nil {
		if err := s.postFlagProcessor.HandlePostFlagEvent(ctx, postFlagEvent); err != nil {
			return 0, nil, err
		}
	} else if err := s.postFlagQueue.Enqueue(ctx, postFlagEvent); err != nil {
		if err := s.postFlagProcessor.HandlePostFlagEvent(ctx, postFlagEvent); err != nil {
			return 0, nil, err
		}
	}

	status, err := s.GetFlagStatus(ctx, targetType, targetId)
	if err != nil {
		return flagID, nil, err
	}
	return flagID, status, nil
}

func (s *Service) GetFlagStatus(ctx context.Context, targetType string, targetId int64) (*TargetStatus, error) {
	if s.cache != nil {
		if cached, ok, err := s.cache.GetFlagStatus(ctx, targetType, targetId); err == nil && ok {
			return targetStatusFromCache(cached), nil
		}
	}

	stats, err := s.flagModel.CountByTarget(ctx, targetType, targetId)
	if err != nil {
		return nil, err
	}

	status := &TargetStatus{
		Exists:       true,
		TargetType:   targetType,
		TargetId:     targetId,
		PendingCount: int32(stats.TotalCount),
		WeightedSum:  stats.WeightedSum,
	}

	switch targetType {
	case consts.FlagTargetPaper:
		if s.cache != nil {
			if cached, ok, err := s.cache.GetPaperModeration(ctx, targetId); err == nil && ok {
				status.FlagCount = cached.FlagCount
				status.Quorum = int32(calcPaperQuorum(cached.RatingCount))
				status.DegradationLevel = cached.DegradationLevel
				s.storeFlagStatus(ctx, status)
				return status, nil
			}
		}

		paper, err := s.paperModel.FindByIdPrimary(ctx, targetId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrTargetNotFound
			}
			return nil, err
		}
		status.FlagCount = paper.FlagCount
		status.Quorum = int32(calcPaperQuorum(paper.RatingCount))
		status.DegradationLevel = effectivePaperDegradationLevel(paper.DegradationLevel, stats.TotalCount, stats.WeightedSum, int(status.Quorum))
		s.storePaperModeration(ctx, &apicache.PaperModerationPayload{
			Id:               paper.Id,
			Zone:             paper.Zone,
			Status:           paper.Status,
			RatingCount:      paper.RatingCount,
			FlagCount:        paper.FlagCount,
			DegradationLevel: status.DegradationLevel,
		})
		s.storeFlagStatus(ctx, status)
		return status, nil
	case consts.FlagTargetRating:
		_, err := s.ratingModel.FindByIdPrimary(ctx, targetId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrTargetNotFound
			}
			return nil, err
		}
		flags, err := s.ListFlagsByTarget(ctx, targetType, targetId)
		if err != nil {
			return nil, err
		}
		status.FlagCount = int32(len(flags))
		status.Quorum = 3
		status.DegradationLevel = degradation.DetermineLevel(stats.TotalCount, stats.WeightedSum, int(status.Quorum))
		s.storeFlagStatus(ctx, status)
		return status, nil
	default:
		return nil, ErrTargetNotFound
	}
}

func (s *Service) ListFlagsByTarget(ctx context.Context, targetType string, targetId int64) ([]*model.Flag, error) {
	return s.flagModel.ListByTarget(ctx, targetType, targetId)
}

func (s *Service) loadTargetOwner(ctx context.Context, targetType string, targetId int64) (int64, error) {
	switch targetType {
	case consts.FlagTargetPaper:
		paper, err := s.paperModel.FindByIdPrimary(ctx, targetId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, ErrTargetNotFound
			}
			return 0, err
		}
		return paper.AuthorId, nil
	case consts.FlagTargetRating:
		item, err := s.ratingModel.FindByIdPrimary(ctx, targetId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, ErrTargetNotFound
			}
			return 0, err
		}
		return item.UserId, nil
	default:
		return 0, ErrTargetNotFound
	}
}

func calcPaperQuorum(ratingCount int32) int {
	q := int(math.Sqrt(float64(ratingCount)))
	if q < 3 {
		return 3
	}
	return q
}

func effectivePaperDegradationLevel(currentLevel int32, pendingCount int, weightedSum float64, quorum int) int32 {
	level := degradation.DetermineLevel(pendingCount, weightedSum, quorum)
	if currentLevel > level {
		return currentLevel
	}
	return level
}

func (s *Service) loadContributionScore(ctx context.Context, userId int64) (float64, error) {
	if s.cache != nil {
		if cached, ok, err := s.cache.GetContributionScore(ctx, userId); err == nil && ok {
			if score, parseErr := strconv.ParseFloat(cached, 64); parseErr == nil {
				return score, nil
			}
		}
	}

	user, err := s.userModel.FindByIdPrimary(ctx, userId)
	if err != nil {
		return 0, err
	}

	if s.cache != nil {
		_ = s.cache.SetContributionScore(ctx, userId, strconv.FormatFloat(user.ContributionScore, 'f', -1, 64))
	}

	return user.ContributionScore, nil
}

func isValidReason(reason string) bool {
	switch reason {
	case consts.FlagReasonAbuse,
		consts.FlagReasonSpam,
		consts.FlagReasonPlagiarism,
		consts.FlagReasonSensitive,
		consts.FlagReasonManipulation:
		return true
	default:
		return false
	}
}

func (s *Service) storeFlagStatus(ctx context.Context, status *TargetStatus) {
	if s.cache == nil || status == nil {
		return
	}

	_ = s.cache.SetFlagStatus(ctx, &apicache.FlagStatusPayload{
		Exists:           status.Exists,
		TargetType:       status.TargetType,
		TargetId:         status.TargetId,
		FlagCount:        status.FlagCount,
		PendingCount:     status.PendingCount,
		WeightedSum:      status.WeightedSum,
		Quorum:           status.Quorum,
		DegradationLevel: status.DegradationLevel,
	})
}

func targetStatusFromCache(payload *apicache.FlagStatusPayload) *TargetStatus {
	if payload == nil {
		return nil
	}

	return &TargetStatus{
		Exists:           payload.Exists,
		TargetType:       payload.TargetType,
		TargetId:         payload.TargetId,
		FlagCount:        payload.FlagCount,
		PendingCount:     payload.PendingCount,
		WeightedSum:      payload.WeightedSum,
		Quorum:           payload.Quorum,
		DegradationLevel: payload.DegradationLevel,
	}
}

func (s *Service) storePaperModeration(ctx context.Context, payload *apicache.PaperModerationPayload) {
	if s.cache == nil || payload == nil {
		return
	}

	_ = s.cache.SetPaperModeration(ctx, payload)
}
