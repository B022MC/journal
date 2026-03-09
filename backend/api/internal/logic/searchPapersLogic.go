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

func (l *SearchPapersLogic) SearchPapers(req *types.SearchPapersReq) (resp *types.ListPapersResp, err error) {
	rpcResp, err := l.svcCtx.PaperRpc.SearchPapers(l.ctx, &paper.SearchPapersReq{
		Query:      req.Query,
		Discipline: req.Discipline,
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
