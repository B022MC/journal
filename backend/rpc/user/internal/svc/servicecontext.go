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
	dao.Register("db", c.DB.MustSqlConf("DB"))
	conn := dao.GetConn("db")
	return &ServiceContext{
		Config:    c,
		UserModel: model.NewUserModel(conn),
	}
}
