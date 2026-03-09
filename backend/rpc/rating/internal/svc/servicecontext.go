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
}

func NewServiceContext(c config.Config) *ServiceContext {
	dao.Register("main", c.DB)
	conn := dao.GetConn("main")
	return &ServiceContext{
		Config:      c,
		RatingModel: model.NewRatingModel(conn),
		PaperModel:  model.NewPaperModel(conn),
	}
}
