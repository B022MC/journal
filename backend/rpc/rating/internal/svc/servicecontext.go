package svc

import (
	"context"

	"journal/common/achievement"
	"journal/common/contribution"
	"journal/common/dao"
	"journal/common/degradation"
	"journal/model"
	"journal/rpc/rating/internal/config"
	"journal/rpc/rating/internal/eventing"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config              config.Config
	RatingModel         *model.RatingModel
	PaperModel          *model.PaperModel
	UserModel           *model.UserModel
	FlagModel           *model.FlagModel
	DegradationEngine   *degradation.Engine
	RatingGuard         *degradation.RatingAnomalyGuard
	ContributionManager *contribution.Manager
	AchievementService  *achievement.Service
	PostRateQueue       *eventing.PostRateQueue
	PostRateProcessor   *eventing.PostRateProcessor
	PostRateConsumer    *eventing.PostRateConsumer
}

func NewServiceContext(c config.Config) *ServiceContext {
	dao.Register("db", c.DB.MustSqlConf("DB"))
	conn := dao.GetConn("db")
	ratingModel := model.NewRatingModel(conn)
	paperModel := model.NewPaperModel(conn)
	userModel := model.NewUserModel(conn)
	flagModel := model.NewFlagModel(conn)
	userAchievementModel := model.NewUserAchievementModel(conn)
	engine := degradation.NewEngine(flagModel, paperModel, userModel)
	ratingGuard := degradation.NewRatingAnomalyGuard(ratingModel, flagModel, engine)
	contributionManager := contribution.NewManager(userModel, paperModel, ratingModel)
	achievementService := achievement.NewService(userAchievementModel, paperModel, ratingModel)

	var postRateQueue *eventing.PostRateQueue
	var postRateConsumer *eventing.PostRateConsumer
	var redisStore *redis.Redis
	if c.Redis.Host != "" {
		redisStore = c.Redis.NewRedis()
		postRateQueue = eventing.NewPostRateQueue(redisStore)
	}

	postRateProcessor := eventing.NewPostRateProcessor(
		paperModel,
		ratingModel,
		contributionManager,
		ratingGuard,
		achievementService,
		redisStore,
	)
	if redisStore != nil && postRateQueue != nil {
		postRateConsumer = eventing.NewPostRateConsumer(redisStore, postRateQueue, postRateProcessor)
	}

	return &ServiceContext{
		Config:              c,
		RatingModel:         ratingModel,
		PaperModel:          paperModel,
		UserModel:           userModel,
		FlagModel:           flagModel,
		DegradationEngine:   engine,
		RatingGuard:         ratingGuard,
		ContributionManager: contributionManager,
		AchievementService:  achievementService,
		PostRateQueue:       postRateQueue,
		PostRateProcessor:   postRateProcessor,
		PostRateConsumer:    postRateConsumer,
	}
}

func (s *ServiceContext) StartBackgroundWorkers(ctx context.Context) {
	if s == nil || s.PostRateConsumer == nil {
		return
	}

	go s.PostRateConsumer.Start(ctx)
}
