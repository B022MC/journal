package newslogic

import (
	"context"

	"journal/rpc/news/internal/svc"
	"journal/rpc/news/news"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListNewsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListNewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListNewsLogic {
	return &ListNewsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListNewsLogic) ListNews(in *news.ListNewsReq) (*news.ListNewsResp, error) {
	page := int(in.Page)
	pageSize := int(in.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := l.svcCtx.NewsModel.List(l.ctx, in.Category, page, pageSize)
	if err != nil {
		return nil, err
	}

	newsItems := make([]*news.NewsItem, 0, len(items))
	for _, n := range items {
		newsItems = append(newsItems, &news.NewsItem{
			Id:        n.Id,
			Title:     n.Title,
			TitleEn:   n.TitleEn,
			Content:   n.Content,
			ContentEn: n.GetContentEn(),
			AuthorId:  n.AuthorId,
			Category:  n.Category,
			IsPinned:  n.GetIsPinned(),
			CreatedAt: n.CreatedAt.Unix(),
		})
	}

	return &news.ListNewsResp{
		Items: newsItems,
		Total: total,
	}, nil
}
