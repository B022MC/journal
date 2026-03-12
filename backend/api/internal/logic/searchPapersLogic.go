package logic

import (
	"context"

	"journal/api/internal/svc"
	"journal/api/internal/types"
	"journal/rpc/paper/paper"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchPapersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchPapersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchPapersLogic {
	return &SearchPapersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchPapersLogic) SearchPapers(req *types.SearchPapersReq) (resp *types.SearchPapersResp, err error) {
	rpcResp, err := l.svcCtx.PaperRpc.SearchPapers(l.ctx, &paper.SearchPapersReq{
		Query:           req.Query,
		Discipline:      req.Discipline,
		Sort:            req.Sort,
		Engine:          req.Engine,
		ShadowCompare:   req.ShadowCompare,
		SuggestionLimit: int32(req.SuggestionLimit),
		Page:            int32(req.Page),
		PageSize:        int32(req.PageSize),
	})
	if err != nil {
		return nil, err
	}
	meta := types.SearchMeta{}
	if rpcResp.Meta != nil {
		meta = types.SearchMeta{
			Engine:         rpcResp.Meta.Engine,
			UsedFallback:   rpcResp.Meta.UsedFallback,
			FallbackReason: rpcResp.Meta.FallbackReason,
			ShadowCompared: rpcResp.Meta.ShadowCompared,
			IndexedDocs:    rpcResp.Meta.IndexedDocs,
			IndexedTerms:   rpcResp.Meta.IndexedTerms,
			IndexSignature: rpcResp.Meta.IndexSignature,
			ExpandedTerms:  rpcResp.Meta.ExpandedTerms,
		}
	}
	return &types.SearchPapersResp{
		Items:       toPaperItems(rpcResp.Items),
		Total:       rpcResp.Total,
		Suggestions: rpcResp.Suggestions,
		Meta:        meta,
	}, nil
}
