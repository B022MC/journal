package logic

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/paper/paper"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPapersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPapersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPapersLogic {
	return &ListPapersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPapersLogic) ListPapers(req *types.ListPapersReq) (resp *types.ListPapersResp, err error) {
	rpcResp, err := l.svcCtx.PaperRpc.ListPapers(l.ctx, &paper.ListPapersReq{
		Zone:       req.Zone,
		Discipline: req.Discipline,
		Sort:       req.Sort,
		Page:       int32(req.Page),
		PageSize:   int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}
	return &types.ListPapersResp{
		Items: toPaperItems(rpcResp.Items),
		Total: rpcResp.Total,
	}, nil
}

func toPaperItems(items []*paper.PaperItem) []types.PaperItem {
	result := make([]types.PaperItem, 0, len(items))
	for _, p := range items {
		result = append(result, toPaperType(p))
	}
	return result
}

func toPaperType(p *paper.PaperItem) types.PaperItem {
	return types.PaperItem{
		Id:               p.Id,
		Title:            p.Title,
		TitleEn:          p.TitleEn,
		Abstract:         p.Abstract,
		AbstractEn:       p.AbstractEn,
		Content:          p.Content,
		AuthorId:         p.AuthorId,
		AuthorName:       p.AuthorName,
		Discipline:       p.Discipline,
		Zone:             p.Zone,
		ShitScore:        p.ShitScore,
		AvgRating:        p.AvgRating,
		RatingCount:      p.RatingCount,
		ViewCount:        p.ViewCount,
		ControversyIndex: p.ControversyIndex,
		Doi:              p.Doi,
		Keywords:         p.Keywords,
		FilePath:         p.FilePath,
		CreatedAt:        p.CreatedAt,
		PromotedAt:       p.PromotedAt,
	}
}
