package svc

import (
	"journal/common/dao"
	"journal/model"
	"journal/rpc/admin/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config         config.Config
	BizConn        sqlx.SqlConn
	AdminConn      sqlx.SqlConn
	UserModel      *model.UserModel
	PaperModel     *model.PaperModel
	FlagModel      *model.FlagModel
	AdminRBACModel *model.AdminRBACModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	dao.Register("biz", c.BizDB.MustSqlConf("BizDB"))
	dao.Register("admin", c.AdminDB.MustSqlConf("AdminDB"))

	bizConn := dao.GetConn("biz")
	adminConn := dao.GetConn("admin")

	return &ServiceContext{
		Config:         c,
		BizConn:        bizConn,
		AdminConn:      adminConn,
		UserModel:      model.NewUserModel(bizConn),
		PaperModel:     model.NewPaperModel(bizConn),
		FlagModel:      model.NewFlagModel(bizConn),
		AdminRBACModel: model.NewAdminRBACModel(adminConn),
	}
}
