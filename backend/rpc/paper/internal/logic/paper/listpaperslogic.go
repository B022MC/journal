package paperlogic

import (
	"context"

	"journal/rpc/paper/internal/svc"
	"journal/rpc/paper/paper"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPapersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPapersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPapersLogic {
	return &ListPapersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListPapersLogic) ListPapers(in *paper.ListPapersReq) (*paper.ListPapersResp, error) {
	page := int(in.Page)
	pageSize := int(in.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	papers, total, err := l.svcCtx.PaperModel.List(l.ctx, in.Zone, in.Discipline, in.Sort, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]*paper.PaperItem, 0, len(papers))
	for _, p := range papers {
		item := toPaperItem(p)
		item.Content = "" // Don't return full content in list
		items = append(items, item)
	}

	return &paper.ListPapersResp{
		Items: items,
		Total: total,
	}, nil
}
