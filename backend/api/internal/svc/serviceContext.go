// Code scaffolded by goctl. Safe to edit.

package svc

import (
	"journal/api/internal/config"
	"journal/rpc/news/client/news"
	"journal/rpc/paper/client/paper"
	"journal/rpc/rating/client/rating"
	"journal/rpc/user/client/user"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	UserRpc   user.User
	PaperRpc  paper.Paper
	RatingRpc rating.Rating
	NewsRpc   news.News
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		UserRpc:   user.NewUser(zrpc.MustNewClient(c.UserRpc)),
		PaperRpc:  paper.NewPaper(zrpc.MustNewClient(c.PaperRpc)),
		RatingRpc: rating.NewRating(zrpc.MustNewClient(c.RatingRpc)),
		NewsRpc:   news.NewNews(zrpc.MustNewClient(c.NewsRpc)),
	}
}
