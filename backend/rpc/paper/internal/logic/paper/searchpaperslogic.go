package paperlogic

import (
	"context"

	"journal/rpc/paper/internal/search"
	"journal/rpc/paper/internal/svc"
	"journal/rpc/paper/paper"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchPapersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchPapersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchPapersLogic {
	return &SearchPapersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SearchPapersLogic) SearchPapers(in *paper.SearchPapersReq) (*paper.ListPapersResp, error) {
	page := int(in.Page)
	pageSize := int(in.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	searchResp, err := l.svcCtx.SearchService.Search(l.ctx, search.Request{
		Query:      in.Query,
		Discipline: in.Discipline,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]*paper.PaperItem, 0, len(searchResp.Papers))
	for _, p := range searchResp.Papers {
		item := toPaperItem(p)
		item.Content = ""
		items = append(items, item)
	}

	return &paper.ListPapersResp{
		Items: items,
		Total: searchResp.Total,
	}, nil
}
