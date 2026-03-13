package svc

import (
	"journal/common/achievement"
	"journal/common/dao"
	"journal/common/degradation"
	"journal/model"
	"journal/rpc/paper/internal/config"
	"journal/rpc/paper/internal/search"
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
	SearchService      *search.Service
}

func NewServiceContext(c config.Config) *ServiceContext {
	dao.Register("db", c.DB.MustSqlConf("DB"))
	conn := dao.GetConn("db")
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
		SearchService:      search.NewService(c.Search, paperModel, paperModel),
	}
}
