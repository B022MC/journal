package svc

import (
	"journal/common/achievement"
	"journal/common/dao"
	"journal/common/degradation"
	"journal/model"
	"journal/rpc/paper/internal/config"
)

type ServiceContext struct {
	Config             config.Config
	PaperModel         *model.PaperModel
	UserModel          *model.UserModel
	FlagModel          *model.FlagModel
	KeywordFilter      *degradation.KeywordFilter
	KeywordRuleModel   *model.KeywordRuleModel
	DegradationEngine  *degradation.Engine
	AchievementService *achievement.Service
}

func NewServiceContext(c config.Config) *ServiceContext {
	dao.Register("biz", c.BizDB.MustSqlConf("BizDB"))
	conn := dao.GetConn("biz")
	redisClient := c.Redis.NewRedis()
	keywordRuleModel := model.NewKeywordRuleModel(conn)
	flagModel := model.NewFlagModel(conn)
	paperModel := model.NewPaperModel(conn)
	ratingModel := model.NewRatingModel(conn)
	userModel := model.NewUserModel(conn)
	userAchievementModel := model.NewUserAchievementModel(conn)
	return &ServiceContext{
		Config:             c,
		PaperModel:         paperModel,
		UserModel:          userModel,
		FlagModel:          flagModel,
		KeywordRuleModel:   keywordRuleModel,
		KeywordFilter:      degradation.NewKeywordFilter(keywordRuleModel, redisClient),
		DegradationEngine:  degradation.NewEngine(flagModel, paperModel, userModel),
		AchievementService: achievement.NewService(userAchievementModel, paperModel, ratingModel),
	}
}
