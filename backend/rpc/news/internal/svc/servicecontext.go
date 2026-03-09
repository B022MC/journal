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
	dao.Register("main", c.DB)
	conn := dao.GetConn("main")
	return &ServiceContext{
		Config:    c,
		NewsModel: model.NewNewsModel(conn),
	}
}
