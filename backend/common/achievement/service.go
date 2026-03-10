package achievement

import (
	"context"
	"sort"
	"time"

	"journal/common/consts"
	"journal/model"
)

type Definition struct {
	Code        string
	Name        string
	Description string
	Tier        string
}

type Badge struct {
	Code        string
	Name        string
	Description string
	Tier        string
	UnlockedAt  time.Time
}

type metrics struct {
	PaperCount         int64
	SedimentPaperCount int64
	ReviewCount        int64
}

var definitions = []Definition{
	{
		Code:        "first_submission",
		Name:        "First Submission",
		Description: "Submitted the first paper to the journal.",
		Tier:        "bronze",
	},
	{
		Code:        "sediment_breakthrough",
		Name:        "Sediment Breakthrough",
		Description: "Had a paper promoted into Sediment.",
		Tier:        "silver",
	},
	{
		Code:        "reviewer_century",
		Name:        "Reviewer Century",
		Description: "Completed 100 reviews.",
		Tier:        "gold",
	},
}

var definitionsByCode = func() map[string]Definition {
	result := make(map[string]Definition, len(definitions))
	for _, item := range definitions {
		result[item.Code] = item
	}
	return result
}()

type Service struct {
	userAchievementModel *model.UserAchievementModel
	paperModel           *model.PaperModel
	ratingModel          *model.RatingModel
}

func NewService(
	userAchievementModel *model.UserAchievementModel,
	paperModel *model.PaperModel,
	ratingModel *model.RatingModel,
) *Service {
	return &Service{
		userAchievementModel: userAchievementModel,
		paperModel:           paperModel,
		ratingModel:          ratingModel,
	}
}

func (s *Service) SyncUser(ctx context.Context, userId int64) error {
	if s == nil || s.userAchievementModel == nil || userId <= 0 {
		return nil
	}

	metrics, err := s.loadMetrics(ctx, userId)
	if err != nil {
		return err
	}

	for _, code := range eligibleAchievementCodes(metrics) {
		if err := s.userAchievementModel.InsertIgnore(ctx, userId, code); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ListByUser(ctx context.Context, userId int64) ([]*Badge, error) {
	if s == nil || s.userAchievementModel == nil || userId <= 0 {
		return []*Badge{}, nil
	}

	items, err := s.userAchievementModel.ListByUserPrimary(ctx, userId)
	if err != nil {
		return nil, err
	}

	badges := make([]*Badge, 0, len(items))
	for _, item := range items {
		def, ok := definitionsByCode[item.Code]
		if !ok {
			continue
		}
		badges = append(badges, &Badge{
			Code:        def.Code,
			Name:        def.Name,
			Description: def.Description,
			Tier:        def.Tier,
			UnlockedAt:  item.UnlockedAt,
		})
	}

	sort.SliceStable(badges, func(i, j int) bool {
		if badges[i].UnlockedAt.Equal(badges[j].UnlockedAt) {
			return badges[i].Code < badges[j].Code
		}
		return badges[i].UnlockedAt.Before(badges[j].UnlockedAt)
	})
	return badges, nil
}

func (s *Service) SyncAndList(ctx context.Context, userId int64) ([]*Badge, error) {
	if err := s.SyncUser(ctx, userId); err != nil {
		return nil, err
	}
	return s.ListByUser(ctx, userId)
}

func (s *Service) loadMetrics(ctx context.Context, userId int64) (*metrics, error) {
	result := &metrics{}

	if s.paperModel != nil {
		paperCount, err := s.paperModel.CountByAuthor(ctx, userId)
		if err != nil {
			return nil, err
		}
		sedimentCount, err := s.paperModel.CountByAuthorZone(ctx, userId, consts.PaperZoneSediment)
		if err != nil {
			return nil, err
		}
		result.PaperCount = paperCount
		result.SedimentPaperCount = sedimentCount
	}

	if s.ratingModel != nil {
		reviewCount, err := s.ratingModel.CountByUser(ctx, userId)
		if err != nil {
			return nil, err
		}
		result.ReviewCount = reviewCount
	}

	return result, nil
}

func eligibleAchievementCodes(metrics *metrics) []string {
	if metrics == nil {
		return nil
	}

	result := make([]string, 0, len(definitions))
	if metrics.PaperCount >= 1 {
		result = append(result, "first_submission")
	}
	if metrics.SedimentPaperCount >= 1 {
		result = append(result, "sediment_breakthrough")
	}
	if metrics.ReviewCount >= 100 {
		result = append(result, "reviewer_century")
	}
	return result
}
