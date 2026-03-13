// Code scaffolded by goctl. Safe to edit.

package svc

import (
	"context"

	apicache "journal/api/internal/cache"
	"journal/api/internal/config"
	"journal/api/internal/flagging"
	apimiddleware "journal/api/internal/middleware"
	"journal/common/achievement"
	"journal/common/dao"
	"journal/common/degradation"
	"journal/common/ratelimit"
	"journal/model"
	"journal/rpc/admin/adminClient"
	"journal/rpc/news/client/news"
	"journal/rpc/paper/client/paper"
	"journal/rpc/rating/client/rating"
	"journal/rpc/user/client/user"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config             config.Config
	AdminRBAC          *model.AdminRBACModel
	UserModel          *model.UserModel
	PaperModel         *model.PaperModel
	RatingModel        *model.RatingModel
	RatingGuard        *degradation.RatingAnomalyGuard
	Cache              *apicache.Service
	AchievementService *achievement.Service
	FlagService        *flagging.Service
	PostFlagQueue      *flagging.PostFlagQueue
	PostFlagProcessor  *flagging.PostFlagProcessor
	PostFlagConsumer   *flagging.PostFlagConsumer
	UserRpc            user.User
	PaperRpc           paper.Paper
	RatingRpc          rating.Rating
	NewsRpc            news.News
	AdminRpc           adminClient.Admin
	RateLimit          rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	dao.Register("db", c.DB.MustSqlConf("DB"))
	conn := dao.GetConn("db")
	var redisStore *redis.Redis
	if c.Redis.Host != "" {
		redisStore = c.Redis.NewRedis()
	}
	userModel := model.NewUserModel(conn)
	paperModel := model.NewPaperModel(conn)
	ratingModel := model.NewRatingModel(conn)
	flagModel := model.NewFlagModel(conn)
	userAchievementModel := model.NewUserAchievementModel(conn)
	engine := degradation.NewEngine(flagModel, paperModel, userModel)
	cacheService := apicache.NewService(redisStore)
	achievementService := achievement.NewService(userAchievementModel, paperModel, ratingModel)
	var postFlagQueue *flagging.PostFlagQueue
	var postFlagConsumer *flagging.PostFlagConsumer
	if redisStore != nil {
		postFlagQueue = flagging.NewPostFlagQueue(redisStore)
	}
	postFlagProcessor := flagging.NewPostFlagProcessor(engine, paperModel, redisStore)
	if redisStore != nil && postFlagQueue != nil {
		postFlagConsumer = flagging.NewPostFlagConsumer(redisStore, postFlagQueue, postFlagProcessor)
	}

	return &ServiceContext{
		Config:             c,
		AdminRBAC:          model.NewAdminRBACModel(conn),
		UserModel:          userModel,
		PaperModel:         paperModel,
		RatingModel:        ratingModel,
		RatingGuard:        degradation.NewRatingAnomalyGuard(ratingModel, flagModel, engine),
		Cache:              cacheService,
		AchievementService: achievementService,
		FlagService:        flagging.NewService(flagModel, paperModel, ratingModel, userModel, engine, cacheService, postFlagQueue, postFlagProcessor),
		PostFlagQueue:      postFlagQueue,
		PostFlagProcessor:  postFlagProcessor,
		PostFlagConsumer:   postFlagConsumer,
		UserRpc:            user.NewUser(zrpc.MustNewClient(c.UserRpc)),
		PaperRpc:           paper.NewPaper(zrpc.MustNewClient(c.PaperRpc)),
		RatingRpc:          rating.NewRating(zrpc.MustNewClient(c.RatingRpc)),
		NewsRpc:            news.NewNews(zrpc.MustNewClient(c.NewsRpc)),
		AdminRpc:           adminClient.NewAdmin(zrpc.MustNewClient(c.AdminRpc)),
		RateLimit: apimiddleware.NewRateLimitMiddleware(
			ratelimit.NewSearchLimiter(redisStore),
			ratelimit.NewRatingLimiter(redisStore),
			ratelimit.NewFlagLimiter(redisStore),
		),
	}
}

func (s *ServiceContext) StartBackgroundWorkers(ctx context.Context) {
	if s == nil || s.PostFlagConsumer == nil {
		return
	}

	go s.PostFlagConsumer.Start(ctx)
}
