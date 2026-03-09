package logic

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/news/news"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListNewsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListNewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListNewsLogic {
	return &ListNewsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListNewsLogic) ListNews(req *types.PageReq) (resp *types.ListNewsResp, err error) {
	rpcResp, err := l.svcCtx.NewsRpc.ListNews(l.ctx, &news.ListNewsReq{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}

	items := make([]types.NewsItem, 0, len(rpcResp.Items))
	for _, n := range rpcResp.Items {
		items = append(items, types.NewsItem{
			Id:        n.Id,
			Title:     n.Title,
			TitleEn:   n.TitleEn,
			Content:   n.Content,
			ContentEn: n.ContentEn,
			AuthorId:  n.AuthorId,
			Category:  n.Category,
			IsPinned:  n.IsPinned,
			CreatedAt: n.CreatedAt,
		})
	}

	return &types.ListNewsResp{Items: items, Total: rpcResp.Total}, nil
}
