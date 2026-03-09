package paperlogic

import (
	"context"

	"journal/model"
	"journal/rpc/paper/internal/svc"
	"journal/rpc/paper/paper"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPaperLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPaperLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaperLogic {
	return &GetPaperLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPaperLogic) GetPaper(in *paper.GetPaperReq) (*paper.GetPaperResp, error) {
	// Increment view count asynchronously
	go func() {
		_ = l.svcCtx.PaperModel.IncrViewCount(context.Background(), in.Id)
	}()

	p, err := l.svcCtx.PaperModel.FindById(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}

	return &paper.GetPaperResp{
		Paper: toPaperItem(p),
	}, nil
}

func toPaperItem(p *model.Paper) *paper.PaperItem {
	item := &paper.PaperItem{
		Id:               p.Id,
		Title:            p.Title,
		TitleEn:          p.TitleEn,
		Abstract:         p.Abstract,
		AbstractEn:       p.GetAbstractEn(),
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
		CreatedAt:        p.CreatedAt.Unix(),
	}
	if pt := p.GetPromotedAt(); pt != nil {
		item.PromotedAt = pt.Unix()
	}
	return item
}
