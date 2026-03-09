package svc

import (
	"journal/common/dao"
	"journal/model"
	"journal/rpc/user/internal/config"
)

type ServiceContext struct {
	Config    config.Config
	UserModel *model.UserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	dao.Register("biz", c.BizDB.MustSqlConf("BizDB"))
	conn := dao.GetConn("biz")
	return &ServiceContext{
		Config:    c,
		UserModel: model.NewUserModel(conn),
	}
}
