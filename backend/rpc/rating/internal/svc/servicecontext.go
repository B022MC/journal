package svc

import (
	"journal/common/dao"
	"journal/model"
	"journal/rpc/rating/internal/config"
)

type ServiceContext struct {
	Config      config.Config
	RatingModel *model.RatingModel
	PaperModel  *model.PaperModel
	UserModel   *model.UserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	dao.Register("biz", c.BizDB.MustSqlConf("BizDB"))
	conn := dao.GetConn("biz")
	return &ServiceContext{
		Config:      c,
		RatingModel: model.NewRatingModel(conn),
		PaperModel:  model.NewPaperModel(conn),
		UserModel:   model.NewUserModel(conn),
	}
}
