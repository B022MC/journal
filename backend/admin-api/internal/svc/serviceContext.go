// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"journal/admin-api/internal/config"
	"journal/common/dao"
	"journal/common/degradation"
	"journal/model"
	"journal/rpc/admin/adminClient"
	"journal/rpc/news/client/news"
	"journal/rpc/paper/client/paper"
	"journal/rpc/user/client/user"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config           config.Config
	AdminRBAC        *model.AdminRBACModel
	UserModel        *model.UserModel
	KeywordRuleModel *model.KeywordRuleModel
	KeywordFilter    *degradation.KeywordFilter
	AdminRpc         adminClient.Admin
	NewsRpc          news.News
	PaperRpc         paper.Paper
	UserRpc          user.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	dao.Register("biz", c.BizDB.MustSqlConf("BizDB"))
	dao.Register("admin", c.AdminDB.MustSqlConf("AdminDB"))
	bizConn := dao.GetConn("biz")
	adminConn := dao.GetConn("admin")
	redisClient := c.Redis.NewRedis()
	userModel := model.NewUserModel(bizConn)
	keywordRuleModel := model.NewKeywordRuleModel(bizConn)

	return &ServiceContext{
		Config:           c,
		AdminRBAC:        model.NewAdminRBACModel(adminConn),
		UserModel:        userModel,
		KeywordRuleModel: keywordRuleModel,
		KeywordFilter:    degradation.NewKeywordFilter(keywordRuleModel, redisClient),
		AdminRpc:         adminClient.NewAdmin(zrpc.MustNewClient(c.AdminRpc)),
		NewsRpc:          news.NewNews(zrpc.MustNewClient(c.NewsRpc)),
		PaperRpc:         paper.NewPaper(zrpc.MustNewClient(c.PaperRpc)),
		UserRpc:          user.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}
}
