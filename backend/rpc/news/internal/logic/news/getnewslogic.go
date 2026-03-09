package newslogic

import (
	"context"

	"journal/rpc/news/internal/svc"
	"journal/rpc/news/news"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNewsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetNewsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNewsLogic {
	return &GetNewsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetNewsLogic) GetNews(in *news.GetNewsReq) (*news.GetNewsResp, error) {
	n, err := l.svcCtx.NewsModel.FindById(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}

	return &news.GetNewsResp{
		News: &news.NewsItem{
			Id:        n.Id,
			Title:     n.Title,
			TitleEn:   n.TitleEn,
			Content:   n.Content,
			ContentEn: n.GetContentEn(),
			AuthorId:  n.AuthorId,
			Category:  n.Category,
			IsPinned:  n.GetIsPinned(),
			CreatedAt: n.CreatedAt.Unix(),
		},
	}, nil
}
