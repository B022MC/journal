package logic

import (
	"context"

	"journal/rpc/admin/admin"
	"journal/rpc/admin/internal/svc"

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

// Paper
func (l *ListPapersLogic) ListPapers(in *admin.ListPapersReq) (*admin.ListPapersResp, error) {
	papers, total, err := l.svcCtx.PaperModel.ListPapersPaginatedAdmin(l.ctx, int(in.Page), int(in.PageSize), in.Zone, in.Status)
	if err != nil {
		return nil, err
	}

	var items []*admin.PaperItem
	for _, p := range papers {
		items = append(items, &admin.PaperItem{
			Id:               p.Id,
			Title:            p.Title,
			AuthorId:         p.AuthorId,
			AuthorName:       p.AuthorName,
			Zone:             p.Zone,
			Status:           p.Status,
			DegradationLevel: p.DegradationLevel,
			CreatedAt:        p.CreatedAt.Unix(),
			ShitScore:        p.ShitScore,
		})
	}

	return &admin.ListPapersResp{
		Items: items,
		Total: total,
	}, nil
}
