package paperlogic

import (
	"context"

	"journal/rpc/paper/internal/svc"
	"journal/rpc/paper/paper"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserPapersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUserPapersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserPapersLogic {
	return &UserPapersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UserPapersLogic) UserPapers(in *paper.UserPapersReq) (*paper.ListPapersResp, error) {
	page := int(in.Page)
	pageSize := int(in.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	papers, total, err := l.svcCtx.PaperModel.ListByAuthor(l.ctx, in.AuthorId, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]*paper.PaperItem, 0, len(papers))
	for _, p := range papers {
		item := toPaperItem(p)
		item.Content = ""
		items = append(items, item)
	}

	return &paper.ListPapersResp{
		Items: items,
		Total: total,
	}, nil
}
