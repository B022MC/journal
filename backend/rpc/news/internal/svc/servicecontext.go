package svc

import (
	"journal/common/dao"
	"journal/model"
	"journal/rpc/news/internal/config"
)

type ServiceContext struct {
	Config    config.Config
	NewsModel *model.NewsModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	dao.Register("biz", c.BizDB.MustSqlConf("BizDB"))
	conn := dao.GetConn("biz")
	return &ServiceContext{
		Config:    c,
		NewsModel: model.NewNewsModel(conn),
	}
}
