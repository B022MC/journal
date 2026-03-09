// Code scaffolded by goctl. Safe to edit.

package svc

import (
	"journal/api/internal/config"
	"journal/common/dao"
	"journal/model"
	"journal/rpc/admin/adminClient"
	"journal/rpc/news/client/news"
	"journal/rpc/paper/client/paper"
	"journal/rpc/rating/client/rating"
	"journal/rpc/user/client/user"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	AdminRBAC *model.AdminRBACModel
	UserRpc   user.User
	PaperRpc  paper.Paper
	RatingRpc rating.Rating
	NewsRpc   news.News
	AdminRpc  adminClient.Admin
}

func NewServiceContext(c config.Config) *ServiceContext {
	dao.Register("admin", c.AdminDB.MustSqlConf("AdminDB"))
	adminConn := dao.GetConn("admin")

	return &ServiceContext{
		Config:    c,
		AdminRBAC: model.NewAdminRBACModel(adminConn),
		UserRpc:   user.NewUser(zrpc.MustNewClient(c.UserRpc)),
		PaperRpc:  paper.NewPaper(zrpc.MustNewClient(c.PaperRpc)),
		RatingRpc: rating.NewRating(zrpc.MustNewClient(c.RatingRpc)),
		NewsRpc:   news.NewNews(zrpc.MustNewClient(c.NewsRpc)),
		AdminRpc:  adminClient.NewAdmin(zrpc.MustNewClient(c.AdminRpc)),
	}
}
