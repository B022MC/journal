package cachekeys

import "strconv"

const (
	HotPapersPrefix       = "api:papers:hot"
	ContributionPrefix    = "api:user:contribution"
	FlagStatusPrefix      = "api:flags:status"
	PaperModerationPrefix = "api:paper:moderation"
)

func HotPapers(zone string) string {
	if zone == "" {
		return HotPapersPrefix + ":all"
	}
	return HotPapersPrefix + ":" + zone
}

func Contribution(userId int64) string {
	return ContributionPrefix + ":" + strconv.FormatInt(userId, 10)
}

func FlagStatus(targetType string, targetId int64) string {
	return FlagStatusPrefix + ":" + targetType + ":" + strconv.FormatInt(targetId, 10)
}

func PaperModeration(paperId int64) string {
	return PaperModerationPrefix + ":" + strconv.FormatInt(paperId, 10)
}
