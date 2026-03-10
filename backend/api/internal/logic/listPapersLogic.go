package logic

import (
	"context"

	apicache "journal/api/internal/cache"
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
	page, pageSize := normalizePagination(req.Page, req.PageSize)

	if shouldUseHotPapersCache(req, page, pageSize) {
		if cached, ok, cacheErr := l.svcCtx.Cache.GetHotPapers(l.ctx, req.Zone); cacheErr != nil {
			l.Errorf("load hot papers cache failed for zone=%q: %v", req.Zone, cacheErr)
		} else if ok {
			return sliceHotPapers(cached, page, pageSize), nil
		}

		hotResp, hotErr := l.fetchHotPapers(req.Zone)
		if hotErr != nil {
			return nil, hotErr
		}
		if cacheErr := l.svcCtx.Cache.SetHotPapers(l.ctx, req.Zone, hotResp); cacheErr != nil {
			l.Errorf("store hot papers cache failed for zone=%q: %v", req.Zone, cacheErr)
		}
		return sliceHotPapers(hotResp, page, pageSize), nil
	}

	rpcResp, err := l.svcCtx.PaperRpc.ListPapers(l.ctx, &paper.ListPapersReq{
		Zone:       req.Zone,
		Discipline: req.Discipline,
		Sort:       req.Sort,
		Page:       int32(page),
		PageSize:   int32(pageSize),
	})
	if err != nil {
		return nil, err
	}
	return &types.ListPapersResp{
		Items: toPaperItems(rpcResp.Items),
		Total: rpcResp.Total,
	}, nil
}

func (l *ListPapersLogic) fetchHotPapers(zone string) (*types.ListPapersResp, error) {
	rpcResp, err := l.svcCtx.PaperRpc.ListPapers(l.ctx, &paper.ListPapersReq{
		Zone:       zone,
		Discipline: "",
		Sort:       "highest_rated",
		Page:       1,
		PageSize:   apicache.HotPapersLimit,
	})
	if err != nil {
		return nil, err
	}

	return &types.ListPapersResp{
		Items: toPaperItems(rpcResp.Items),
		Total: rpcResp.Total,
	}, nil
}

func normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > apicache.HotPapersLimit {
		pageSize = 20
	}
	return page, pageSize
}

func shouldUseHotPapersCache(req *types.ListPapersReq, page, pageSize int) bool {
	if req == nil || req.Sort != "highest_rated" || req.Discipline != "" {
		return false
	}

	offset := (page - 1) * pageSize
	return offset >= 0 && offset+pageSize <= apicache.HotPapersLimit
}

func sliceHotPapers(payload *types.ListPapersResp, page, pageSize int) *types.ListPapersResp {
	if payload == nil {
		return &types.ListPapersResp{}
	}

	offset := (page - 1) * pageSize
	if offset < 0 || offset >= len(payload.Items) {
		return &types.ListPapersResp{
			Items: []types.PaperItem{},
			Total: payload.Total,
		}
	}

	end := offset + pageSize
	if end > len(payload.Items) {
		end = len(payload.Items)
	}

	items := append([]types.PaperItem(nil), payload.Items[offset:end]...)
	return &types.ListPapersResp{
		Items: items,
		Total: payload.Total,
	}
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
