package logic

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/news/news"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNewsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNewsLogic {
	return &GetNewsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNewsLogic) GetNews(req *types.IdReq) (resp *types.NewsItem, err error) {
	rpcResp, err := l.svcCtx.NewsRpc.GetNews(l.ctx, &news.GetNewsReq{Id: req.Id})
	if err != nil {
		return nil, err
	}
	n := rpcResp.News
	return &types.NewsItem{
		Id:        n.Id,
		Title:     n.Title,
		TitleEn:   n.TitleEn,
		Content:   n.Content,
		ContentEn: n.ContentEn,
		AuthorId:  n.AuthorId,
		Category:  n.Category,
		IsPinned:  n.IsPinned,
		CreatedAt: n.CreatedAt,
	}, nil
}
