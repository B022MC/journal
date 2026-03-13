package svc

import (
	"journal/common/dao"
	"journal/model"
	"journal/rpc/admin/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config         config.Config
	DBConn         sqlx.SqlConn
	UserModel      *model.UserModel
	PaperModel     *model.PaperModel
	FlagModel      *model.FlagModel
	AdminRBACModel *model.AdminRBACModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	dao.Register("db", c.DB.MustSqlConf("DB"))
	conn := dao.GetConn("db")

	return &ServiceContext{
		Config:         c,
		DBConn:         conn,
		UserModel:      model.NewUserModel(conn),
		PaperModel:     model.NewPaperModel(conn),
		FlagModel:      model.NewFlagModel(conn),
		AdminRBACModel: model.NewAdminRBACModel(conn),
	}
}
