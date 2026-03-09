package svc

import (
	"journal/common/dao"
	"journal/model"
	"journal/rpc/paper/internal/config"
)

type ServiceContext struct {
	Config     config.Config
	PaperModel *model.PaperModel
	UserModel  *model.UserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	dao.Register("biz", c.BizDB.MustSqlConf("BizDB"))
	conn := dao.GetConn("biz")
	return &ServiceContext{
		Config:     c,
		PaperModel: model.NewPaperModel(conn),
		UserModel:  model.NewUserModel(conn),
	}
}
