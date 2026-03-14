package svc

import (
	"journal/common/dao"
	"journal/common/degradation"
	"journal/model"
	"journal/rpc/news/internal/config"
)

type ServiceContext struct {
	Config           config.Config
	NewsModel        *model.NewsModel
	KeywordRuleModel *model.KeywordRuleModel
	KeywordFilter    *degradation.KeywordFilter
}

func NewServiceContext(c config.Config) *ServiceContext {
	dao.Register("db", c.DB.MustSqlConf("DB"))
	conn := dao.GetConn("db")
	redisClient := c.CacheRedis.NewRedis()
	keywordRuleModel := model.NewKeywordRuleModel(conn)
	return &ServiceContext{
		Config:           c,
		NewsModel:        model.NewNewsModel(conn),
		KeywordRuleModel: keywordRuleModel,
		KeywordFilter:    degradation.NewKeywordFilter(keywordRuleModel, redisClient),
	}
}
