package cache

import (
	"context"
	"encoding/json"
	"time"

	"journal/api/internal/types"
	"journal/common/cachekeys"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	HotPapersLimit       = 100
	hotPapersTTL         = 5 * time.Minute
	contributionScoreTTL = time.Hour
	flagStatusTTL        = 10 * time.Minute
	paperModerationTTL   = 30 * time.Minute
)

type Service struct {
	store *redis.Redis
}

type FlagStatusPayload struct {
	Exists           bool    `json:"exists"`
	TargetType       string  `json:"target_type"`
	TargetId         int64   `json:"target_id"`
	FlagCount        int32   `json:"flag_count"`
	PendingCount     int32   `json:"pending_count"`
	WeightedSum      float64 `json:"weighted_sum"`
	Quorum           int32   `json:"quorum"`
	DegradationLevel int32   `json:"degradation_level"`
}

type PaperModerationPayload struct {
	Id               int64  `json:"id"`
	Zone             string `json:"zone"`
	Status           int32  `json:"status"`
	RatingCount      int32  `json:"rating_count"`
	FlagCount        int32  `json:"flag_count"`
	DegradationLevel int32  `json:"degradation_level"`
}

func NewService(store *redis.Redis) *Service {
	return &Service{store: store}
}

func (s *Service) GetHotPapers(ctx context.Context, zone string) (*types.ListPapersResp, bool, error) {
	if s.store == nil {
		return nil, false, nil
	}

	value, err := s.store.GetCtx(ctx, hotPapersKey(zone))
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, err
	}

	var payload types.ListPapersResp
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return nil, false, err
	}
	return &payload, true, nil
}

func (s *Service) SetHotPapers(ctx context.Context, zone string, payload *types.ListPapersResp) error {
	if s.store == nil || payload == nil {
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return s.store.SetexCtx(ctx, hotPapersKey(zone), string(body), int(hotPapersTTL.Seconds()))
}

func (s *Service) InvalidateHotPapers(ctx context.Context, zones ...string) error {
	if s.store == nil {
		return nil
	}

	keys := []string{hotPapersKey("")}
	seen := map[string]struct{}{
		keys[0]: {},
	}
	for _, zone := range zones {
		key := hotPapersKey(zone)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}

	_, err := s.store.DelCtx(ctx, keys...)
	return err
}

func (s *Service) GetContributionScore(ctx context.Context, userId int64) (string, bool, error) {
	if s.store == nil || userId <= 0 {
		return "", false, nil
	}

	value, err := s.store.GetCtx(ctx, contributionKey(userId))
	if err != nil {
		if err == redis.Nil {
			return "", false, nil
		}
		return "", false, err
	}
	return value, true, nil
}

func (s *Service) SetContributionScore(ctx context.Context, userId int64, score string) error {
	if s.store == nil || userId <= 0 || score == "" {
		return nil
	}
	return s.store.SetexCtx(ctx, contributionKey(userId), score, int(contributionScoreTTL.Seconds()))
}

func (s *Service) GetOrLoadContributionScore(ctx context.Context, userId int64, loader func() (string, error)) (string, error) {
	if score, ok, err := s.GetContributionScore(ctx, userId); err != nil {
		return "", err
	} else if ok {
		return score, nil
	}

	score, err := loader()
	if err != nil {
		return "", err
	}
	if err := s.SetContributionScore(ctx, userId, score); err != nil {
		return "", err
	}
	return score, nil
}

func (s *Service) DeleteContributionScores(ctx context.Context, userIds ...int64) error {
	if s.store == nil {
		return nil
	}

	keys := make([]string, 0, len(userIds))
	seen := make(map[string]struct{}, len(userIds))
	for _, userId := range userIds {
		if userId <= 0 {
			continue
		}
		key := contributionKey(userId)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return nil
	}

	_, err := s.store.DelCtx(ctx, keys...)
	return err
}

func (s *Service) GetFlagStatus(ctx context.Context, targetType string, targetId int64) (*FlagStatusPayload, bool, error) {
	if s.store == nil || targetType == "" || targetId <= 0 {
		return nil, false, nil
	}

	value, err := s.store.GetCtx(ctx, flagStatusKey(targetType, targetId))
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, err
	}

	var payload FlagStatusPayload
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return nil, false, err
	}
	return &payload, true, nil
}

func (s *Service) SetFlagStatus(ctx context.Context, payload *FlagStatusPayload) error {
	if s.store == nil || payload == nil || payload.TargetType == "" || payload.TargetId <= 0 {
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return s.store.SetexCtx(ctx, flagStatusKey(payload.TargetType, payload.TargetId), string(body), int(flagStatusTTL.Seconds()))
}

func (s *Service) DeleteFlagStatus(ctx context.Context, targetType string, targetId int64) error {
	if s.store == nil || targetType == "" || targetId <= 0 {
		return nil
	}

	_, err := s.store.DelCtx(ctx, flagStatusKey(targetType, targetId))
	return err
}

func (s *Service) GetPaperModeration(ctx context.Context, paperId int64) (*PaperModerationPayload, bool, error) {
	if s.store == nil || paperId <= 0 {
		return nil, false, nil
	}

	value, err := s.store.GetCtx(ctx, paperModerationKey(paperId))
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, err
	}

	var payload PaperModerationPayload
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return nil, false, err
	}
	return &payload, true, nil
}

func (s *Service) SetPaperModeration(ctx context.Context, payload *PaperModerationPayload) error {
	if s.store == nil || payload == nil || payload.Id <= 0 {
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return s.store.SetexCtx(ctx, paperModerationKey(payload.Id), string(body), int(paperModerationTTL.Seconds()))
}

func (s *Service) DeletePaperModeration(ctx context.Context, paperIds ...int64) error {
	if s.store == nil {
		return nil
	}

	keys := make([]string, 0, len(paperIds))
	seen := make(map[string]struct{}, len(paperIds))
	for _, paperId := range paperIds {
		if paperId <= 0 {
			continue
		}
		key := paperModerationKey(paperId)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return nil
	}

	_, err := s.store.DelCtx(ctx, keys...)
	return err
}

func hotPapersKey(zone string) string {
	return cachekeys.HotPapers(zone)
}

func contributionKey(userId int64) string {
	return cachekeys.Contribution(userId)
}

func flagStatusKey(targetType string, targetId int64) string {
	return cachekeys.FlagStatus(targetType, targetId)
}

func paperModerationKey(paperId int64) string {
	return cachekeys.PaperModeration(paperId)
}
