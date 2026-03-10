package logic

import (
	"journal/api/internal/types"
	"journal/common/achievement"
)

func toAchievementBadges(items []*achievement.Badge) []types.AchievementBadge {
	if len(items) == 0 {
		return []types.AchievementBadge{}
	}

	result := make([]types.AchievementBadge, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		result = append(result, types.AchievementBadge{
			Code:        item.Code,
			Name:        item.Name,
			Description: item.Description,
			Tier:        item.Tier,
			UnlockedAt:  item.UnlockedAt.Unix(),
		})
	}
	return result
}
