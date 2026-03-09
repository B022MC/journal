package svc

import (
	"journal/common/dao"
	"journal/model"
	"journal/rpc/paper/internal/config"
)

type ServiceContext struct {
	Config     config.Config
	PaperModel *model.PaperModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	dao.Register("main", c.DB)
	conn := dao.GetConn("main")
	return &ServiceContext{
		Config:     c,
		PaperModel: model.NewPaperModel(conn),
	}
}
