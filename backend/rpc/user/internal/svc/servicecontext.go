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
	dao.Register("main", c.DB)
	conn := dao.GetConn("main")
	return &ServiceContext{
		Config:    c,
		UserModel: model.NewUserModel(conn),
	}
}
